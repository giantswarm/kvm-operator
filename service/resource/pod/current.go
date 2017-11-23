package pod

import (
	"context"

	"github.com/giantswarm/microerror"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/kvm-operator/service/key"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	reconciledPod, err := key.ToPod(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "debug", "looking for the current version of the reconciled pod in the Kubernetes API")

	currentPod, err := r.k8sClient.CoreV1().Pods(reconciledPod.Namespace).Get(reconciledPod.Name, apismetav1.GetOptions{})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "debug", "found the current version of the reconciled pod in the Kubernetes API")

	return currentPod, nil
}
