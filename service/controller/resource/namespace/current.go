package namespace

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/v5/pkg/controller/context/finalizerskeptcontext"
	"github.com/giantswarm/operatorkit/v5/pkg/controller/context/resourcecanceledcontext"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/kvm-operator/v4/pkg/label"
	"github.com/giantswarm/kvm-operator/v4/service/controller/key"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var namespace *corev1.Namespace
	{
		r.logger.Debugf(ctx, "finding the namespace in the Kubernetes API")

		var retrieved corev1.Namespace
		err := r.ctrlClient.Get(ctx, client.ObjectKey{Name: key.ClusterNamespace(customObject)}, &retrieved)
		if apierrors.IsNotFound(err) {
			r.logger.Debugf(ctx, "did not find the namespace in the Kubernetes API")
			// fall through
		} else if err != nil {
			return nil, microerror.Mask(err)
		} else {
			r.logger.Debugf(ctx, "found the namespace in the Kubernetes API")
			namespace = &retrieved
		}
	}

	// In case the namespace is already terminating we do not need to do any
	// further work. We cancel the namespace resource to prevent any further work,
	// but keep the finalizers until the namespace got finally deleted. This is to
	// prevent issues with the monitoring and alerting systems. The cluster status
	// conditions of the watched CR are used to inhibit alerts, for instance when
	// the cluster is being deleted.
	if namespace != nil && namespace.Status.Phase == corev1.NamespaceTerminating {
		r.logger.Debugf(ctx, "namespace is %#q", corev1.NamespaceTerminating)

		r.logger.Debugf(ctx, "keeping finalizers")
		finalizerskeptcontext.SetKept(ctx)

		r.logger.Debugf(ctx, "canceling resource")
		resourcecanceledcontext.SetCanceled(ctx)

		return nil, nil
	}

	if namespace == nil && key.IsDeleted(&customObject) {
		r.logger.Debugf(ctx, "resource deletion completed")

		r.logger.Debugf(ctx, "canceling resource")
		resourcecanceledcontext.SetCanceled(ctx)

		return nil, nil
	}

	// In case a cluster deletion happens, we want to delete the workload cluster
	// namespace. We still need to use the namespace for resource creation in
	// order to drain nodes on KVM though. So as long as pods are there we delay
	// the deletion of the namespace here in order to still be able to create
	// resources in it. As soon as the draining was done and the pods got removed
	// we get an empty list here after the delete event got replayed. Then we just
	// remove the namespace as usual.
	if key.IsDeleted(&customObject) {
		var list corev1.PodList
		err := r.ctrlClient.List(ctx, &list, &client.ListOptions{
			Namespace: key.ClusterNamespace(customObject),
			LabelSelector: labels.SelectorFromSet(map[string]string{
				label.ManagedBy: key.OperatorName,
			}),
		})
		if err != nil {
			return nil, microerror.Mask(err)
		}

		var deployments v1.DeploymentList
		err = r.ctrlClient.List(ctx, &deployments, &client.ListOptions{
			LabelSelector: labels.SelectorFromSet(map[string]string{
				label.ManagedBy: key.OperatorName,
			}),
			Namespace: key.ClusterNamespace(customObject),
		})
		if err != nil {
			return nil, microerror.Mask(err)
		}

		// Ensure PVCs have been deleted so that bound PVs are properly cleaned up.
		var volumeClaims corev1.PersistentVolumeClaimList
		err = r.ctrlClient.List(ctx, &volumeClaims, &client.ListOptions{
			LabelSelector: labels.SelectorFromSet(map[string]string{
				label.ManagedBy: key.OperatorName,
			}),
			Namespace: key.ClusterNamespace(customObject),
		})
		if err != nil {
			return nil, microerror.Mask(err)
		}

		if len(deployments.Items) != 0 || len(volumeClaims.Items) != 0 {
			if len(deployments.Items) != 0 {
				r.logger.Debugf(ctx, "cannot finish deletion of namespace due to existing deployments")
			} else {
				r.logger.Debugf(ctx, "cannot finish deletion of namespace due to existing PVCs")
			}

			resourcecanceledcontext.SetCanceled(ctx)
			finalizerskeptcontext.SetKept(ctx)
			r.logger.Debugf(ctx, "canceling resource")

			return nil, nil
		}
	}

	return namespace, nil
}
