package configmap

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/v4/pkg/resource/crud"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

func (r *Resource) ApplyDeleteChange(ctx context.Context, obj, deleteChange interface{}) error {
	configMapsToDelete, err := toConfigMaps(deleteChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if len(configMapsToDelete) != 0 {
		r.logger.Debugf(ctx, "deleting the config maps in the Kubernetes API")

		for _, configMap := range configMapsToDelete {
			err := r.ctrlClient.Delete(ctx, configMap)
			if apierrors.IsNotFound(err) {
				// fall through
			} else if err != nil {
				return microerror.Mask(err)
			}
		}

		r.logger.Debugf(ctx, "deleted the config maps in the Kubernetes API")
	} else {
		r.logger.Debugf(ctx, "the config maps do not need to be deleted from the Kubernetes API")
	}

	return nil
}

func (r *Resource) NewDeletePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*crud.Patch, error) {
	delete, err := r.newDeleteChangeForDeletePatch(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := crud.NewPatch()
	patch.SetDeleteChange(delete)

	return patch, nil
}

func (r *Resource) newDeleteChangeForDeletePatch(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	currentConfigMaps, err := toConfigMaps(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredConfigMaps, err := toConfigMaps(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.Debugf(ctx, "finding out which config maps have to be deleted")

	var configMapsToDelete []*corev1.ConfigMap

	for _, currentConfigMap := range currentConfigMaps {
		if containsConfigMap(desiredConfigMaps, currentConfigMap) {
			configMapsToDelete = append(configMapsToDelete, currentConfigMap)
		}
	}

	r.logger.Debugf(ctx, "found %d config maps that have to be deleted", len(configMapsToDelete))

	return configMapsToDelete, nil
}

func (r *Resource) newDeleteChangeForUpdatePatch(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	currentConfigMaps, err := toConfigMaps(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredConfigMaps, err := toConfigMaps(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.Debugf(ctx, "finding out which config maps have to be deleted")

	var configMapsToDelete []*corev1.ConfigMap

	for _, currentConfigMap := range currentConfigMaps {
		if !containsConfigMap(desiredConfigMaps, currentConfigMap) {
			configMapsToDelete = append(configMapsToDelete, currentConfigMap)
		}
	}

	r.logger.Debugf(ctx, "found %d config maps that have to be deleted", len(configMapsToDelete))

	return configMapsToDelete, nil
}
