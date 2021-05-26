package pvc

import (
	"context"

	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/kvm-operator/service/controller/key"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var PVCs []*corev1.PersistentVolumeClaim

	if key.StorageType(customObject) == "persistentVolume" {
		r.logger.Debugf(ctx, "computing the new master PVCs")

		etcdPVCs, err := newEtcdPVCs(customObject)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		PVCs = append(PVCs, etcdPVCs...)

		r.logger.Debugf(ctx, "computed the %d new PVCs", len(PVCs))
	} else {
		r.logger.Debugf(ctx, "not computing the new PVCs because storage type is not 'persistentVolume'")
	}

	if key.HasHostVolumes(customObject) {
		r.logger.Debugf(ctx, "computing the new worker PVCs")

		// Retrieve the existing Persistent Volume in the management cluster to get the storage size
		// and create the Persistent Volume Claims for the workload cluster's workers
		pvsList, err := r.k8sClient.CoreV1().PersistentVolumes().List(ctx, metav1.ListOptions{LabelSelector: LabelMountTag})
		if err != nil {
			return nil, microerror.Mask(err)
		}

		localPVCs := newLocalPVCs(customObject, pvsList)

		PVCs = append(PVCs, localPVCs...)

		r.logger.Debugf(ctx, "computed the %d new worker PVCs", len(PVCs))
	} else {
		r.logger.Debugf(ctx, "not computing the new PVCs because no worker has defined host volumes")
	}

	return PVCs, nil
}
