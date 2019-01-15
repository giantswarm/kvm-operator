package endpoint

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/resourcecanceledcontext"
	corev1 "k8s.io/api/core/v1"
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

	var service *corev1.Service
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding service")

		service, err = r.k8sClient.CoreV1().Services(pod.GetNamespace()).Get(serviceName, metav1.GetOptions{})
		if errors.IsNotFound(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "did not find service")
		} else if err != nil {
			return nil, microerror.Mask(err)
		} else {
			r.logger.LogCtx(ctx, "level", "debug", "message", "found service")
		}
	}

	var endpoint *Endpoint
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding endpoint")

		endpoint = &Endpoint{
			Ports:            serviceToPorts(service),
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
				if !containsIP(endpoint.IPs, endpointAddress.IP) {
					endpoint.IPs = append(endpoint.IPs, endpointAddress.IP)
				}
			}
		}

		endpoint.Addresses = ipsToAddresses(endpoint.IPs)

		r.logger.LogCtx(ctx, "level", "debug", "message", "found endpoint")
	}

	return endpoint, nil
}

func serviceToPorts(s *corev1.Service) []corev1.EndpointPort {
	if s == nil {
		return nil
	}

	var ports []corev1.EndpointPort

	for _, p := range s.Spec.Ports {
		port := corev1.EndpointPort{
			Name: p.Name,
			Port: p.Port,
		}

		ports = append(ports, port)
	}

	return ports
}
