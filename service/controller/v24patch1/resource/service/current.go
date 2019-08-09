package service

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/finalizerskeptcontext"
	"github.com/giantswarm/operatorkit/controller/context/resourcecanceledcontext"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/kvm-operator/service/controller/v24patch1/key"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "looking for services in the Kubernetes API")

	var services []*corev1.Service

	namespace := key.ClusterNamespace(customObject)
	serviceNames := []string{
		key.MasterID,
		key.WorkerID,
	}

	for _, name := range serviceNames {
		manifest, err := r.k8sClient.CoreV1().Services(namespace).Get(name, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "did not find a service in the Kubernetes API")
			// fall through
		} else if err != nil {
			return nil, microerror.Mask(err)
		} else {
			r.logger.LogCtx(ctx, "level", "debug", "message", "found a service in the Kubernetes API")
			services = append(services, manifest)
		}
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found %d services in the Kubernetes API", len(services)))

	// In case a cluster deletion happens, we want to delete the tenant cluster
	// services. We still need to use the services for ingress routing in order to
	// drain nodes on KVM though. So as long as pods are there we delay the
	// deletion of the services here in order to still be able to route traffic to
	// the tenant cluster API. As soon as the draining was done and the pods got
	// removed we get an empty list here after the delete event got replayed. Then
	// we just remove the services as usual.
	if key.IsDeleted(customObject) {
		n := key.ClusterNamespace(customObject)
		list, err := r.k8sClient.CoreV1().Pods(n).List(metav1.ListOptions{})
		if err != nil {
			return nil, microerror.Mask(err)
		}
		if len(list.Items) != 0 {
			r.logger.LogCtx(ctx, "level", "debug", "message", "cannot finish deletion of services due to existing pods")
			resourcecanceledcontext.SetCanceled(ctx)
			finalizerskeptcontext.SetKept(ctx)
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

			return nil, nil
		}
	}

	return services, nil
}
