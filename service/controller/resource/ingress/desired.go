package ingress

import (
	"context"

	"github.com/giantswarm/microerror"
	"k8s.io/api/networking/v1beta1"

	"github.com/giantswarm/kvm-operator/v4/service/controller/key"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.Debugf(ctx, "computing the new ingresses")

	var ingresses []*v1beta1.Ingress

	ingresses = append(ingresses, newAPIIngress(customObject))
	ingresses = append(ingresses, newEtcdIngress(customObject))

	r.logger.Debugf(ctx, "computed the %d new ingresses", len(ingresses))

	return ingresses, nil
}
