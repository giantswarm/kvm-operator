package configmap

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/v5/pkg/resource/crud"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/kvm-operator/v4/service/controller/key"
)

func (r *Resource) ApplyUpdateChange(ctx context.Context, obj, updateChange interface{}) error {
	customResource, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	configMapsToUpdate, err := toConfigMaps(updateChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if len(configMapsToUpdate) != 0 {
		r.logger.Debugf(ctx, "updating the config maps in the Kubernetes API")

		// Create the config maps in the Kubernetes API.
		namespace := key.ClusterNamespace(customResource)
		for _, configMap := range configMapsToUpdate {
			_, err := r.k8sClient.CoreV1().ConfigMaps(namespace).Update(ctx, configMap, v1.UpdateOptions{})
			if err != nil {
				return microerror.Mask(err)
			}
		}

		r.logger.Debugf(ctx, "updated the config maps in the Kubernetes API")
	} else {
		r.logger.Debugf(ctx, "the config maps do not need to be updated in the Kubernetes API")
	}

	return nil
}

func (r *Resource) NewUpdatePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*crud.Patch, error) {
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

	patch := crud.NewPatch()
	patch.SetCreateChange(create)
	patch.SetDeleteChange(delete)
	patch.SetUpdateChange(update)

	return patch, nil
}

func (r *Resource) newUpdateChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	currentConfigMaps, err := toConfigMaps(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredConfigMaps, err := toConfigMaps(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var configMapsToUpdate []*corev1.ConfigMap
	{
		r.logger.Debugf(ctx, "finding out which config maps have to be updated")

		for _, currentConfigMap := range currentConfigMaps {
			desiredConfigMap, err := getConfigMapByName(desiredConfigMaps, currentConfigMap.Name)
			if IsNotFound(err) {
				continue
			} else if err != nil {
				return nil, microerror.Mask(err)
			}

			isModified := !isEmpty(currentConfigMap) && !equals(currentConfigMap, desiredConfigMap)
			if isModified {
				configMapsToUpdate = append(configMapsToUpdate, desiredConfigMap)
			}
		}

		r.logger.Debugf(ctx, "found %d config maps that have to be updated", len(configMapsToUpdate))
	}

	return configMapsToUpdate, nil
}
