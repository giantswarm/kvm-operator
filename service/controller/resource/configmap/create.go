package configmap

import (
	"context"

	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

func (r *Resource) ApplyCreateChange(ctx context.Context, obj, createChange interface{}) error {
	configMapsToCreate, err := toConfigMaps(createChange)
	if err != nil {
		return microerror.Mask(err)
	}

	// Create the config maps in the Kubernetes API.
	if len(configMapsToCreate) != 0 {
		r.logger.Debugf(ctx, "creating the config maps in the Kubernetes API")

		for _, configMap := range configMapsToCreate {
			err := r.ctrlClient.Create(ctx, configMap)
			if apierrors.IsAlreadyExists(err) {
				// fall through
			} else if err != nil {
				return microerror.Mask(err)
			}
		}

		r.logger.Debugf(ctx, "created the config maps in the Kubernetes API")
	} else {
		r.logger.Debugf(ctx, "the config maps do not need to be created in the Kubernetes API")
	}

	return nil
}

func (r *Resource) newCreateChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	currentConfigMaps, err := toConfigMaps(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredConfigMaps, err := toConfigMaps(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.Debugf(ctx, "finding out which config maps have to be created")

	var configMapsToCreate []*corev1.ConfigMap

	for _, desiredConfigMap := range desiredConfigMaps {
		if !containsConfigMap(currentConfigMaps, desiredConfigMap) {
			configMapsToCreate = append(configMapsToCreate, desiredConfigMap)
		}
	}

	r.logger.Debugf(ctx, "found %d config maps that have to be created", len(configMapsToCreate))

	return configMapsToCreate, nil
}
