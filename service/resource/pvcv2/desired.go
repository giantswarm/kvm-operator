package pvcv2

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	apiv1 "k8s.io/client-go/pkg/api/v1"

	"github.com/giantswarm/kvm-operator/service/keyv2"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := keyv2.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var PVCs []*apiv1.PersistentVolumeClaim

	if keyv2.StorageType(customObject) == "persistentVolume" {
		r.logger.LogCtx(ctx, "debug", "computing the new PVCs")

		PVCs, err = newEtcdPVCs(customObject)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "debug", fmt.Sprintf("computed the %d new PVCs", len(PVCs)))
	} else {
		r.logger.LogCtx(ctx, "debug", "not computing the new PVCs because storage type is not 'persistentVolume'")
	}

	return PVCs, nil
}
