package pod

import (
	"context"
	"time"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/framework"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiv1 "k8s.io/client-go/pkg/api/v1"

	"github.com/giantswarm/kvm-operator/service/key"
)

const (
	PodDeletionGracePeriod = 0 * time.Second
)

func (r *Resource) ApplyDeleteChange(ctx context.Context, obj, deleteChange interface{}) error {
	podToDelete, err := key.ToPod(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	if podToDelete != nil {
		{
			r.logger.Log("cluster", key.ClusterIDFromPod(podToDelete), "debug", "updating the pod in the Kubernetes API to remove the pod's 'draining-nodes' finalizer")

			_, err := r.k8sClient.CoreV1().Pods(podToDelete.Namespace).Update(podToDelete)
			if err != nil {
				return microerror.Mask(err)
			}

			r.logger.Log("cluster", key.ClusterIDFromPod(podToDelete), "debug", "updated the pod in the Kubernetes API to remove the pod's 'draining-nodes' finalizer")
		}

		{
			r.logger.Log("cluster", key.ClusterIDFromPod(podToDelete), "debug", "deleting the pod in the Kubernetes API")

			gracePeriodSeconds := int64(PodDeletionGracePeriod.Seconds())
			options := &apismetav1.DeleteOptions{
				GracePeriodSeconds: &gracePeriodSeconds,
			}
			err = r.k8sClient.CoreV1().Pods(podToDelete.Namespace).Delete(podToDelete.Name, options)
			if err != nil {
				return microerror.Mask(err)
			}

			r.logger.Log("cluster", key.ClusterIDFromPod(podToDelete), "debug", "deleted the pod in the Kubernetes API")
		}
	} else {
		r.logger.Log("cluster", key.ClusterIDFromPod(podToDelete), "debug", "the pod does not need to be updated nor to be deleted in the Kubernetes API")
	}

	return nil
}

func (r *Resource) NewDeletePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*framework.Patch, error) {
	delete, err := r.newDeleteChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := framework.NewPatch()
	patch.SetDeleteChange(delete)

	return patch, nil
}

func (r *Resource) newDeleteChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	currentPod, err := key.ToPod(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// TODO go ahead without doing anything in case the pod is already terminated.

	// TODO drain guest cluster node and only remove the finalizer as soon as the
	// guest cluster node associated with the reconciled host cluster pod is
	// drained.

	// TODO go ahead and do not block once the draining is initialized. Using
	// finalizers on the pods the delete event will be replayed and we can check
	// if the draining completed on the next reconciliation loop.

	// TODO remove sleep. We just simlulate waiting for draining here.
	time.Sleep(15 * time.Second)

	// Here we remove the 'draining-nodes' finalizer from the reconciled pod, if
	// any. This frees the garbage collection lock in the Kubernetes API and makes
	// the pod vanish.
	var podToDelete *apiv1.Pod
	{
		var newFinalizers []string
		var changed bool

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

	return podToDelete, nil
}
