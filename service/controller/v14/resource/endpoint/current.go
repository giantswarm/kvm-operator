package endpoint

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/resourcecanceledcontext"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	pod, err := toPod(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var serviceName string
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding annotations")

		_, serviceName, err = getAnnotations(*pod, IPAnnotation, ServiceAnnotation)
		if IsMissingAnnotationError(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "did not find annotations")
			resourcecanceledcontext.SetCanceled(ctx)
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

			return nil, nil
		} else if err != nil {
			return nil, microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "found annotations")
	}

	var endpoint *Endpoint
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding endpoint")

		endpoint = &Endpoint{
			IPs:              []string{},
			ServiceName:      serviceName,
			ServiceNamespace: pod.GetNamespace(),
		}

		k8sEndpoints, err := r.k8sClient.CoreV1().Endpoints(pod.GetNamespace()).Get(serviceName, metav1.GetOptions{})
		if errors.IsNotFound(err) {
			// In case the endpoint manifest cannot be found in the Kubernetes API we
			// return the endpoint structure we dispatch without filling any IP.
			r.logger.LogCtx(ctx, "level", "debug", "message", "did not find endpoint")

			return endpoint, nil
		} else if err != nil {
			return nil, microerror.Mask(err)
		}

		for _, endpointSubset := range k8sEndpoints.Subsets {
			for _, endpointAddress := range endpointSubset.Addresses {
				foundIP := endpointAddress.IP

				if !containsIP(endpoint.IPs, foundIP) {
					endpoint.IPs = append(endpoint.IPs, foundIP)
				}
			}
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "found endpoint")
	}

	return endpoint, nil
}
