package pvc

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	apiv1 "k8s.io/client-go/pkg/api/v1"

	"github.com/giantswarm/kvm-operator/service/key"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var PVCs []*apiv1.PersistentVolumeClaim

	if key.StorageType(customObject) == "persistentVolume" {
		r.logger.Log("cluster", key.ClusterID(customObject), "debug", "computing the new PVCs")

		PVCs, err = newEtcdPVCs(customObject)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		r.logger.Log("cluster", key.ClusterID(customObject), "debug", fmt.Sprintf("computed the %d new PVCs", len(PVCs)))
	} else {
		r.logger.Log("cluster", key.ClusterID(customObject), "debug", "not computing the new PVCs because storage type is not 'persistentVolume'")
	}

	return PVCs, nil
}
