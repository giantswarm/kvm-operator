package pvc

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	apiv1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/kvm-operator/service/kvmconfig/v9/key"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "looking for PVCs in the Kubernetes API")

	var PVCs []*apiv1.PersistentVolumeClaim

	namespace := key.ClusterNamespace(customObject)
	pvcNames := key.PVCNames(customObject)

	for _, name := range pvcNames {
		manifest, err := r.k8sClient.Core().PersistentVolumeClaims(namespace).Get(name, apismetav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "did not find a PVC in the Kubernetes API")
			// fall through
		} else if err != nil {
			return nil, microerror.Mask(err)
		} else {
			r.logger.LogCtx(ctx, "level", "debug", "message", "found a PVC in the Kubernetes API")
			PVCs = append(PVCs, manifest)
		}
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found %d PVCs in the Kubernetes API", len(PVCs)))

	return PVCs, nil
}
