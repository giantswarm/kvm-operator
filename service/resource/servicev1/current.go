package servicev1

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiv1 "k8s.io/client-go/pkg/api/v1"

	"github.com/giantswarm/kvm-operator/service/keyv1"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := keyv1.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "debug", "looking for services in the Kubernetes API")

	var services []*apiv1.Service

	namespace := keyv1.ClusterNamespace(customObject)
	serviceNames := []string{
		keyv1.MasterID,
		keyv1.WorkerID,
	}

	for _, name := range serviceNames {
		manifest, err := r.k8sClient.CoreV1().Services(namespace).Get(name, apismetav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			r.logger.LogCtx(ctx, "debug", "did not find a service in the Kubernetes API")
			// fall through
		} else if err != nil {
			return nil, microerror.Mask(err)
		} else {
			r.logger.LogCtx(ctx, "debug", "found a service in the Kubernetes API")
			services = append(services, manifest)
		}
	}

	r.logger.LogCtx(ctx, "debug", fmt.Sprintf("found %d services in the Kubernetes API", len(services)))

	return services, nil
}
