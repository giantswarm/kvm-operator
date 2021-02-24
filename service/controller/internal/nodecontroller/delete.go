package nodecontroller

import (
	"context"

	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/giantswarm/kvm-operator/service/controller/key"
)

func (c *Controller) ensureDeleted(ctx context.Context, workloadNode corev1.Node) (reconcile.Result, error) {
	var managementPod corev1.Pod
	err := c.managementK8sClient.Get(ctx, key.NodePodObjectKey(c.cluster, workloadNode), &managementPod)
	if errors.IsNotFound(err) {
		// assume the pod is already deleted
		c.logger.Debugf(ctx, "node pod not found")
		return key.RequeueNone, nil
	} else if err != nil {
		return key.RequeueShort, microerror.Mask(err)
	}

	condition, shouldUpdate := calculateDeletedPodNodeCondition(managementPod)
	if !shouldUpdate {
		return key.RequeueNone, nil
	}

	c.logger.Debugf(ctx, "patching pod node status condition to %#v", condition)
	err = c.managementK8sClient.Status().Patch(ctx, &managementPod, podConditionPatch{PodCondition: condition})
	if err != nil {
		return key.RequeueShort, microerror.Mask(err)
	}

	return key.RequeueNone, nil
}

func calculateDeletedPodNodeCondition(pod corev1.Pod) (corev1.PodCondition, bool) {
	currentPodCondition, currentPodConditionFound := key.FindPodCondition(pod, key.WorkloadClusterNodeReady)

	desiredPodCondition := corev1.PodCondition{
		Type:    key.WorkloadClusterNodeReady,
		Reason:  "NodeDeleted",
		Message: "node deleted in workload cluster",
		Status:  corev1.ConditionFalse,
	}

	transition := !currentPodConditionFound || desiredPodCondition.Status != currentPodCondition.Status
	if transition {
		desiredPodCondition.LastTransitionTime = metav1.Now()
	}

	return desiredPodCondition, transition
}
