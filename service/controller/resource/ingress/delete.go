package ingress

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/v5/pkg/resource/crud"
	"k8s.io/api/networking/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/kvm-operator/v4/service/controller/key"
)

func (r *Resource) ApplyDeleteChange(ctx context.Context, obj, deleteChange interface{}) error {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	ingressesToDelete, err := toIngresses(deleteChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if len(ingressesToDelete) != 0 {
		r.logger.Debugf(ctx, "deleting the ingresses in the Kubernetes API")

		namespace := key.ClusterNamespace(customObject)
		for _, ingress := range ingressesToDelete {
			err := r.k8sClient.NetworkingV1beta1().Ingresses(namespace).Delete(ctx, ingress.Name, metav1.DeleteOptions{})
			if apierrors.IsNotFound(err) {
				// fall through
			} else if err != nil {
				return microerror.Mask(err)
			}
		}

		r.logger.Debugf(ctx, "deleted the ingresses in the Kubernetes API")
	} else {
		r.logger.Debugf(ctx, "the ingresses do not need to be deleted from the Kubernetes API")
	}

	return nil
}

func (r *Resource) NewDeletePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*crud.Patch, error) {
	delete, err := r.newDeleteChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := crud.NewPatch()
	patch.SetDeleteChange(delete)

	return patch, nil
}

func (r *Resource) newDeleteChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	currentIngresses, err := toIngresses(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredIngresses, err := toIngresses(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.Debugf(ctx, "finding out which ingresses have to be deleted")

	var ingressesToDelete []*v1beta1.Ingress

	for _, currentIngress := range currentIngresses {
		if containsIngress(desiredIngresses, currentIngress) {
			ingressesToDelete = append(ingressesToDelete, currentIngress)
		}
	}

	r.logger.Debugf(ctx, "found %d ingresses that have to be deleted", len(ingressesToDelete))

	return ingressesToDelete, nil
}
