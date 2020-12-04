package ingress

import (
	"context"

	"github.com/giantswarm/microerror"
	"k8s.io/api/networking/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/kvm-operator/service/controller/key"
)

func (r *Resource) ApplyCreateChange(ctx context.Context, obj, createChange interface{}) error {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	ingressesToCreate, err := toIngresses(createChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if len(ingressesToCreate) != 0 {
		r.logger.Debugf(ctx, "creating the ingresses in the Kubernetes API")

		namespace := key.ClusterNamespace(customObject)
		for _, ingress := range ingressesToCreate {
			_, err := r.k8sClient.NetworkingV1beta1().Ingresses(namespace).Create(ctx, ingress, v1.CreateOptions{})
			if apierrors.IsAlreadyExists(err) {
				// fall through
			} else if err != nil {
				return microerror.Mask(err)
			}
		}

		r.logger.Debugf(ctx, "created the ingresses in the Kubernetes API")
	} else {
		r.logger.Debugf(ctx, "the ingresses do not need to be created in the Kubernetes API")
	}

	return nil
}

func (r *Resource) newCreateChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	currentIngresses, err := toIngresses(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredIngresses, err := toIngresses(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.Debugf(ctx, "finding out which ingresses have to be created")

	var ingressesToCreate []*v1beta1.Ingress

	for _, desiredIngress := range desiredIngresses {
		if !containsIngress(currentIngresses, desiredIngress) {
			ingressesToCreate = append(ingressesToCreate, desiredIngress)
		}
	}

	r.logger.Debugf(ctx, "found %d ingresses that have to be created", len(ingressesToCreate))

	return ingressesToCreate, nil
}
