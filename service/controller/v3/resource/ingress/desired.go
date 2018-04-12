package ingress

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	"k8s.io/api/extensions/v1beta1"

	"github.com/giantswarm/kvm-operator/service/controller/v3/key"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "debug", "computing the new ingresses")

	var ingresses []*v1beta1.Ingress

	ingresses = append(ingresses, newAPIIngress(customObject))
	ingresses = append(ingresses, newEtcdIngress(customObject))

	r.logger.LogCtx(ctx, "debug", fmt.Sprintf("computed the %d new ingresses", len(ingresses)))

	return ingresses, nil
}
