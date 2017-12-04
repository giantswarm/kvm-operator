package configmapv2

import (
	"context"
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/apis/cluster/v1alpha1"
	"github.com/giantswarm/microerror"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiv1 "k8s.io/client-go/pkg/api/v1"

	"github.com/giantswarm/kvm-operator/service/keyv2"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := keyv2.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "debug", "computing the new config maps")

	configMaps, err := r.newConfigMaps(customObject)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "debug", fmt.Sprintf("computed the %d new config maps", len(configMaps)))

	return configMaps, nil
}

func (r *Resource) newConfigMaps(customObject v1alpha1.KVMConfig) ([]*apiv1.ConfigMap, error) {
	var configMaps []*apiv1.ConfigMap

	certs, err := r.certWatcher.SearchCerts(keyv2.ClusterID(customObject))
	if err != nil {
		return nil, microerror.Mask(err)
	}

	for _, node := range customObject.Spec.Cluster.Masters {
		template, err := r.cloudConfig.NewMasterTemplate(customObject, certs, node)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		configMap, err := r.newConfigMap(customObject, template, node, keyv2.MasterID)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		configMaps = append(configMaps, configMap)
	}

	for _, node := range customObject.Spec.Cluster.Workers {
		template, err := r.cloudConfig.NewWorkerTemplate(customObject, certs, node)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		configMap, err := r.newConfigMap(customObject, template, node, keyv2.WorkerID)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		configMaps = append(configMaps, configMap)
	}

	return configMaps, nil
}

// newConfigMap creates a new Kubernetes configmap using the provided
// information. customObject is used for name and label creation. params serves
// as structure being injected into the template execution to interpolate
// variables. prefix can be either "master" or "worker" and is used to prefix
// the configmap name.
func (r *Resource) newConfigMap(customObject v1alpha1.KVMConfig, template string, node v1alpha1.ClusterNode, prefix string) (*apiv1.ConfigMap, error) {
	var newConfigMap *apiv1.ConfigMap
	{
		newConfigMap = &apiv1.ConfigMap{
			ObjectMeta: apismetav1.ObjectMeta{
				Name: keyv2.ConfigMapName(customObject, node, prefix),
				Labels: map[string]string{
					"cluster":  keyv2.ClusterID(customObject),
					"customer": keyv2.ClusterCustomer(customObject),
				},
			},
			Data: map[string]string{
				KeyUserData: template,
			},
		}
	}

	return newConfigMap, nil
}
