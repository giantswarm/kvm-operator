package nodecontroller

import (
	"context"
	"reflect"

	"github.com/giantswarm/apiextensions/v3/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/k8sclient/v5/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"k8s.io/client-go/rest"

	"github.com/giantswarm/kvm-operator/service/controller/internal/nodecontroller"
	"github.com/giantswarm/kvm-operator/service/controller/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.Debugf(ctx, "ensuring node controller is running for workload cluster")

	restConfig, err := r.workloadCluster.NewRestConfig(ctx, key.ClusterID(cr), key.ClusterAPIEndpoint(cr))
	if err != nil {
		return microerror.Mask(err)
	}

	var needCreate bool
	controllerKey := controllerMapKey(cr)
	r.controllerMutex.Lock() // Atomically read and check the existing controller
	if controller, ok := r.controllers[controllerKey]; ok {
		if !restConfigsEqual(controller.restConfig, restConfig) {
			r.logger.Debugf(ctx, "rest configs changed")
			needCreate = true
		} else if !clusterSpecsEqual(controller.cluster.Spec, cr.Spec) {
			r.logger.Debugf(ctx, "cluster spec changed")
			needCreate = true
		}

		if needCreate {
			controller.Stop(ctx)
			delete(r.controllers, controllerKey)
		}
	} else {
		r.logger.Debugf(ctx, "controller not found")
		needCreate = true
	}
	r.controllerMutex.Unlock()

	if !needCreate {
		r.logger.Debugf(ctx, "ensured controller is running")
		return nil
	}

	r.logger.Debugf(ctx, "controller needs to be created")

	k8sClient, err := k8sclient.NewClients(k8sclient.ClientsConfig{
		Logger:     r.logger,
		RestConfig: restConfig,
	})
	if err != nil {
		return microerror.Mask(err)
	}

	controller, err := nodecontroller.New(nodecontroller.Config{
		Cluster:             cr,
		ManagementK8sClient: r.k8sClient,
		WorkloadK8sClient:   k8sClient,
		Logger:              r.logger,
	})
	if err != nil {
		return microerror.Mask(err)
	}

	go func() {
		r.controllerMutex.Lock() // Atomically register the new controller
		r.controllers[controllerKey] = controllerWithConfig{
			cluster:    cr,
			restConfig: restConfig,
			Controller: controller,
		}
		controller.Boot(ctx)
		r.controllerMutex.Unlock()

		<-controller.Booted()

		r.logger.Debugf(ctx, "controller booted and registered")
	}()

	return nil
}

func restConfigsEqual(left *rest.Config, right *rest.Config) bool {
	if left.Host != right.Host {
		return false
	}
	return reflect.DeepEqual(left.TLSClientConfig, right.TLSClientConfig)
}

func clusterSpecsEqual(left v1alpha1.KVMConfigSpec, right v1alpha1.KVMConfigSpec) bool {
	return reflect.DeepEqual(left, right)
}
