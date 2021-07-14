package pvc

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/v5/pkg/resource/crud"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/kvm-operator/v4/service/controller/key"
)

func (r *Resource) ApplyDeleteChange(ctx context.Context, obj, deleteChange interface{}) error {
	pvcsToDelete, err := toPVCs(deleteChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if len(pvcsToDelete) != 0 {
		r.logger.Debugf(ctx, "deleting the PVCs in the Kubernetes API")

		persistentVolumeSelector, err := labels.Parse(key.LabelMountTag)
		if err != nil {
			return microerror.Mask(err)
		}

		var persistentVolumes corev1.PersistentVolumeList
		err = r.ctrlClient.List(ctx, &persistentVolumes, &client.ListOptions{
			LabelSelector: persistentVolumeSelector,
		})
		if err != nil {
			return microerror.Mask(err)
		}

		for _, pvc := range pvcsToDelete {
			err := r.ctrlClient.Delete(ctx, pvc.DeepCopy(), &client.DeleteOptions{})
			if apierrors.IsNotFound(err) {
				// fall through
			} else if err != nil {
				return microerror.Mask(err)
			}

			// the following logic only applies to host volume PVCs (local storage class), etcd PVCs are automatically managed
			if pvc.Spec.StorageClassName == nil || *pvc.Spec.StorageClassName != LocalStorageClass {
				continue
			}

			// Find PV with claim pointing to the deleted PVC
			for _, persistentVolume := range persistentVolumes.Items {
				if persistentVolume.Spec.ClaimRef != nil &&
					persistentVolume.Spec.ClaimRef.Name == pvc.Name &&
					persistentVolume.Spec.ClaimRef.Namespace == pvc.Namespace {
					persistentVolume := persistentVolume.DeepCopy()
					// Remove the claim
					persistentVolume.Spec.ClaimRef = nil

					err = r.ctrlClient.Update(ctx, persistentVolume, &client.UpdateOptions{})
					if err != nil {
						return microerror.Mask(err)
					}

					r.logger.Debugf(ctx, "unbound PV %#q from PVC %#q", persistentVolume.Name, pvc.Name)
					break
				}
			}
		}

		r.logger.Debugf(ctx, "deleted the PVCs in the Kubernetes API")
	} else {
		r.logger.Debugf(ctx, "the PVCs do not need to be deleted from the Kubernetes API")
	}

	return nil
}

func (r *Resource) NewDeletePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*crud.Patch, error) {
	deleteChange, err := r.newDeleteChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := crud.NewPatch()
	patch.SetDeleteChange(deleteChange)

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

	var pvcsToDelete []corev1.PersistentVolumeClaim

	for _, currentPVC := range currentPVCs {
		if containsPVC(desiredPVCs, currentPVC) {
			pvcsToDelete = append(pvcsToDelete, currentPVC)
		}
	}

	r.logger.Debugf(ctx, "found %d PVCs that have to be deleted", len(pvcsToDelete))

	return pvcsToDelete, nil
}
