package pvc

import (
	"context"

	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"

	"github.com/giantswarm/kvm-operator/v4/service/controller/key"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var PVCs []corev1.PersistentVolumeClaim

	if key.EtcdStorageType(customObject) == "persistentVolume" {
		r.logger.Debugf(ctx, "computing the new master PVCs")

		etcdPVCs, err := r.getDesiredMasterPVCs(customObject)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		PVCs = append(PVCs, etcdPVCs...)

		r.logger.Debugf(ctx, "computed the %d new master PVCs", len(PVCs))
	} else {
		r.logger.Debugf(ctx, "not computing the new master PVCs because storage type is not 'persistentVolume'")
	}

	if key.HasHostVolumes(customObject) {
		r.logger.Debugf(ctx, "computing the new worker PVCs")

		localPVCs, err := r.getDesiredWorkerPVCs(ctx, customObject)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		PVCs = append(PVCs, localPVCs...)

		r.logger.Debugf(ctx, "computed the %d new worker PVCs", len(PVCs))
	} else {
		r.logger.Debugf(ctx, "not computing the new PVCs because no worker has defined host volumes")
	}

	return PVCs, nil
}
