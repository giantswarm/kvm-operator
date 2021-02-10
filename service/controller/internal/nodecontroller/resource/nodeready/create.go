package nodeready

import (
	"context"

	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"

	"github.com/giantswarm/kvm-operator/service/controller/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	workloadNode, err := key.ToNode(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	nodeReady := key.NodeIsReady(workloadNode)

	var managementPod corev1.Pod
	err = r.managementK8sClient.Get(ctx, key.NodePodObjectKey(r.cluster, workloadNode), &managementPod)
	if err != nil {
		return microerror.Mask(err)
	}

	var patch annotationPatch
	{
		patch = annotationPatch{
			key:   key.AnnotationPodNodeStatus,
			value: key.PodNodeStatusReady,
		}
		if !nodeReady {
			patch.value = key.PodNodeStatusNotReady
		}
	}

	existingAnnotation, ok := managementPod.Annotations[key.AnnotationPodNodeStatus]
	if ok && existingAnnotation == patch.value {
		return nil
	}

	r.logger.Debugf(ctx, "patching node pod with node status annotation %#q", patch.value)
	err = r.managementK8sClient.Patch(ctx, &managementPod, patch)
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.Debugf(ctx, "patched node pod")

	return nil
}
