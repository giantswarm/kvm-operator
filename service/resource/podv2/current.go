package podv1

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/framework/context/canceledcontext"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/kvm-operator/service/keyv2"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	reconciledPod, err := keyv2.ToPod(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "debug", "looking for the current version of the reconciled pod in the Kubernetes API")

	currentPod, err := r.k8sClient.CoreV1().Pods(reconciledPod.Namespace).Get(reconciledPod.Name, apismetav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		// In case we reconcile a pod we cannot find anymore this means the
		// informer's watch event is outdated and the pod got already deleted in the
		// Kubernetes API. This is a normal transition behaviour, so we just ignore
		// it and assume we are done.

		r.logger.LogCtx(ctx, "debug", "cannot find the current version of the reconciled pod in the Kubernetes API")

		canceledcontext.SetCanceled(ctx)
		if canceledcontext.IsCanceled(ctx) {
			r.logger.LogCtx(ctx, "debug", "canceling reconciliation for pod")

			return nil, nil
		}

		return nil, nil
	} else if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "debug", "found the current version of the reconciled pod in the Kubernetes API")

	return currentPod, nil
}
