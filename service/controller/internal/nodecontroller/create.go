package nodecontroller

import (
	"context"

	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/giantswarm/kvm-operator/v4/service/controller/key"
)

func (c *Controller) ensureCreated(ctx context.Context, workloadNode corev1.Node) (reconcile.Result, error) {
	var managementPod corev1.Pod
	err := c.managementClient.Get(ctx, key.NodePodObjectKey(c.cluster, workloadNode), &managementPod)
	if errors.IsNotFound(err) {
		// assume the pod is already deleted
		c.logger.Debugf(ctx, "node pod not found")
		return key.RequeueNone, nil
	} else if err != nil {
		return key.RequeueErrorShort, microerror.Mask(err)
	}

	condition, shouldUpdate := calculateCreatedPodNodeCondition(workloadNode, managementPod)
	if !shouldUpdate {
		return key.RequeueNone, nil
	}

	c.logger.Debugf(ctx, "patching pod node status condition to %#v", condition)
	err = c.managementClient.Status().Patch(ctx, &managementPod, podConditionPatch{PodCondition: condition})
	if err != nil {
		return key.RequeueErrorShort, microerror.Mask(err)
	}

	return key.RequeueNone, nil
}

func calculateCreatedPodNodeCondition(node corev1.Node, pod corev1.Pod) (corev1.PodCondition, bool) {
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
