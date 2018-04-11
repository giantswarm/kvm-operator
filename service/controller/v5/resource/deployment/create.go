package deployment

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	"k8s.io/api/extensions/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/giantswarm/kvm-operator/service/controller/v5/key"
)

func (r *Resource) ApplyCreateChange(ctx context.Context, obj, createChange interface{}) error {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	deploymentsToCreate, err := toDeployments(createChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if len(deploymentsToCreate) != 0 {
		r.logger.LogCtx(ctx, "debug", "creating the deployments in the Kubernetes API")

		namespace := key.ClusterNamespace(customObject)
		for _, deployment := range deploymentsToCreate {
			_, err := r.k8sClient.Extensions().Deployments(namespace).Create(deployment)
			if apierrors.IsAlreadyExists(err) {
				// fall through
			} else if err != nil {
				return microerror.Mask(err)
			}
		}

		r.logger.LogCtx(ctx, "debug", "created the deployments in the Kubernetes API")
	} else {
		r.logger.LogCtx(ctx, "debug", "the deployments do not need to be created in the Kubernetes API")
	}

	return nil
}

func (r *Resource) newCreateChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	currentDeployments, err := toDeployments(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredDeployments, err := toDeployments(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "debug", "finding out which deployments have to be created")

	var deploymentsToCreate []*v1beta1.Deployment

	for _, desiredDeployment := range desiredDeployments {
		if !containsDeployment(currentDeployments, desiredDeployment) {
			deploymentsToCreate = append(deploymentsToCreate, desiredDeployment)
		}
	}

	r.logger.LogCtx(ctx, "debug", fmt.Sprintf("found %d deployments that have to be created", len(deploymentsToCreate)))

	return deploymentsToCreate, nil
}
