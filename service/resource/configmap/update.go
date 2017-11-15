package configmap

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/framework"
	apiv1 "k8s.io/client-go/pkg/api/v1"

	"github.com/giantswarm/kvm-operator/service/key"
	"github.com/giantswarm/kvm-operator/service/messagecontext"
)

func (r *Resource) ApplyUpdateChange(ctx context.Context, obj, updateChange interface{}) error {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	configMapsToUpdate, err := toConfigMaps(updateChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if len(configMapsToUpdate) != 0 {
		r.logger.Log("cluster", key.ClusterID(customObject), "debug", "updating the config maps in the Kubernetes API")

		// Create the config maps in the Kubernetes API.
		namespace := key.ClusterNamespace(customObject)
		for _, configMap := range configMapsToUpdate {
			_, err := r.k8sClient.CoreV1().ConfigMaps(namespace).Update(configMap)
			if err != nil {
				return microerror.Mask(err)
			}
		}

		r.logger.Log("cluster", key.ClusterID(customObject), "debug", "updated the config maps in the Kubernetes API")
	} else {
		r.logger.Log("cluster", key.ClusterID(customObject), "debug", "the config maps do not need to be updated in the Kubernetes API")
	}

	return nil
}

func (r *Resource) NewUpdatePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*framework.Patch, error) {
	create, err := r.newCreateChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	delete, err := r.newDeleteChangeForUpdatePatch(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	update, err := r.newUpdateChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := framework.NewPatch()
	patch.SetCreateChange(create)
	patch.SetDeleteChange(delete)
	patch.SetUpdateChange(update)

	return patch, nil
}

func (r *Resource) newUpdateChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	currentConfigMaps, err := toConfigMaps(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredConfigMaps, err := toConfigMaps(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var configMapsToUpdate []*apiv1.ConfigMap
	{
		r.logger.Log("cluster", key.ClusterID(customObject), "debug", "finding out which config maps have to be updated")

		for _, currentConfigMap := range currentConfigMaps {
			desiredConfigMap, err := getConfigMapByName(desiredConfigMaps, currentConfigMap.Name)
			if IsNotFound(err) {
				continue
			} else if err != nil {
				return nil, microerror.Mask(err)
			}

			if isConfigMapModified(desiredConfigMap, currentConfigMap) {
				m, ok := messagecontext.FromContext(ctx)
				if ok {
					m.ConfigMapNames = append(m.ConfigMapNames, desiredConfigMap.Name)
				}
				configMapsToUpdate = append(configMapsToUpdate, desiredConfigMap)
			}
		}

		r.logger.Log("cluster", key.ClusterID(customObject), "debug", fmt.Sprintf("found %d config maps that have to be updated", len(configMapsToUpdate)))
	}

	return configMapsToUpdate, nil
}
