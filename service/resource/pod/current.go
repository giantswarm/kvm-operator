package pod

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/framework/context/canceledcontext"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiv1 "k8s.io/client-go/pkg/api/v1"

	"github.com/giantswarm/kvm-operator/service/key"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	reconciledPod, err := key.ToPod(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "debug", "looking for the reconciled pod's namespace in the Kubernetes API")

	var namespace *apiv1.Namespace
	{
		namespace, err = r.k8sClient.CoreV1().Namespaces().Get(reconciledPod.Namespace, apismetav1.GetOptions{})
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	r.logger.LogCtx(ctx, "debug", "found the reconciled pod's namespace in the Kubernetes API")

	// In case the namespace is already terminating we do not need to do any
	// further work. Then we cancel the reconciliation to prevent the current and
	// any further resource from being processed.
	if namespace != nil && namespace.Status.Phase == "Terminating" {
		r.logger.LogCtx(ctx, "debug", "namespace of the reconciled pod is in state 'Terminating'")

		canceledcontext.SetCanceled(ctx)
		if canceledcontext.IsCanceled(ctx) {
			r.logger.LogCtx(ctx, "debug", "canceling further pod reconciliation")

			return nil, nil
		}
	}

	return nil, nil
}
