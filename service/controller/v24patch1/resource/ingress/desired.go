package ingress

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	"k8s.io/api/extensions/v1beta1"

<<<<<<< HEAD
	"github.com/giantswarm/kvm-operator/service/controller/v24patch1/key"
=======
	"github.com/giantswarm/kvm-operator/service/controller/v24/key"
>>>>>>> c4c6c79d... copy v24 to v24patch1
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "computing the new ingresses")

	var ingresses []*v1beta1.Ingress

	ingresses = append(ingresses, newAPIIngress(customObject))
	ingresses = append(ingresses, newEtcdIngress(customObject))

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("computed the %d new ingresses", len(ingresses)))

	return ingresses, nil
}
