package ingress

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"

	"github.com/giantswarm/kvm-operator/service/key"
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
		r.logger.Log("cluster", key.ClusterID(customObject), "debug", "creating the ingresses in the Kubernetes API")

		namespace := key.ClusterNamespace(customObject)
		for _, ingress := range ingressesToCreate {
			_, err := r.k8sClient.Extensions().Ingresses(namespace).Create(ingress)
			if apierrors.IsAlreadyExists(err) {
				// fall through
			} else if err != nil {
				return microerror.Mask(err)
			}
		}

		r.logger.Log("cluster", key.ClusterID(customObject), "debug", "created the ingresses in the Kubernetes API")
	} else {
		r.logger.Log("cluster", key.ClusterID(customObject), "debug", "the ingresses do not need to be created in the Kubernetes API")
	}

	return nil
}

func (r *Resource) newCreateChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	currentIngresses, err := toIngresses(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredIngresses, err := toIngresses(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.Log("cluster", key.ClusterID(customObject), "debug", "finding out which ingresses have to be created")

	var ingressesToCreate []*v1beta1.Ingress

	for _, desiredIngress := range desiredIngresses {
		if !containsIngress(currentIngresses, desiredIngress) {
			ingressesToCreate = append(ingressesToCreate, desiredIngress)
		}
	}

	r.logger.Log("cluster", key.ClusterID(customObject), "debug", fmt.Sprintf("found %d ingresses that have to be created", len(ingressesToCreate)))

	return ingressesToCreate, nil
}
