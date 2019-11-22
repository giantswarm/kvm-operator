package configmap

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/resourcecanceledcontext"
	corev1 "k8s.io/api/core/v1"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/kvm-operator/service/controller/v26/key"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	customResource, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	if key.IsDeleted(customResource) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "redirecting responsibility of deletion of config maps to namespace termination")
		resourcecanceledcontext.SetCanceled(ctx)
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

		return nil, nil
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "looking for a list of config maps in the Kubernetes API")

	var currentConfigMaps []*corev1.ConfigMap
	{
		namespace := key.ClusterNamespace(customResource)
		configMapList, err := r.k8sClient.CoreV1().ConfigMaps(namespace).List(apismetav1.ListOptions{})
		if err != nil {
			return nil, microerror.Mask(err)
		} else {
			r.logger.LogCtx(ctx, "level", "debug", "message", "found a list of config maps in the Kubernetes API")

			for _, item := range configMapList.Items {
				c := item
				currentConfigMaps = append(currentConfigMaps, &c)
			}
		}
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found a list of %d config maps in the Kubernetes API", len(currentConfigMaps)))

	return currentConfigMaps, nil
}
