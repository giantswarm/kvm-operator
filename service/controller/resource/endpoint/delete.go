package endpoint

import (
	"context"

	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"

	"github.com/giantswarm/kvm-operator/service/controller/key"
)

func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	pod, err := key.ToPod(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	nodeIP, serviceName, err := r.podEndpointData(ctx, pod)
	if err != nil {
		return microerror.Mask(err)
	}

	var endpoints corev1.Endpoints
	{
		r.logger.Debugf(ctx, "retrieving endpoints %#q", serviceName)
		err := r.ctrlClient.Get(ctx, serviceName, &endpoints)
		if err != nil {
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
