package configmapv4

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	apiv1 "k8s.io/api/core/v1"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/kvm-operator/service/keyv2"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := keyv2.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "debug", "looking for a list of config maps in the Kubernetes API")

	var currentConfigMaps []*apiv1.ConfigMap
	{
		namespace := keyv2.ClusterNamespace(customObject)
		configMapList, err := r.k8sClient.CoreV1().ConfigMaps(namespace).List(apismetav1.ListOptions{})
		if err != nil {
			return nil, microerror.Mask(err)
		} else {
			r.logger.LogCtx(ctx, "debug", "found a list of config maps in the Kubernetes API")

			for _, item := range configMapList.Items {
				c := item
				currentConfigMaps = append(currentConfigMaps, &c)
			}
		}
	}

	r.logger.LogCtx(ctx, "debug", fmt.Sprintf("found a list of %d config maps in the Kubernetes API", len(currentConfigMaps)))

	return currentConfigMaps, nil
}
