package configmap

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiv1 "k8s.io/client-go/pkg/api/v1"

	"github.com/giantswarm/kvm-operator/service/key"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	fmt.Printf("custom object state start\n")
	fmt.Printf("%#v\n", customObject)
	fmt.Printf("custom object state end\n")

	r.logger.Log("cluster", key.ClusterID(customObject), "debug", "looking for a list of config maps in the Kubernetes API")

	var currentConfigMaps []*apiv1.ConfigMap
	{
		namespace := key.ClusterNamespace(customObject)
		configMapList, err := r.k8sClient.CoreV1().ConfigMaps(namespace).List(apismetav1.ListOptions{})
		if err != nil {
			return nil, microerror.Mask(err)
		} else {
			r.logger.Log("cluster", key.ClusterID(customObject), "debug", "found a list of config maps in the Kubernetes API")

			for _, c := range configMapList.Items {
				currentConfigMaps = append(currentConfigMaps, &c)
			}
		}
	}

	r.logger.Log("cluster", key.ClusterID(customObject), "debug", fmt.Sprintf("found a list of %d config maps in the Kubernetes API", len(currentConfigMaps)))

	return currentConfigMaps, nil
}
