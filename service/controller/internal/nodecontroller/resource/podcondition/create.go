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
	nodeReadyCondition, nodeReadyFound := key.FindNodeCondition(node, corev1.NodeReady)
	podReadyCondition, podReadyFound := key.FindPodCondition(pod, key.WorkloadClusterNodeReady)

	condition := podReadyCondition
	condition.Type = key.WorkloadClusterNodeReady // In case FindPodCondition returned an empty object
	condition.Message = nodeReadyCondition.Message
	condition.Reason = nodeReadyCondition.Reason

	if nodeReadyFound {
		if nodeReadyCondition.Status == corev1.ConditionTrue {
			condition.Status = corev1.ConditionTrue
		} else if nodeReadyCondition.Status == corev1.ConditionFalse {
			condition.Status = corev1.ConditionFalse
		} else {
			condition.Status = corev1.ConditionUnknown
		}
	} else {
		condition.Status = corev1.ConditionUnknown
	}

	transition := !podReadyFound || condition.Status != podReadyCondition.Status
	if transition {
		condition.LastTransitionTime = metav1.Now()
	}

	return condition, transition
}
