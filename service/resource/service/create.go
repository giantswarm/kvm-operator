package service

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apiv1 "k8s.io/client-go/pkg/api/v1"

	"github.com/giantswarm/kvm-operator/service/key"
)

func (r *Resource) ApplyCreateChange(ctx context.Context, obj, createChange interface{}) error {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	servicesToCreate, err := toServices(createChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if len(servicesToCreate) != 0 {
		r.logger.LogCtx(ctx, "debug", "creating the services in the Kubernetes API")

		namespace := key.ClusterNamespace(customObject)
		for _, service := range servicesToCreate {
			_, err := r.k8sClient.CoreV1().Services(namespace).Create(service)
			if apierrors.IsAlreadyExists(err) {
				// fall through
			} else if err != nil {
				return microerror.Mask(err)
			}
		}

		r.logger.LogCtx(ctx, "debug", "created the services in the Kubernetes API")
	} else {
		r.logger.LogCtx(ctx, "debug", "the services do not need to be created in the Kubernetes API")
	}

	return nil
}

func (r *Resource) newCreateChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	currentServices, err := toServices(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredServices, err := toServices(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "debug", "finding out which services have to be created")

	var servicesToCreate []*apiv1.Service

	for _, desiredService := range desiredServices {
		if !containsService(currentServices, desiredService) {
			servicesToCreate = append(servicesToCreate, desiredService)
		}
	}

	r.logger.LogCtx(ctx, "debug", fmt.Sprintf("found %d services that have to be created", len(servicesToCreate)))

	return servicesToCreate, nil
}
