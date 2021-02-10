package endpoint

import (
	"context"

	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/kvm-operator/service/controller/key"
)

func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	pod, err := key.ToPod(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	var nodeIP string
	var serviceName string
	{
		var ok bool
		nodeIP, ok = pod.Annotations[key.AnnotationIp]
		if !ok || nodeIP == "" {
			r.logger.Debugf(ctx, "node pod has no ip annotation %#q, skipping", key.AnnotationIp)
			return nil
		}

		serviceName, ok = pod.Annotations[key.AnnotationService]
		if !ok || serviceName == "" {
			r.logger.Debugf(ctx, "node pod has no service annotation %#q, skipping", key.AnnotationService)
			return nil
		} else if serviceName == key.MasterID {
			r.logger.Debugf(ctx, "node pod contains a workload cluster master node, skipping")
			return nil
		}
	}

	var endpoints *corev1.Endpoints
	{
		r.logger.Debugf(ctx, "retrieving endpoints %#q", serviceName)
		var err error
		endpoints, err = r.k8sClient.CoreV1().Endpoints(pod.Namespace).Get(ctx, serviceName, v1.GetOptions{})
		if err != nil {
			r.logger.Debugf(ctx, "error retrieving endpoints")
			return microerror.Mask(err)
		}
		r.logger.Debugf(ctx, "retrieved endpoints")
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
		err := r.k8sClient.CoreV1().Endpoints(endpoints.Namespace).Delete(ctx, endpoints.Name, v1.DeleteOptions{})
		if err != nil {
			return microerror.Mask(err)
		}
	} else if needUpdate {
		_, err := r.k8sClient.CoreV1().Endpoints(endpoints.Namespace).Update(ctx, endpoints, v1.UpdateOptions{})
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}
