package nodecontroller

import (
	"context"

	"github.com/giantswarm/apiextensions/v3/pkg/apis/provider/v1alpha1"
	workload "github.com/giantswarm/errors/tenant"
	"github.com/giantswarm/k8sclient/v5/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	workloadcluster "github.com/giantswarm/tenantcluster/v4/pkg/tenantcluster"

	"github.com/giantswarm/kvm-operator/service/controller/internal/nodecontroller"
	"github.com/giantswarm/kvm-operator/service/controller/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.Debugf(ctx, "ensuring node controller is running for workload cluster")

	if _, ok := r.controllers[nodeControllerKey(cr)]; !ok {
		r.logger.Debugf(ctx, "node controller not found, creating")

		controller, err := r.newNodeController(ctx, cr)
		if err != nil {
			r.logger.Debugf(ctx, "error during node controller creation")
			return microerror.Mask(err)
		}

		go controller.Boot(context.Background())
		<-controller.Booted()

		r.controllerMutex.Lock()
		r.controllers[nodeControllerKey(cr)] = controller
		r.controllerMutex.Unlock()

		r.logger.Debugf(ctx, "node controller booted and registered")
	}

	r.logger.Debugf(ctx, "ensured node controller is running for workload cluster")

	return nil
}

func (r *Resource) newNodeController(ctx context.Context, cluster v1alpha1.KVMConfig) (*nodecontroller.Controller, error) {
	r.logger.Debugf(ctx, "creating Kubernetes client for workload cluster")

	var workloadK8sClient k8sclient.Interface
	{
		var err error
		workloadK8sClient, err = key.CreateK8sClientForWorkloadCluster(ctx, cluster, r.logger, r.workloadCluster)
		if workloadcluster.IsTimeout(err) {
			r.logger.Debugf(ctx, "did not create Kubernetes client for workload cluster")
			r.logger.Debugf(ctx, "waiting for certificates timed out")
			r.logger.Debugf(ctx, "canceling resource")

			return nil, nil
		} else if workload.IsAPINotAvailable(err) {
			r.logger.Debugf(ctx, "workload cluster is not available")
			r.logger.Debugf(ctx, "canceling resource")

			return nil, nil
		} else if err != nil {
			return nil, microerror.Mask(err)
		}

		r.logger.Debugf(ctx, "created Kubernetes client for workload cluster")
	}

	controller, err := nodecontroller.New(nodecontroller.Config{
		Cluster:             cluster,
		ManagementK8sClient: r.k8sClient,
		WorkloadK8sClient:   workloadK8sClient,
		Logger:              r.logger,
	})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return controller, nil
}
