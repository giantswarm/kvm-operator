package pvc

import (
	"context"

	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/kvm-operator/service/controller/key"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.Debugf(ctx, "looking for PVCs in the Kubernetes API")

	var PVCs []*corev1.PersistentVolumeClaim

	namespace := key.ClusterNamespace(customObject)
	pvcNames := key.PVCNames(customObject)

	for _, name := range pvcNames {
		var manifest corev1.PersistentVolumeClaim
		err := r.ctrlClient.Get(ctx, client.ObjectKey{
			Namespace: namespace,
			Name:      name,
		}, &manifest)
		if apierrors.IsNotFound(err) {
			r.logger.Debugf(ctx, "did not find a PVC in the Kubernetes API")
			// fall through
		} else if err != nil {
			return nil, microerror.Mask(err)
		} else {
			r.logger.Debugf(ctx, "found a PVC in the Kubernetes API")
			PVCs = append(PVCs, &manifest)
		}
	}

	r.logger.Debugf(ctx, "found %d PVCs in the Kubernetes API", len(PVCs))

	return PVCs, nil
}
