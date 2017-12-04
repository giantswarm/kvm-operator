package ingressv1

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"

	"github.com/giantswarm/kvm-operator/service/keyv1"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := keyv1.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "debug", "looking for ingresses in the Kubernetes API")

	var ingresses []*v1beta1.Ingress

	namespace := keyv1.ClusterNamespace(customObject)
	ingressNames := []string{
		APIID,
		EtcdID,
	}

	for _, name := range ingressNames {
		manifest, err := r.k8sClient.Extensions().Ingresses(namespace).Get(name, apismetav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			r.logger.LogCtx(ctx, "debug", "did not find a ingress in the Kubernetes API")
			// fall through
		} else if err != nil {
			return nil, microerror.Mask(err)
		} else {
			r.logger.LogCtx(ctx, "debug", "found a ingress in the Kubernetes API")
			ingresses = append(ingresses, manifest)
		}
	}

	r.logger.LogCtx(ctx, "debug", fmt.Sprintf("found %d ingresses in the Kubernetes API", len(ingresses)))

	return ingresses, nil
}
