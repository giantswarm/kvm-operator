package configmap

import (
	"context"
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"
	apiv1 "k8s.io/api/core/v1"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/kvm-operator/pkg/label"
	"github.com/giantswarm/kvm-operator/pkg/project"
<<<<<<< HEAD
	"github.com/giantswarm/kvm-operator/service/controller/v24patch1/key"
=======
	"github.com/giantswarm/kvm-operator/service/controller/v24/key"
>>>>>>> c4c6c79d... copy v24 to v24patch1
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	customResource, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "computing the new config maps")

	configMaps, err := r.newConfigMaps(customResource)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("computed the %d new config maps", len(configMaps)))

	return configMaps, nil
}

func (r *Resource) newConfigMaps(customResource v1alpha1.KVMConfig) ([]*apiv1.ConfigMap, error) {
	var configMaps []*apiv1.ConfigMap

	certs, err := r.certsSearcher.SearchCluster(key.ClusterID(customResource))
	if err != nil {
		return nil, microerror.Mask(err)
	}

	keys, err := r.keyWatcher.SearchCluster(key.ClusterID(customResource))
	if err != nil {
		return nil, microerror.Mask(err)
	}

	for _, node := range customResource.Spec.Cluster.Masters {
		nodeIdx, exists := key.NodeIndex(customResource, node.ID)
		if !exists {
			return nil, microerror.Maskf(notFoundError, fmt.Sprintf("node index for master (%q) is not available", node.ID))
		}

		template, err := r.cloudConfig.NewMasterTemplate(customResource, certs, node, keys, nodeIdx)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		configMap, err := r.newConfigMap(customResource, template, node, key.MasterID)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		configMaps = append(configMaps, configMap)
	}

	for _, node := range customResource.Spec.Cluster.Workers {
		nodeIdx, exists := key.NodeIndex(customResource, node.ID)
		if !exists {
			return nil, microerror.Maskf(notFoundError, fmt.Sprintf("node index for worker (%q) is not available", node.ID))
		}

		template, err := r.cloudConfig.NewWorkerTemplate(customResource, certs, node, nodeIdx)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		configMap, err := r.newConfigMap(customResource, template, node, key.WorkerID)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		configMaps = append(configMaps, configMap)
	}

	return configMaps, nil
}

// newConfigMap creates a new Kubernetes configmap using the provided
// information. customResource is used for name and label creation. params
// serves as structure being injected into the template execution to interpolate
// variables. prefix can be either "master" or "worker" and is used to prefix
// the configmap name.
func (r *Resource) newConfigMap(customResource v1alpha1.KVMConfig, template string, node v1alpha1.ClusterNode, prefix string) (*apiv1.ConfigMap, error) {
	var newConfigMap *apiv1.ConfigMap
	{
		newConfigMap = &apiv1.ConfigMap{
			ObjectMeta: apismetav1.ObjectMeta{
				Name: key.ConfigMapName(customResource, node, prefix),
				Labels: map[string]string{
					"cluster":          key.ClusterID(customResource),
					"customer":         key.ClusterCustomer(customResource),
					label.Cluster:      key.ClusterID(customResource),
					label.ManagedBy:    project.Name(),
					label.Organization: key.ClusterCustomer(customResource),
				},
			},
			Data: map[string]string{
				KeyUserData: template,
			},
		}
	}

	return newConfigMap, nil
}
