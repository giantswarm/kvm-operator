package ingress

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/v4/pkg/controller/context/finalizerskeptcontext"
	"github.com/giantswarm/operatorkit/v4/pkg/controller/context/resourcecanceledcontext"
	v1 "k8s.io/api/core/v1"
	"k8s.io/api/networking/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/kvm-operator/service/controller/key"
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
		var manifest v1beta1.Ingress
		err := r.ctrlClient.Get(ctx, client.ObjectKey{
			Namespace: namespace,
			Name:      name,
		}, &manifest)
		if apierrors.IsNotFound(err) {
			r.logger.Debugf(ctx, "did not find a ingress in the Kubernetes API")
			// fall through
		} else if err != nil {
			return nil, microerror.Mask(err)
		} else {
			r.logger.Debugf(ctx, "found a ingress in the Kubernetes API")
			ingresses = append(ingresses, &manifest)
		}
	}

	r.logger.Debugf(ctx, "found %d ingresses in the Kubernetes API", len(ingresses))

	// In case a cluster deletion happens, we want to delete the tenant cluster
	// ingresses. We still need to use the ingresses for ingress routing in order
	// to drain nodes on KVM though. So as long as pods are there we delay the
	// deletion of the ingresses here in order to still be able to route traffic
	// to the tenant cluster API. As soon as the draining was done and the pods
	// got removed we get an empty list here after the delete event got replayed.
	// Then we just remove the ingresses as usual.
	if key.IsDeleted(customObject) {
		var list v1.PodList
		err := r.ctrlClient.List(ctx, &list, &client.ListOptions{
			Namespace: key.ClusterNamespace(customObject),
		})
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
