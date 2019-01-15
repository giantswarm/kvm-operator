package endpoint

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
)

func (r *Resource) ApplyCreateChange(ctx context.Context, obj, createState interface{}) error {
	endpointToCreate, err := toK8sEndpoint(createState)
	if err != nil {
		return microerror.Mask(err)
	}

	if !isEmptyEndpoint(endpointToCreate) {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("creating endpoint '%s'", endpointToCreate.GetName()))

		_, err = r.k8sClient.CoreV1().Endpoints(endpointToCreate.Namespace).Create(endpointToCreate)
		if errors.IsAlreadyExists(err) {
			// fall through
		} else if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("created endpoint '%s'", endpointToCreate.GetName()))
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("not creating endpoint '%s'", endpointToCreate.GetName()))
	}

	return nil
}

func (r *Resource) newCreateChange(ctx context.Context, obj, currentState, desiredState interface{}) (*corev1.Endpoints, error) {
	currentEndpoint, err := toEndpoint(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredEndpoint, err := toEndpoint(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var ips []string
	{
		var l []string
		for _, ip := range desiredEndpoint.IPs {
			if !containsIP(l, ip) {
				l = append(l, ip)
			}
		}
		if len(currentEndpoint.IPs) == 0 {
			ips = l
		}
	}

	var createChange *corev1.Endpoints
	{
		e := &Endpoint{
			Addresses:        ipsToAddresses(ips),
			IPs:              ips,
			Ports:            currentEndpoint.Ports,
			ServiceName:      currentEndpoint.ServiceName,
			ServiceNamespace: currentEndpoint.ServiceNamespace,
		}

		createChange = r.newK8sEndpoint(e)
	}

	return createChange, nil
}
