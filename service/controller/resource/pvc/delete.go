package pvc

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/v5/pkg/resource/crud"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/kvm-operator/service/controller/key"
)

func (r *Resource) ApplyDeleteChange(ctx context.Context, obj, deleteChange interface{}) error {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	pvcsToDelete, err := toPVCs(deleteChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if len(pvcsToDelete) != 0 {
		r.logger.Debugf(ctx, "deleting the PVCs in the Kubernetes API")

		pvsList, err := r.k8sClient.CoreV1().PersistentVolumes().List(ctx, metav1.ListOptions{LabelSelector: LabelMountTag})
		if err != nil {
			return microerror.Mask(err)
		}

		namespace := key.ClusterNamespace(customObject)
		for _, pvc := range pvcsToDelete {
			err := r.k8sClient.CoreV1().PersistentVolumeClaims(namespace).Delete(ctx, pvc.Name, metav1.DeleteOptions{})
			if apierrors.IsNotFound(err) {
				// fall through
			} else if err != nil {
				return microerror.Mask(err)
			}

			// in case of local storage we need to manually remove the claim
			storageClassName := *pvc.Spec.StorageClassName
			if storageClassName != LocalStorageClass {
				continue
			}

			var boundPV *corev1.PersistentVolume
			for _, pv := range pvsList.Items {
				ref := pv.Spec.ClaimRef
				if ref.Name != pvc.Name && ref.Namespace != namespace {
					continue
				}

				boundPV = pv.DeepCopy()
			}

			if boundPV == nil || boundPV.Spec.ClaimRef == nil {
				continue
			}

			boundPV.Spec.ClaimRef = nil
			_, err = r.k8sClient.CoreV1().PersistentVolumes().Update(ctx, boundPV, metav1.UpdateOptions{})
			if err != nil {
				return microerror.Mask(err)
			}

			r.logger.Debugf(ctx, "unbound PV %s from pvc %s", boundPV.Name, pvc.Name)
		}

		r.logger.Debugf(ctx, "deleted the PVCs in the Kubernetes API")
	} else {
		r.logger.Debugf(ctx, "the PVCs do not need to be deleted from the Kubernetes API")
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
	currentPVCs, err := toPVCs(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredPVCs, err := toPVCs(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.Debugf(ctx, "finding out which PVCs have to be deleted")

	var pvcsToDelete []*corev1.PersistentVolumeClaim

	for _, currentPVC := range currentPVCs {
		if containsPVC(desiredPVCs, currentPVC) {
			pvcsToDelete = append(pvcsToDelete, currentPVC)
		}
	}

	r.logger.Debugf(ctx, "found %d PVCs that have to be deleted", len(pvcsToDelete))

	return pvcsToDelete, nil
}
