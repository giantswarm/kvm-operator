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

	controllerKey := controllerMapKey(cr)
	r.controllerMutex.Lock()
	if controller, ok := r.controllers[controllerKey]; ok {
		controller.Stop(ctx)
		delete(r.controllers, controllerKey)
		r.logger.Debugf(ctx, "controller stopped and deleted")
	} else {
		r.logger.Debugf(ctx, "controller not found")
	}
	r.controllerMutex.Unlock()

	r.logger.Debugf(ctx, "ensured node controller is shut down")

	return nil
}
