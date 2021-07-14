package pvc

import (
	"context"

	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/kvm-operator/v4/service/controller/key"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.Debugf(ctx, "looking for PVCs in the Kubernetes API")

	var pvcs corev1.PersistentVolumeClaimList
	err = r.ctrlClient.List(ctx, &pvcs, &client.ListOptions{
		LabelSelector: labels.SelectorFromSet(map[string]string{
			key.LegacyLabelCluster: key.ClusterID(customObject),
		}),
		Namespace: key.ClusterNamespace(customObject),
	})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.Debugf(ctx, "found %d PVCs in the Kubernetes API", len(pvcs.Items))

	return pvcs.Items, nil
}
