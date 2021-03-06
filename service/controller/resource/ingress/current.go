package ingress

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/v5/pkg/controller/context/finalizerskeptcontext"
	"github.com/giantswarm/operatorkit/v5/pkg/controller/context/resourcecanceledcontext"
	"k8s.io/api/networking/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/kvm-operator/v4/service/controller/key"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.Debugf(ctx, "looking for ingresses in the Kubernetes API")

	var ingresses []*v1beta1.Ingress

	namespace := key.ClusterNamespace(customObject)
	ingressNames := []string{
		APIID,
		EtcdID,
	}

	for _, name := range ingressNames {
		manifest, err := r.k8sClient.NetworkingV1beta1().Ingresses(namespace).Get(ctx, name, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			r.logger.Debugf(ctx, "did not find a ingress in the Kubernetes API")
			// fall through
		} else if err != nil {
			return nil, microerror.Mask(err)
		} else {
			r.logger.Debugf(ctx, "found a ingress in the Kubernetes API")
			ingresses = append(ingresses, manifest)
		}
	}

	r.logger.Debugf(ctx, "found %d ingresses in the Kubernetes API", len(ingresses))

	// In case a cluster deletion happens, we want to delete the workload cluster
	// ingresses. We still need to use the ingresses for ingress routing in order
	// to drain nodes on KVM though. So as long as pods are there we delay the
	// deletion of the ingresses here in order to still be able to route traffic
	// to the workload cluster API. As soon as the draining was done and the pods
	// got removed we get an empty list here after the delete event got replayed.
	// Then we just remove the ingresses as usual.
	if key.IsDeleted(&customObject) {
		n := key.ClusterNamespace(customObject)
		list, err := r.k8sClient.CoreV1().Pods(n).List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, microerror.Mask(err)
		}
		if len(list.Items) != 0 {
			r.logger.Debugf(ctx, "cannot finish deletion of ingresses due to existing pods")
			resourcecanceledcontext.SetCanceled(ctx)
			finalizerskeptcontext.SetKept(ctx)
			r.logger.Debugf(ctx, "canceling resource")

			return nil, nil
		}
	}

	return ingresses, nil
}
