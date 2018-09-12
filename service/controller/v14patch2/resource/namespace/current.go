package namespace

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/finalizerskeptcontext"
	"github.com/giantswarm/operatorkit/controller/context/reconciliationcanceledcontext"
	"github.com/giantswarm/operatorkit/controller/context/resourcecanceledcontext"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/kvm-operator/service/controller/v14patch1/key"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "looking for the namespace in the Kubernetes API")

	// Lookup the current state of the namespace.
	var namespace *corev1.Namespace
	{
		manifest, err := r.k8sClient.CoreV1().Namespaces().Get(key.ClusterNamespace(customObject), metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "did not find the namespace in the Kubernetes API")
			// fall through
		} else if err != nil {
			return nil, microerror.Mask(err)
		} else {
			r.logger.LogCtx(ctx, "level", "debug", "message", "found the namespace in the Kubernetes API")
			namespace = manifest
		}
	}

	// In case the namespace is already terminating we do not need to do any
	// further work. Then we cancel the reconciliation to prevent the current and
	// any further resource from being processed.
	if namespace != nil && namespace.Status.Phase == "Terminating" {
		r.logger.LogCtx(ctx, "level", "debug", "message", "namespace is in state 'Terminating'")
		reconciliationcanceledcontext.SetCanceled(ctx)
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling reconciliation")

		return nil, nil
	}

	// In case a cluster deletion happens, we want to delete the guest cluster
	// namespace. We still need to use the namespace for resource creation in
	// order to drain nodes on KVM though. So as long as pods are there we delay
	// the deletion of the namespace here in order to still be able to create
	// resources in it. As soon as the draining was done and the pods got removed
	// we get an empty list here after the delete event got replayed. Then we just
	// remove the namespace as usual.
	if key.IsDeleted(customObject) {
		n := key.ClusterNamespace(customObject)
		list, err := r.k8sClient.CoreV1().Pods(n).List(metav1.ListOptions{})
		if err != nil {
			return nil, microerror.Mask(err)
		}
		if len(list.Items) != 0 {
			r.logger.LogCtx(ctx, "level", "debug", "message", "cannot finish deletion of namespace due to existing pods")
			resourcecanceledcontext.SetCanceled(ctx)
			finalizerskeptcontext.SetKept(ctx)
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

			return nil, nil
		}
	}

	return namespace, nil
}
