package nodecontroller

import (
	"context"
	"time"

	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/giantswarm/kvm-operator/service/controller/key"
)

func (c *Controller) ensureCreated(ctx context.Context, workloadNode corev1.Node) (reconcile.Result, error) {
	var managementPod corev1.Pod
	err := c.managementK8sClient.Get(ctx, key.NodePodObjectKey(c.cluster, workloadNode), &managementPod)
	if errors.IsNotFound(err) {
		return reconcile.Result{Requeue: false}, microerror.Mask(err)
	} else if err != nil {
		return reconcile.Result{
			Requeue:      true,
			RequeueAfter: time.Second * 10,
		}, microerror.Mask(err)
	}

	condition, shouldUpdate := calculateCreatedPodNodeCondition(workloadNode, managementPod)
	if !shouldUpdate {
		return reconcile.Result{Requeue: false}, nil
	}

	c.logger.Debugf(ctx, "patching pod node status condition to %#v", condition)
	err = c.managementK8sClient.Status().Patch(ctx, &managementPod, podConditionPatch{PodCondition: condition})
	if err != nil {
		return reconcile.Result{
			Requeue:      true,
			RequeueAfter: time.Second * 10,
		}, microerror.Mask(err)
	}

	return reconcile.Result{}, nil
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

	if node.Spec.Unschedulable && desiredPodCondition.Status != corev1.ConditionFalse {
		desiredPodCondition.Status = corev1.ConditionFalse
		desiredPodCondition.Reason = "NodeUnschedulable"
		desiredPodCondition.Message = "node is unschedulable"
	}

	transition := !currentPodConditionFound || desiredPodCondition.Status != currentPodCondition.Status
	if transition {
		desiredPodCondition.LastTransitionTime = metav1.Now()
	}

	return desiredPodCondition, transition
}
