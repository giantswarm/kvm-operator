package endpoint

import (
	"context"

	"github.com/giantswarm/apiextensions/v3/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/v5/pkg/controller/context/finalizerskeptcontext"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/kvm-operator/v4/service/controller/key"
)

func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	reconciledPod, err := key.ToPod(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	var currentPod corev1.Pod
	{
		r.logger.Debugf(ctx, "looking for the current version of the reconciled pod in the Kubernetes API")

		err = r.ctrlClient.Get(ctx, client.ObjectKey{
			Namespace: reconciledPod.Namespace,
			Name:      reconciledPod.Name,
		}, &currentPod)
		if apierrors.IsNotFound(err) {
			// In case we reconcile a pod we cannot find anymore this means the
			// informer's watch event is outdated and the pod got already deleted in
			// the Kubernetes API. This is a normal transition behaviour, so we just
			// ignore it and assume we are done.
			r.logger.Debugf(ctx, "cannot find the current version of the reconciled pod in the Kubernetes API")

			return nil
		} else if err != nil {
			return microerror.Mask(err)
		}
		r.logger.Debugf(ctx, "found the current version of the reconciled pod in the Kubernetes API")
	}

	{
		clusterID, ok := currentPod.Labels[key.LabelCluster]
		if !ok {
			return microerror.Maskf(missingClusterLabelError, "pod is missing cluster label")
		}

		var kvmConfig v1alpha1.KVMConfig
		err = r.ctrlClient.Get(ctx, client.ObjectKey{
			Namespace: metav1.NamespaceDefault,
			Name:      clusterID,
		}, &kvmConfig)
		if err != nil {
			return microerror.Mask(err)
		}

		if key.IsDeleted(&kvmConfig) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "cluster is being deleted")
			// Endpoints deletion will be handled when the namespace is deleted
			return nil
		}
	}

	{
		isDrained, err := key.IsPodDrained(currentPod)
		if err != nil {
			return microerror.Mask(err)
		}
		if !isDrained {
			r.logger.Debugf(ctx, "pod is not yet drained, not removing finalizer")
			finalizerskeptcontext.SetKept(ctx)

			return nil
		}
	}

	nodeIP, serviceName, err := r.podEndpointData(ctx, currentPod)
	if IsMissingAnnotation(err) {
		r.logger.Debugf(ctx, "missing annotations")
		return nil
	} else if err != nil {
		return microerror.Mask(err)
	}

	if serviceName.Name == key.MasterID {
		if key.AnyPodContainerRunning(currentPod) {
			r.logger.Debugf(ctx, "some pod containers are still running")
			finalizerskeptcontext.SetKept(ctx)

			return nil
		}
	}

	var endpoints corev1.Endpoints
	{
		r.logger.Debugf(ctx, "retrieving endpoints %#q", serviceName)
		err := r.ctrlClient.Get(ctx, serviceName, &endpoints)
		if apierrors.IsNotFound(err) {
			r.logger.Debugf(ctx, "endpoints not found")
			// There's nothing to be updated/removed so we can return early and drop the finalizer
			return nil
		} else if err != nil {
			r.logger.Debugf(ctx, "error retrieving endpoints")
			return microerror.Mask(err)
		}
	}

	var needUpdate bool
	{
		addresses, updated := removeFromAddresses(endpoints.Subsets[0].Addresses, nodeIP)
		if updated {
			needUpdate = true
			endpoints.Subsets[0].Addresses = addresses
		}
	}

	{
		addresses, updated := removeFromAddresses(endpoints.Subsets[0].NotReadyAddresses, nodeIP)
		if updated {
			needUpdate = true
			endpoints.Subsets[0].NotReadyAddresses = addresses
		}
	}

	if needUpdate && len(endpoints.Subsets[0].NotReadyAddresses) == 0 && len(endpoints.Subsets[0].Addresses) == 0 {
		err := r.ctrlClient.Delete(ctx, &endpoints)
		if err != nil {
			return microerror.Mask(err)
		}
	} else if needUpdate {
		err := r.ctrlClient.Update(ctx, &endpoints)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}
