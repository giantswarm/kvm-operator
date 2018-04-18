package ingress

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller"
	"k8s.io/api/extensions/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/kvm-operator/service/controller/v2/key"
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
		r.logger.LogCtx(ctx, "debug", "deleting the ingresses in the Kubernetes API")

		namespace := key.ClusterNamespace(customObject)
		for _, ingress := range ingressesToDelete {
			err := r.k8sClient.Extensions().Ingresses(namespace).Delete(ingress.Name, &apismetav1.DeleteOptions{})
			if apierrors.IsNotFound(err) {
				// fall through
			} else if err != nil {
				return microerror.Mask(err)
			}
		}

		r.logger.LogCtx(ctx, "debug", "deleted the ingresses in the Kubernetes API")
	} else {
		r.logger.LogCtx(ctx, "debug", "the ingresses do not need to be deleted from the Kubernetes API")
	}

	return nil
}

func (r *Resource) NewDeletePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*controller.Patch, error) {
	delete, err := r.newDeleteChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := controller.NewPatch()
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

	r.logger.LogCtx(ctx, "debug", "finding out which ingresses have to be deleted")

	var ingressesToDelete []*v1beta1.Ingress

	for _, currentIngress := range currentIngresses {
		if containsIngress(desiredIngresses, currentIngress) {
			ingressesToDelete = append(ingressesToDelete, currentIngress)
		}
	}

	r.logger.LogCtx(ctx, "debug", fmt.Sprintf("found %d ingresses that have to be deleted", len(ingressesToDelete)))

	return ingressesToDelete, nil
}
