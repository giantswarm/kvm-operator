package service

import (
	"context"

	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"

	"github.com/giantswarm/kvm-operator/service/controller/key"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	cr, err := key.ToKVMCluster(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.Debugf(ctx, "computing the new services")

	var services []*corev1.Service

	services = append(services, newMasterService(cr))
	services = append(services, newWorkerService(cr))

	r.logger.Debugf(ctx, "computed the %d new services", len(services))

	return services, nil
}
