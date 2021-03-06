package nodecontroller

import (
	"context"
	"fmt"
	"time"

	workloaderrors "github.com/giantswarm/errors/tenant"
	"github.com/giantswarm/k8sclient/v5/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	workloadcluster "github.com/giantswarm/tenantcluster/v4/pkg/tenantcluster"
	"k8s.io/apimachinery/pkg/labels"

	"github.com/giantswarm/kvm-operator/v4/pkg/label"
	"github.com/giantswarm/kvm-operator/v4/pkg/project"
	"github.com/giantswarm/kvm-operator/v4/service/controller/internal/nodecontroller"
	"github.com/giantswarm/kvm-operator/v4/service/controller/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.Debugf(ctx, "ensuring controller is running for workload cluster")

	clusterID := key.ClusterID(cr)
	controllerKey := controllerMapKey(cr)

	var shouldStop bool
	var k8sClient *k8sclient.Clients
	{
		k8sClient, err = key.CreateK8sClientForWorkloadCluster(ctx, cr, r.logger, r.workloadCluster)
		if workloadcluster.IsTimeout(err) {
			r.logger.Debugf(ctx, "waiting for certificates timed out")
			shouldStop = true
		} else if workloaderrors.IsAPINotAvailable(err) || k8sclient.IsTimeout(err) {
			r.logger.Debugf(ctx, "workload cluster is not available")
			shouldStop = true
		} else if err != nil {
			return microerror.Mask(err)
		}
	}

	var desiredController *nodecontroller.Controller
	if k8sClient != nil {
		config := nodecontroller.Config{
			Cluster:             cr,
			ManagementK8sClient: r.k8sClient,
			WorkloadK8sClient:   k8sClient,
			Logger:              r.logger,
			Name:                fmt.Sprintf("%s-%s-nodes", project.Name(), clusterID),
			Selector: labels.SelectorFromSet(map[string]string{
				// When managing endpoints, we only consider node- and pod-readiness for workers. Master endpoint IPs
				// are always present in the master endpoints object. This may change when we implement HA for KVM.
				"role":                key.WorkerID,
				label.OperatorVersion: project.Version(),
			}),
		}
		desiredController, err = nodecontroller.New(config)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	// Defer all changes to the controller map until the end of this function so we can wrap in a mutex
	r.controllerMutex.Lock()
	defer r.controllerMutex.Unlock()

	if current, ok := r.controllers[controllerKey]; ok {
		if !current.Equal(desiredController) {
			r.logger.Debugf(ctx, "controllers don't match")
			shouldStop = true
		} else if time.Since(current.LastReconciled()) > 2*nodecontroller.ResyncPeriod {
			r.logger.Debugf(ctx, "controller hasn't reconciled in more than %d minutes", int(2*nodecontroller.ResyncPeriod.Minutes()))
			shouldStop = true
		}

		if shouldStop {
			r.logger.Debugf(ctx, "stopping existing controller")
			current.Stop()
			delete(r.controllers, controllerKey)
		} else {
			// Current controller matches expected and is reconciling, do nothing
			return nil
		}
	}

	if desiredController != nil {
		r.logger.Debugf(ctx, "booting new controller")
		err = desiredController.Boot()
		if err != nil {
			return microerror.Mask(err)
		}

		r.controllers[controllerKey] = desiredController
		r.logger.Debugf(ctx, "controller booted and registered")
	}

	return nil
}
