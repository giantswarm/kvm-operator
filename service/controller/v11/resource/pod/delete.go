package pod

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/framework/context/reconciliationcanceledcontext"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/kvm-operator/service/controller/v11/key"
)

func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	reconciledPod, err := key.ToPod(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "debug", "looking for the current version of the reconciled pod in the Kubernetes API")

	var currentPod *corev1.Pod
	{
		currentPod, err = r.k8sClient.CoreV1().Pods(reconciledPod.GetNamespace()).Get(reconciledPod.Name, metav1.GetOptions{})
		if errors.IsNotFound(err) {
			// In case we reconcile a pod we cannot find anymore this means the
			// informer's watch event is outdated and the pod got already deleted in the
			// Kubernetes API. This is a normal transition behaviour, so we just ignore
			// it and assume we are done.
			r.logger.LogCtx(ctx, "debug", "cannot find the current version of the reconciled pod in the Kubernetes API")
			reconciliationcanceledcontext.SetCanceled(ctx)
			r.logger.LogCtx(ctx, "debug", "canceling reconciliation for pod")

			return nil
		} else if err != nil {
			return microerror.Mask(err)
		}
	}

	r.logger.LogCtx(ctx, "debug", "found the current version of the reconciled pod in the Kubernetes API")

	// TODO drain guest cluster node and only remove the finalizer as soon as the
	// guest cluster node associated with the reconciled host cluster pod is
	// drained.

	// TODO go ahead and do not block once the draining is initialized. Using
	// finalizers on the pods the delete event will be replayed and we can check
	// if the draining completed on the next reconciliation loop.

	// Here we remove the 'draining-nodes' finalizer from the reconciled pod, if
	// any. This frees the garbage collection lock in the Kubernetes API and makes
	// the pod vanish.
	var podToDelete *corev1.Pod
	{
		var changed bool
		var newFinalizers []string

		for _, f := range currentPod.GetFinalizers() {
			if f == key.DrainingNodesFinalizer {
				changed = true
				continue
			}

			newFinalizers = append(newFinalizers, f)
		}

		if changed {
			podToDelete = currentPod
			podToDelete.SetFinalizers(newFinalizers)
		}
	}

	r.logger.LogCtx(ctx, "debug", "looking if the pod has to be updated or to be deleted in the Kubernetes API")

	if podToDelete != nil {
		r.logger.LogCtx(ctx, "debug", "the pod has to be updated or to be deleted in the Kubernetes API")

		{
			r.logger.LogCtx(ctx, "debug", "updating the pod in the Kubernetes API to remove the pod's 'draining-nodes' finalizer")

			_, err := r.k8sClient.CoreV1().Pods(podToDelete.Namespace).Update(podToDelete)
			if errors.IsConflict(err) {
				// The reconciled pod may be updated by other processes or even humans
				// meanwhile. In case the resource version we currently know does not
				// match the latest existing one, we give up here and wait for the
				// delete event to be replayed. Then we try again later until we
				// succeed.
				r.logger.LogCtx(ctx, "debug", "cannot update the pod in the Kubernetes API to remove the pod's 'draining-nodes' finalizer because of outdated resource version")
				return nil
			} else if err != nil {
				return microerror.Mask(err)
			}

			r.logger.LogCtx(ctx, "debug", "updated the pod in the Kubernetes API to remove the pod's 'node-drainer' finalizer")
		}

		{
			r.logger.LogCtx(ctx, "debug", "deleting the pod in the Kubernetes API")

			gracePeriodSeconds := int64(0)
			options := &metav1.DeleteOptions{
				GracePeriodSeconds: &gracePeriodSeconds,
			}
			err = r.k8sClient.CoreV1().Pods(podToDelete.Namespace).Delete(podToDelete.Name, options)
			if err != nil {
				return microerror.Mask(err)
			}

			r.logger.LogCtx(ctx, "debug", "deleted the pod in the Kubernetes API")
		}
	} else {
		r.logger.LogCtx(ctx, "debug", "the pod does not need to be updated nor to be deleted in the Kubernetes API")
	}

	return nil
}
