package pod

import (
	"context"
	"time"

	corev1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/finalizerskeptcontext"
	"github.com/giantswarm/operatorkit/controller/context/resourcecanceledcontext"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/kvm-operator/service/controller/v13/key"
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
		if apierrors.IsNotFound(err) {
			// In case we reconcile a pod we cannot find anymore this means the
			// informer's watch event is outdated and the pod got already deleted in
			// the Kubernetes API. This is a normal transition behaviour, so we just
			// ignore it and assume we are done.
			r.logger.LogCtx(ctx, "debug", "cannot find the current version of the reconciled pod in the Kubernetes API")
			resourcecanceledcontext.SetCanceled(ctx)
			r.logger.LogCtx(ctx, "debug", "canceling reconciliation for pod")

			return nil
		} else if err != nil {
			return microerror.Mask(err)
		}
	}

	{
		isDrained, err := key.IsPodDrained(currentPod)
		if err != nil {
			return microerror.Mask(err)
		}
		if isDrained {
			r.logger.LogCtx(ctx, "level", "debug", "message", "pod is already drained")
			resourcecanceledcontext.SetCanceled(ctx)
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource for custom object")

			return nil
		}
	}

	if !forcePodCleanup(currentPod) {
		r.logger.LogCtx(ctx, "debug", "found the current version of the reconciled pod in the Kubernetes API")

		n := currentPod.GetNamespace()
		p := currentPod.GetName()
		o := metav1.GetOptions{}

		nodeConfig, err := r.g8sClient.CoreV1alpha1().NodeConfigs(n).Get(p, o)
		if apierrors.IsNotFound(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "did not find node config for guest cluster node")

			err := r.createNodeConfig(ctx, currentPod)
			if err != nil {
				return microerror.Mask(err)
			}

			resourcecanceledcontext.SetCanceled(ctx)
			finalizerskeptcontext.SetKept(ctx)
			r.logger.LogCtx(ctx, "debug", "canceling reconciliation for pod")

		} else if err != nil {
			return microerror.Mask(err)
		} else {
			r.logger.LogCtx(ctx, "level", "debug", "message", "found node config for the guest cluster")

			r.logger.LogCtx(ctx, "level", "debug", "message", "waiting for inspection of the reconciled pod")
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "inspecting node config for the guest cluster")

		if !nodeConfig.Status.HasFinalCondition() {
			r.logger.LogCtx(ctx, "level", "debug", "message", "node config of guest cluster has no final state")
			resourcecanceledcontext.SetCanceled(ctx)
			finalizerskeptcontext.SetKept(ctx)
			r.logger.LogCtx(ctx, "debug", "canceling reconciliation for pod")

			return nil
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "node config of guest cluster has final state")
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "deleting node config for guest cluster node")

		n := currentPod.GetNamespace()
		i := currentPod.GetName()
		o := &metav1.DeleteOptions{}

		err := r.g8sClient.CoreV1alpha1().NodeConfigs(n).Delete(i, o)
		if apierrors.IsNotFound(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "node config for guest cluster node already deleted")
		} else if err != nil {
			return microerror.Mask(err)
		} else {
			r.logger.LogCtx(ctx, "level", "debug", "message", "deleted node config for guest cluster node")
		}
	}

	var podToDelete *corev1.Pod
	{
		podToDelete = currentPod

		a := podToDelete.GetAnnotations()
		a[key.AnnotationPodDrained] = "True"
		podToDelete.SetAnnotations(a)
	}

	{
		r.logger.LogCtx(ctx, "debug", "updating the pod in the Kubernetes API")

		_, err := r.k8sClient.CoreV1().Pods(podToDelete.Namespace).Update(podToDelete)
		if apierrors.IsConflict(err) {
			// The reconciled pod may be updated by other processes or even humans
			// meanwhile. In case the resource version we currently know does not
			// match the latest existing one, we give up here and wait for the
			// delete event to be replayed. Then we try again later until we
			// succeed.
			r.logger.LogCtx(ctx, "debug", "cannot update the pod in the Kubernetes API due to outdated resource version")
			resourcecanceledcontext.SetCanceled(ctx)
			finalizerskeptcontext.SetKept(ctx)
			r.logger.LogCtx(ctx, "debug", "canceling reconciliation for pod")

			return nil
		} else if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "debug", "updated the pod in the Kubernetes API")
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

	return nil
}

func (r *Resource) createNodeConfig(ctx context.Context, pod *corev1.Pod) error {
	r.logger.LogCtx(ctx, "level", "debug", "message", "creating node config for guest cluster node")

	apiEndpoint, err := key.ClusterAPIEndpointFromPod(pod)
	if err != nil {
		return microerror.Mask(err)
	}

	n := pod.GetNamespace()
	c := &corev1alpha1.NodeConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: pod.GetName(),
		},
		Spec: corev1alpha1.NodeConfigSpec{
			Guest: corev1alpha1.NodeConfigSpecGuest{
				Cluster: corev1alpha1.NodeConfigSpecGuestCluster{
					API: corev1alpha1.NodeConfigSpecGuestClusterAPI{
						Endpoint: apiEndpoint,
					},
					ID: pod.GetNamespace(),
				},
				Node: corev1alpha1.NodeConfigSpecGuestNode{
					Name: pod.GetName(),
				},
			},
			VersionBundle: corev1alpha1.NodeConfigSpecVersionBundle{
				Version: "0.1.0",
			},
		},
	}

	_, err = r.g8sClient.CoreV1alpha1().NodeConfigs(n).Create(c)
	if apierrors.IsAlreadyExists(err) {
		r.logger.LogCtx(ctx, "level", "warning", "message", "node config for guest cluster node already exists")
	} else if err != nil {
		return microerror.Mask(err)
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "created node config for guest cluster node")
	}

	return nil
}

func forcePodCleanup(pod *corev1.Pod) bool {
	if !key.IsPodDeleted(pod) {
		return false
	}

	if pod.GetDeletionTimestamp().Add(30 * time.Minute).After(time.Now()) {
		return false
	}

	return true
}
