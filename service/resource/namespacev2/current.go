package namespacev2

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/framework/context/canceledcontext"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiv1 "k8s.io/client-go/pkg/api/v1"

	"github.com/giantswarm/kvm-operator/service/keyv2"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := keyv2.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "debug", "looking for the namespace in the Kubernetes API")

	// Lookup the current state of the namespace.
	var namespace *apiv1.Namespace
	{
		manifest, err := r.k8sClient.CoreV1().Namespaces().Get(keyv2.ClusterNamespace(customObject), apismetav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			r.logger.LogCtx(ctx, "debug", "did not find the namespace in the Kubernetes API")
			// fall through
		} else if err != nil {
			return nil, microerror.Mask(err)
		} else {
			r.logger.LogCtx(ctx, "debug", "found the namespace in the Kubernetes API")
			namespace = manifest
		}
	}

	// In case the namespace is already terminating we do not need to do any
	// further work. Then we cancel the reconciliation to prevent the current and
	// any further resource from being processed.
	if namespace != nil && namespace.Status.Phase == "Terminating" {
		r.logger.LogCtx(ctx, "debug", "namespace is in state 'Terminating'")

		canceledcontext.SetCanceled(ctx)
		if canceledcontext.IsCanceled(ctx) {
			r.logger.LogCtx(ctx, "debug", "canceling reconciliation for custom object")

			return nil, nil
		}
	}

	return namespace, nil
}
