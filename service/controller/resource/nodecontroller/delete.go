package nodecontroller

import (
	"context"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/kvm-operator/service/controller/key"
)

func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.Debugf(ctx, "ensuring node controller is shut down")

	if controller, ok := r.controllers[nodeControllerKey(cr)]; ok {
		r.logger.Debugf(ctx, "node controller found, shutting down")

		controller.Shutdown()

		r.controllerMutex.Lock()
		delete(r.controllers, nodeControllerKey(cr))
		r.controllerMutex.Unlock()

		r.logger.Debugf(ctx, "node controller shut down and unregistered")
	}

	r.logger.Debugf(ctx, "ensured node controller is shut down")

	return nil
}
