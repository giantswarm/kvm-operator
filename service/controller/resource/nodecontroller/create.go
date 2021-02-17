package nodecontroller

import (
	"context"
	"fmt"
	"time"

	"github.com/giantswarm/errors/tenant"
	"github.com/giantswarm/k8sclient/v5/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/tenantcluster/v4/pkg/tenantcluster"
	"k8s.io/apimachinery/pkg/labels"

	"github.com/giantswarm/kvm-operator/pkg/label"
	"github.com/giantswarm/kvm-operator/pkg/project"
	"github.com/giantswarm/kvm-operator/service/controller/internal/nodecontroller"
	"github.com/giantswarm/kvm-operator/service/controller/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.Debugf(ctx, "ensuring controller is running for workload cluster")

	clusterID := key.ClusterID(cr)
	controllerKey := controllerMapKey(cr)

	var desiredController *nodecontroller.Controller
	{
		k8sClient, err := key.CreateK8sClientForWorkloadCluster(ctx, cr, r.logger, r.workloadCluster)
		var shouldStop bool
		if tenantcluster.IsTimeout(err) {
			r.logger.Debugf(ctx, "waiting for certificates timed out")
			shouldStop = true
		} else if tenant.IsAPINotAvailable(err) || k8sclient.IsTimeout(err) {
			r.logger.Debugf(ctx, "workload cluster is not available")
			shouldStop = true
		} else if err != nil {
			return microerror.Mask(err)
		}

		if shouldStop {
			r.controllerMutex.Lock()
			if current, ok := r.controllers[controllerKey]; ok {
				r.logger.Debugf(ctx, "stopping controller")
				current.Stop()
				delete(r.controllers, controllerKey)
			}
			r.controllerMutex.Unlock()
			// Return early and wait for the next loop as there's no reason to watch an inaccessible Kubernetes API.
			return nil
		}

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

	r.controllerMutex.Lock()
	defer r.controllerMutex.Unlock()

	if current, ok := r.controllers[controllerKey]; ok {
		var shouldStop bool
		if !current.Equal(desiredController) {
			r.logger.Debugf(ctx, "controllers don't match")
			shouldStop = true
		} else if time.Since(current.LastReconciled()) > 2*nodecontroller.ResyncPeriod {
			r.logger.Debugf(ctx, "controller hasn't reconciled in more than %d minutes", nodecontroller.ResyncPeriod.Minutes())
			shouldStop = true
		}

		if shouldStop {
			r.logger.Debugf(ctx, "stopping controller")
			current.Stop()
		}
	}

	r.logger.Debugf(ctx, "booting new controller")
	err = desiredController.Boot()
	if err != nil {
		return microerror.Mask(err)
	}

	r.controllers[controllerKey] = desiredController
	r.logger.Debugf(ctx, "controller booted and registered")

	return nil
}
