package pvc

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/kvm-operator/service/controller/key"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.Debugf(ctx, "looking for PVCs in the Kubernetes API")

	namespace := key.ClusterNamespace(customObject)
	pvcs, err := r.k8sClient.CoreV1().PersistentVolumeClaims(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", "cluster", key.ClusterID(customObject)),
	})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.Debugf(ctx, "found %d PVCs in the Kubernetes API", len(pvcs.Items))

	return pvcs.Items, nil
}
