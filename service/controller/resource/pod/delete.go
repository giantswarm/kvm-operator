package pod

import (
	"context"

	corev1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/v4/pkg/controller/context/finalizerskeptcontext"
	"github.com/giantswarm/operatorkit/v4/pkg/controller/context/resourcecanceledcontext"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/kvm-operator/service/controller/key"
)

func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	reconciledPod, err := key.ToPod(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	var currentPod *corev1.Pod
	{
		r.logger.Debugf(ctx, "looking for the current version of the reconciled pod in the Kubernetes API")

		currentPod, err = r.k8sClient.CoreV1().Pods(reconciledPod.GetNamespace()).Get(ctx, reconciledPod.Name, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			// In case we reconcile a pod we cannot find anymore this means the
			// informer's watch event is outdated and the pod got already deleted in
			// the Kubernetes API. This is a normal transition behaviour, so we just
			// ignore it and assume we are done.
			r.logger.Debugf(ctx, "cannot find the current version of the reconciled pod in the Kubernetes API")
			resourcecanceledcontext.SetCanceled(ctx)
			r.logger.Debugf(ctx, "canceling reconciliation for pod")

			return nil
		} else if err != nil {
			return microerror.Mask(err)
		}
		r.logger.Debugf(ctx, "found the current version of the reconciled pod in the Kubernetes API")
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "checking if cluster is being deleted")

		clusterID, ok := currentPod.Labels[key.LabelCluster]
		if !ok {
			return microerror.Maskf(missingClusterLabelError, "pod is missing cluster label")
		}

		kvmConfig, err := r.g8sClient.ProviderV1alpha1().KVMConfigs(metav1.NamespaceDefault).Get(ctx, clusterID, metav1.GetOptions{})
		if err != nil {
			return microerror.Mask(err)
		}

		if key.IsDeleted(kvmConfig) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "cluster is being deleted")
			return nil
		}
	}

	{
		isDrained, err := key.IsPodDrained(*currentPod)
		if err != nil {
			return microerror.Mask(err)
		}
		if isDrained {
			r.logger.Debugf(ctx, "pod is already drained")
			return nil
		}
	}

	{
		r.logger.Debugf(ctx, "looking for the drainer config for the tenant cluster")

		n := currentPod.GetNamespace()
		p := currentPod.GetName()
		o := metav1.GetOptions{}

		drainerConfig, err := r.g8sClient.CoreV1alpha1().DrainerConfigs(n).Get(ctx, p, o)
		if apierrors.IsNotFound(err) {
			r.logger.Debugf(ctx, "did not find drainer config for tenant cluster node")

			err := r.createDrainerConfig(ctx, currentPod)
			if err != nil {
				return microerror.Mask(err)
			}

			finalizerskeptcontext.SetKept(ctx)
			r.logger.Debugf(ctx, "canceling reconciliation for pod")
			return nil

		} else if err != nil {
			return microerror.Mask(err)
		} else {
			r.logger.Debugf(ctx, "found drainer config for the tenant cluster")

			r.logger.Debugf(ctx, "waiting for inspection of the reconciled pod")
		}

		if drainerConfig.Status.HasDrainedCondition() {
			r.logger.Debugf(ctx, "drainer config of tenant cluster has drained condition")

			err := r.finishDraining(ctx, currentPod, drainerConfig)
			if err != nil {
				return microerror.Mask(err)
			}
		} else if drainerConfig.Status.HasTimeoutCondition() {
			r.logger.Debugf(ctx, "drainer config of tenant cluster has timeout condition")

			err := r.finishDraining(ctx, currentPod, drainerConfig)
			if err != nil {
				return microerror.Mask(err)
			}
		} else if key.PodContainersNotRunning(*currentPod) {
			r.logger.Debugf(ctx, "pod is treated as drained")
			r.logger.Debugf(ctx, "all pod's containers are not running")

			err := r.finishDraining(ctx, currentPod, drainerConfig)
			if err != nil {
				return microerror.Mask(err)
			}
		} else {
			r.logger.Debugf(ctx, "node termination is still in progress")
		}
	}

	r.logger.Debugf(ctx, "keeping finalizers")
	finalizerskeptcontext.SetKept(ctx)

	return nil
}

func (r *Resource) createDrainerConfig(ctx context.Context, pod *corev1.Pod) error {
	r.logger.Debugf(ctx, "creating drainer config for tenant cluster node")

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

	_, err = r.g8sClient.CoreV1alpha1().DrainerConfigs(n).Create(ctx, c, metav1.CreateOptions{})
	if apierrors.IsAlreadyExists(err) {
		r.logger.LogCtx(ctx, "level", "warning", "message", "drainer config for tenant cluster node already exists")
	} else if err != nil {
		return microerror.Mask(err)
	} else {
		r.logger.Debugf(ctx, "created drainer config for tenant cluster node")
	}

	return nil
}

func (r *Resource) finishDraining(ctx context.Context, currentPod *corev1.Pod, drainerConfig *corev1alpha1.DrainerConfig) error {
	var err error

	{
		r.logger.Debugf(ctx, "deleting drainer config for tenant cluster node")

		n := currentPod.GetNamespace()
		i := currentPod.GetName()
		o := metav1.DeleteOptions{}

		err := r.g8sClient.CoreV1alpha1().DrainerConfigs(n).Delete(ctx, i, o)
		if apierrors.IsNotFound(err) {
			r.logger.Debugf(ctx, "drainer config for tenant cluster node already deleted")
		} else if err != nil {
			return microerror.Mask(err)
		} else {
			r.logger.Debugf(ctx, "deleted drainer config for tenant cluster node")
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
		r.logger.Debugf(ctx, "updating the pod in the Kubernetes API")

		_, err := r.k8sClient.CoreV1().Pods(podToDelete.Namespace).Update(ctx, podToDelete, metav1.UpdateOptions{})
		if apierrors.IsConflict(err) {
			// The reconciled pod may be updated by other processes or even humans
			// meanwhile. In case the resource version we currently know does not
			// match the latest existing one, we give up here and wait for the
			// delete event to be replayed. Then we try again later until we
			// succeed.
			r.logger.Debugf(ctx, "cannot update the pod in the Kubernetes API due to outdated resource version")
			resourcecanceledcontext.SetCanceled(ctx)
			finalizerskeptcontext.SetKept(ctx)
			r.logger.Debugf(ctx, "canceling reconciliation")

			return nil
		} else if err != nil {
			return microerror.Mask(err)
		}

		r.logger.Debugf(ctx, "updated the pod in the Kubernetes API")
	}

	{
		r.logger.Debugf(ctx, "deleting the pod in the Kubernetes API")

		gracePeriodSeconds := int64(0)
		options := metav1.DeleteOptions{
			GracePeriodSeconds: &gracePeriodSeconds,
		}
		err = r.k8sClient.CoreV1().Pods(podToDelete.Namespace).Delete(ctx, podToDelete.Name, options)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.Debugf(ctx, "deleted the pod in the Kubernetes API")
	}

	return nil
}
