package configmap

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/v5/pkg/controller/context/resourcecanceledcontext"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/kvm-operator/pkg/label"
	"github.com/giantswarm/kvm-operator/pkg/project"
	"github.com/giantswarm/kvm-operator/service/controller/key"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	customResource, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	if key.IsDeleted(&customResource) {
		r.logger.Debugf(ctx, "redirecting responsibility of deletion of config maps to namespace termination")
		resourcecanceledcontext.SetCanceled(ctx)
		r.logger.Debugf(ctx, "canceling resource")

		return nil, nil
	}

	r.logger.Debugf(ctx, "looking for a list of config maps in the Kubernetes API")

	var currentConfigMaps []*corev1.ConfigMap
	{
		namespace := key.ClusterNamespace(customResource)

		lo := metav1.ListOptions{
			LabelSelector: fmt.Sprintf("%s=%s", label.ManagedBy, project.Name()),
		}

		configMapList, err := r.k8sClient.CoreV1().ConfigMaps(namespace).List(ctx, lo)
		if err != nil {
			return nil, microerror.Mask(err)
		} else {
			r.logger.Debugf(ctx, "found a list of config maps in the Kubernetes API")

			for _, item := range configMapList.Items {
				c := item
				currentConfigMaps = append(currentConfigMaps, &c)
			}
		}
	}

	r.logger.Debugf(ctx, "found a list of %d config maps in the Kubernetes API", len(currentConfigMaps))

	return currentConfigMaps, nil
}
