package pod

import (
	"context"

	corev1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/finalizerskeptcontext"
	"github.com/giantswarm/operatorkit/controller/context/resourcecanceledcontext"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/kvm-operator/service/controller/v25/key"
)

func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	reconciledPod, err := key.ToPod(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	var currentPod *corev1.Pod
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "looking for the current version of the reconciled pod in the Kubernetes API")

		currentPod, err = r.k8sClient.CoreV1().Pods(reconciledPod.GetNamespace()).Get(reconciledPod.Name, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			// In case we reconcile a pod we cannot find anymore this means the
			// informer's watch event is outdated and the pod got already deleted in
			// the Kubernetes API. This is a normal transition behaviour, so we just
			// ignore it and assume we are done.
			r.logger.LogCtx(ctx, "level", "debug", "message", "cannot find the current version of the reconciled pod in the Kubernetes API")
			resourcecanceledcontext.SetCanceled(ctx)
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling reconciliation for pod")

			return nil
		} else if err != nil {
			return microerror.Mask(err)
		}
		r.logger.LogCtx(ctx, "level", "debug", "message", "found the current version of the reconciled pod in the Kubernetes API")
	}

	{
		isDrained, err := key.IsPodDrained(currentPod)
		if err != nil {
			return microerror.Mask(err)
		}
		if isDrained {
			r.logger.LogCtx(ctx, "level", "debug", "message", "pod is already drained")
			resourcecanceledcontext.SetCanceled(ctx)
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

			return nil
		}
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "looking for the drainer config for the tenant cluster")

		n := currentPod.GetNamespace()
		p := currentPod.GetName()
		o := metav1.GetOptions{}

		drainerConfig, err := r.g8sClient.CoreV1alpha1().DrainerConfigs(n).Get(p, o)
		if apierrors.IsNotFound(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "did not find drainer config for tenant cluster node")

			err := r.createDrainerConfig(ctx, currentPod)
			if err != nil {
				return microerror.Mask(err)
			}

			resourcecanceledcontext.SetCanceled(ctx)
			finalizerskeptcontext.SetKept(ctx)
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling reconciliation for pod")
			return nil

		} else if err != nil {
			return microerror.Mask(err)
		} else {
			r.logger.LogCtx(ctx, "level", "debug", "message", "found drainer config for the tenant cluster")

			r.logger.LogCtx(ctx, "level", "debug", "message", "waiting for inspection of the reconciled pod")
		}

		if drainerConfig.Status.HasDrainedCondition() {
			r.logger.LogCtx(ctx, "level", "debug", "message", "drainer config of tenant cluster has drained condition")

			err := r.finishDraining(ctx, currentPod, drainerConfig)
			if err != nil {
				return microerror.Mask(err)
			}
		} else if drainerConfig.Status.HasTimeoutCondition() {
			r.logger.LogCtx(ctx, "level", "debug", "message", "drainer config of tenant cluster has timeout condition")

			err := r.finishDraining(ctx, currentPod, drainerConfig)
			if err != nil {
				return microerror.Mask(err)
			}
		} else if key.ArePodContainersTerminated(currentPod) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "pod is treated as drained")
			r.logger.LogCtx(ctx, "level", "debug", "message", "all pod's containers are terminated")

			err := r.finishDraining(ctx, currentPod, drainerConfig)
			if err != nil {
				return microerror.Mask(err)
			}
		} else {
			r.logger.LogCtx(ctx, "level", "debug", "message", "node termination is still in progress")
		}
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "canceling reconciliation")
	resourcecanceledcontext.SetCanceled(ctx)

	r.logger.LogCtx(ctx, "level", "debug", "message", "keeping finalizers")
	finalizerskeptcontext.SetKept(ctx)

	return nil
}

func (r *Resource) createDrainerConfig(ctx context.Context, pod *corev1.Pod) error {
	r.logger.LogCtx(ctx, "level", "debug", "message", "creating drainer config for tenant cluster node")

	apiEndpoint, err := key.ClusterAPIEndpointFromPod(pod)
	if err != nil {
		return microerror.Mask(err)
	}

	n := pod.GetNamespace()
	c := &corev1alpha1.DrainerConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: pod.GetName(),
		},
		Spec: corev1alpha1.DrainerConfigSpec{
			Guest: corev1alpha1.DrainerConfigSpecGuest{
				Cluster: corev1alpha1.DrainerConfigSpecGuestCluster{
					API: corev1alpha1.DrainerConfigSpecGuestClusterAPI{
						Endpoint: apiEndpoint,
					},
					ID: pod.GetNamespace(),
				},
				Node: corev1alpha1.DrainerConfigSpecGuestNode{
					Name: pod.GetName(),
				},
			},
			VersionBundle: corev1alpha1.DrainerConfigSpecVersionBundle{
				Version: "0.2.0",
			},
		},
	}

	_, err = r.g8sClient.CoreV1alpha1().DrainerConfigs(n).Create(c)
	if apierrors.IsAlreadyExists(err) {
		r.logger.LogCtx(ctx, "level", "warning", "message", "drainer config for tenant cluster node already exists")
	} else if err != nil {
		return microerror.Mask(err)
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "created drainer config for tenant cluster node")
	}

	return nil
}

func (r *Resource) finishDraining(ctx context.Context, currentPod *corev1.Pod, drainerConfig *corev1alpha1.DrainerConfig) error {
	var err error

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "deleting drainer config for tenant cluster node")

		n := currentPod.GetNamespace()
		i := currentPod.GetName()
		o := &metav1.DeleteOptions{}

		err := r.g8sClient.CoreV1alpha1().DrainerConfigs(n).Delete(i, o)
		if apierrors.IsNotFound(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "drainer config for tenant cluster node already deleted")
		} else if err != nil {
			return microerror.Mask(err)
		} else {
			r.logger.LogCtx(ctx, "level", "debug", "message", "deleted drainer config for tenant cluster node")
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
		r.logger.LogCtx(ctx, "level", "debug", "message", "updating the pod in the Kubernetes API")

		_, err := r.k8sClient.CoreV1().Pods(podToDelete.Namespace).Update(podToDelete)
		if apierrors.IsConflict(err) {
			// The reconciled pod may be updated by other processes or even humans
			// meanwhile. In case the resource version we currently know does not
			// match the latest existing one, we give up here and wait for the
			// delete event to be replayed. Then we try again later until we
			// succeed.
			r.logger.LogCtx(ctx, "level", "debug", "message", "cannot update the pod in the Kubernetes API due to outdated resource version")
			resourcecanceledcontext.SetCanceled(ctx)
			finalizerskeptcontext.SetKept(ctx)
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling reconciliation")

			return nil
		} else if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "updated the pod in the Kubernetes API")
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "deleting the pod in the Kubernetes API")

		gracePeriodSeconds := int64(0)
		options := &metav1.DeleteOptions{
			GracePeriodSeconds: &gracePeriodSeconds,
		}
		err = r.k8sClient.CoreV1().Pods(podToDelete.Namespace).Delete(podToDelete.Name, options)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "deleted the pod in the Kubernetes API")
	}

	return nil
}
