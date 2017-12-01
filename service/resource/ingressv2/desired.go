package ingressv1

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"

	"github.com/giantswarm/kvm-operator/service/keyv1"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := keyv1.ToCustomObject(obj)
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
