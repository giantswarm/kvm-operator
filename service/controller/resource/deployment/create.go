package deployment

import (
	"context"

	"github.com/giantswarm/microerror"
	appsv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/kvm-operator/service/controller/key"
)

func (r *Resource) ApplyCreateChange(ctx context.Context, obj, createChange interface{}) error {
	customResource, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	deploymentsToCreate, err := toDeployments(createChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if len(deploymentsToCreate) != 0 {
		r.logger.Debugf(ctx, "creating the deployments in the Kubernetes API")

		namespace := key.ClusterNamespace(customResource)
		for _, deployment := range deploymentsToCreate {
			_, err := r.k8sClient.AppsV1().Deployments(namespace).Create(ctx, deployment, metav1.CreateOptions{})
			if apierrors.IsAlreadyExists(err) {
				// fall through
			} else if err != nil {
				return microerror.Mask(err)
			}
		}

		r.logger.Debugf(ctx, "created the deployments in the Kubernetes API")
	} else {
		r.logger.Debugf(ctx, "the deployments do not need to be created in the Kubernetes API")
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

	r.logger.Debugf(ctx, "finding out which deployments have to be created")

	var deploymentsToCreate []*appsv1.Deployment

	for _, desiredDeployment := range desiredDeployments {
		if !containsDeployment(currentDeployments, desiredDeployment) {
			deploymentsToCreate = append(deploymentsToCreate, desiredDeployment)
		}
	}

	r.logger.Debugf(ctx, "found %d deployments that have to be created", len(deploymentsToCreate))

	return deploymentsToCreate, nil
}
