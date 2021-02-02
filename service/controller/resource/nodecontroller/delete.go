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

	if controller, ok := r.controllers[nodeControllerKey(cr)]; ok {
		controller.Shutdown()

		r.controllerMutex.Lock()
		delete(r.controllers, nodeControllerKey(cr))
		r.controllerMutex.Unlock()
	}

	return nil
}
