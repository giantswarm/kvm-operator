package podcondition

import (
	"context"

	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/kvm-operator/service/controller/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	workloadNode, err := key.ToNode(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	var managementPod corev1.Pod
	err = r.managementK8sClient.Get(ctx, key.NodePodObjectKey(r.cluster, workloadNode), &managementPod)
	if err != nil {
		return microerror.Mask(err)
	}

	condition, shouldUpdate := calculatePodNodeCondition(workloadNode, managementPod)
	if !shouldUpdate {
		return nil
	}

	r.logger.Debugf(ctx, "patching pod node status condition to %#v", condition)
	err = r.managementK8sClient.Status().Patch(ctx, &managementPod, podConditionPatch{PodCondition: condition})
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func calculatePodNodeCondition(node corev1.Node, pod corev1.Pod) (corev1.PodCondition, bool) {
	nodeCondition, _ := key.FindNodeCondition(node, corev1.NodeReady)
	currentPodCondition, currentPodConditionFound := key.FindPodCondition(pod, key.WorkloadClusterNodeReady)

	desiredPodCondition := corev1.PodCondition{
		Type:    key.WorkloadClusterNodeReady,
		Reason:  nodeCondition.Reason,
		Message: nodeCondition.Message,
	}

	switch nodeCondition.Status {
	case corev1.ConditionTrue, corev1.ConditionFalse:
		desiredPodCondition.Status = nodeCondition.Status
	default:
		desiredPodCondition.Status = corev1.ConditionUnknown
	}

	transition := !currentPodConditionFound || desiredPodCondition.Status != currentPodCondition.Status
	if transition {
		desiredPodCondition.LastTransitionTime = metav1.Now()
	}

	return desiredPodCondition, transition
}
