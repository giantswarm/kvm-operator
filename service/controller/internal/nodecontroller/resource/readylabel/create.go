package readylabel

import (
	"context"

	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"

	"github.com/giantswarm/kvm-operator/service/controller/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	workloadNode, ok := obj.(*corev1.Node)
	if !ok {
		return microerror.Mask(invalidConfigError)
	}

	var nodeReady bool
	for _, condition := range workloadNode.Status.Conditions {
		if condition.Type == corev1.NodeReady {
			nodeReady = condition.Status == corev1.ConditionTrue
			break
		}
	}

	patch := annotationPatch{
		key:   key.AnnotationPodNodeStatus,
		value: key.PodNodeStatusReady,
	}
	if !nodeReady {
		patch.value = key.PodNodeStatusNotReady
	}

	var managementPod corev1.Pod
	err := r.managementK8sClient.Get(ctx, key.NodePodKey(r.cluster, *workloadNode), &managementPod)
	if err != nil {
		return microerror.Mask(err)
	}
	err = r.managementK8sClient.Patch(ctx, &managementPod, patch)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
