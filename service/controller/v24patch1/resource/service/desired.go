package service

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	apiv1 "k8s.io/api/core/v1"

<<<<<<< HEAD
<<<<<<< HEAD
	"github.com/giantswarm/kvm-operator/service/controller/v24patch1/key"
=======
	"github.com/giantswarm/kvm-operator/service/controller/v24/key"
>>>>>>> c4c6c79d... copy v24 to v24patch1
=======
	"github.com/giantswarm/kvm-operator/service/controller/v24patch1/key"
>>>>>>> d6f149c2... wire v24patch1
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "computing the new services")

	var services []*apiv1.Service

	services = append(services, newMasterService(customObject))
	services = append(services, newWorkerService(customObject))

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("computed the %d new services", len(services)))

	return services, nil
}
