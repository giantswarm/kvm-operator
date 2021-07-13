package service

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/v5/pkg/controller/context/finalizerskeptcontext"
	"github.com/giantswarm/operatorkit/v5/pkg/controller/context/resourcecanceledcontext"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/kvm-operator/v4/service/controller/key"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.Debugf(ctx, "looking for services in the Kubernetes API")

	var services []*corev1.Service

	namespace := key.ClusterNamespace(customObject)
	serviceNames := []string{
		key.MasterID,
		key.WorkerID,
	}

	for _, name := range serviceNames {
		var manifest corev1.Service
		err := r.ctrlClient.Get(ctx, client.ObjectKey{
			Name:      name,
			Namespace: namespace,
		}, &manifest)
		if apierrors.IsNotFound(err) {
			r.logger.Debugf(ctx, "did not find a service in the Kubernetes API")
			// fall through
		} else if err != nil {
			return nil, microerror.Mask(err)
		} else {
			r.logger.Debugf(ctx, "found a service in the Kubernetes API")
			services = append(services, &manifest)
		}
	}

	r.logger.Debugf(ctx, "found %d services in the Kubernetes API", len(services))

	// In case a cluster deletion happens, we want to delete the workload cluster
	// services. We still need to use the services for ingress routing in order to
	// drain nodes on KVM though. So as long as pods are there we delay the
	// deletion of the services here in order to still be able to route traffic to
	// the workload cluster API. As soon as the draining was done and the pods got
	// removed we get an empty list here after the delete event got replayed. Then
	// we just remove the services as usual.
	if key.IsDeleted(&customObject) {
		var list corev1.PodList
		err := r.ctrlClient.List(ctx, &list, &client.ListOptions{
			Namespace: key.ClusterNamespace(customObject),
		})
		if err != nil {
			return nil, microerror.Mask(err)
		}
		if len(list.Items) != 0 {
			r.logger.Debugf(ctx, "cannot finish deletion of services due to existing pods")
			resourcecanceledcontext.SetCanceled(ctx)
			finalizerskeptcontext.SetKept(ctx)
			r.logger.Debugf(ctx, "canceling resource")

			return nil, nil
		}
	}

	return services, nil
}
