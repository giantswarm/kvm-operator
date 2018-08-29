package service

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	apiv1 "k8s.io/api/core/v1"

	"github.com/giantswarm/kvm-operator/service/controller/v14patch1/key"
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
