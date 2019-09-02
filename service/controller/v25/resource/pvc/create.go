package pvc

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	apiv1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/giantswarm/kvm-operator/service/controller/v25/key"
)

func (r *Resource) ApplyCreateChange(ctx context.Context, obj, createChange interface{}) error {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	pvcsToCreate, err := toPVCs(createChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if len(pvcsToCreate) != 0 {
		r.logger.LogCtx(ctx, "level", "debug", "message", "creating the PVCs in the Kubernetes API")

		namespace := key.ClusterNamespace(customObject)
		for _, PVC := range pvcsToCreate {
			_, err := r.k8sClient.Core().PersistentVolumeClaims(namespace).Create(PVC)
			if apierrors.IsAlreadyExists(err) {
				// fall through
			} else if err != nil {
				return microerror.Mask(err)
			}
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "created the PVCs in the Kubernetes API")
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "the PVCs do not need to be created in the Kubernetes API")
	}

	return nil
}

func (r *Resource) newCreateChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	currentPVCs, err := toPVCs(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredPVCs, err := toPVCs(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "finding out which PVCs have to be created")

	var pvcsToCreate []*apiv1.PersistentVolumeClaim

	for _, desiredPVC := range desiredPVCs {
		if !containsPVC(currentPVCs, desiredPVC) {
			pvcsToCreate = append(pvcsToCreate, desiredPVC)
		}
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found %d PVCs that have to be created", len(pvcsToCreate)))

	return pvcsToCreate, nil
}
