package nodecontroller

import (
	"context"

	"github.com/giantswarm/apiextensions/v3/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/errors/tenant"
	"github.com/giantswarm/k8sclient/v5/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/tenantcluster/v4/pkg/tenantcluster"

	"github.com/giantswarm/kvm-operator/service/controller/internal/nodecontroller"
	"github.com/giantswarm/kvm-operator/service/controller/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.Debugf(ctx, "ensuring controller is running for workload cluster")

	desired, err := r.calculateDesiredController(ctx, cr)
	if tenantcluster.IsTimeout(err) {
		r.logger.Debugf(ctx, "waiting for certificates timed out")
		return nil
	} else if tenant.IsAPINotAvailable(err) {
		r.logger.Debugf(ctx, "tenant cluster is not available")
		return nil
	} else if err != nil {
		return microerror.Mask(err)
	}

	if needCreate := r.applyUpdateChange(ctx, desired); needCreate {
		r.logger.Debugf(ctx, "controller needs to be created")
		go r.applyCreateChangeAsync(ctx, desired)
	} else {
		r.logger.Debugf(ctx, "ensured controller is running")
	}

	return nil
}

// applyCreateChangeAsync boots the desired controller and saves it in the controllers map
// if it boots successfully. It should be run as a goroutine to avoid blocking reconciliation of
// other resources.
func (r *Resource) applyCreateChangeAsync(ctx context.Context, desired controllerWithConfig) {
	go desired.Boot(ctx)
	<-desired.Booted() // Wait for the controller to be booted

	r.controllerMutex.Lock()
	controllerKey := controllerMapKey(desired.cluster)
	if _, ok := r.controllers[controllerKey]; ok {
		// There's a small chance for a race if two goroutines are booting controllers for the same clusters
		// because, for example, several changes were made to a cluster CR in a short amount of time and booting
		// of controllers takes a non-zero amount of time. By stopping the desired controller within the mutex,
		// we ensure that only one controller is running at a time and we don't have dangling controllers causing
		// a memory leak.
		desired.Stop(ctx)
	} else {
		r.controllers[controllerKey] = desired
	}
	r.controllerMutex.Unlock()

	r.logger.Debugf(ctx, "controller booted and registered")
}

// applyUpdateChange checks for an existing controller, stops it if it doesn't match the desired controller,
// and returns a value indicating if the desired controller should be started (because the existing one was
// stopped or there was no existing one).
func (r *Resource) applyUpdateChange(ctx context.Context, desired controllerWithConfig) bool {
	r.controllerMutex.Lock()         // Exclude access from other goroutines
	defer r.controllerMutex.Unlock() // Ensure this is run for all function returns

	controllerKey := controllerMapKey(desired.cluster)
	current, ok := r.controllers[controllerKey]

	if !ok {
		r.logger.Debugf(ctx, "existing controller not found")
		return true
	}

	if !current.equal(desired) {
		r.logger.Debugf(ctx, "existing controller needs to be replaced, stopping and deleting")
		current.Stop(ctx) // This is just closing a channel, should be almost instant
		delete(r.controllers, controllerKey)
		return true
	}

	r.logger.Debugf(ctx, "existing controller doesn't need to be replaced")
	return false
}

// calculateDesiredController returns a controllerWithConfig holding a controller which watches the workload cluster
// and other information used to determine when the controller needs to be created or updated.
func (r *Resource) calculateDesiredController(ctx context.Context, cluster v1alpha1.KVMConfig) (controllerWithConfig, error) {
	restConfig, err := r.workloadCluster.NewRestConfig(ctx, key.ClusterID(cluster), key.ClusterAPIEndpoint(cluster))
	if err != nil {
		return controllerWithConfig{}, microerror.Mask(err)
	}

	k8sClient, err := k8sclient.NewClients(k8sclient.ClientsConfig{
		Logger:     r.logger,
		RestConfig: restConfig,
	})
	if err != nil {
		return controllerWithConfig{}, microerror.Mask(err)
	}

	nodeController, err := nodecontroller.New(nodecontroller.Config{
		Cluster:             cluster,
		ManagementK8sClient: r.k8sClient,
		WorkloadK8sClient:   k8sClient,
		Logger:              r.logger,
	})
	if err != nil {
		return controllerWithConfig{}, microerror.Mask(err)
	}

	withConfig := controllerWithConfig{
		cluster:    cluster,
		restConfig: restConfig,
		Controller: nodeController,
	}
	return withConfig, nil
}
