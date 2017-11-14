package configmap

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
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

	r.logger.Log("cluster", key.ClusterID(customObject), "debug", "looking for config maps in the Kubernetes API")

	var configMaps []*apiv1.ConfigMap

	namespace := key.ClusterNamespace(customObject)
	// TODO we have to fetch all config maps within the namespace.
	configMapNames := key.ConfigMapNames(customObject)

	fmt.Printf("%#v\n", configMapNames)

	for _, name := range configMapNames {
		manifest, err := r.k8sClient.CoreV1().ConfigMaps(namespace).Get(name, apismetav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			r.logger.Log("cluster", key.ClusterID(customObject), "debug", "did not find a config map in the Kubernetes API")
			fmt.Printf("current state start\n")
			fmt.Printf("%#v\n", name)
			fmt.Printf("current state end\n")
			// fall through
		} else if err != nil {
			return nil, microerror.Mask(err)
		} else {
			r.logger.Log("cluster", key.ClusterID(customObject), "debug", "found a config map in the Kubernetes API")
			fmt.Printf("current state start\n")
			fmt.Printf("%#v\n", manifest)
			fmt.Printf("current state end\n")

			configMaps = append(configMaps, manifest)
		}
	}

	r.logger.Log("cluster", key.ClusterID(customObject), "debug", fmt.Sprintf("found %d config maps in the Kubernetes API", len(configMaps)))

	return configMaps, nil
}
