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

	var createChange *corev1.Endpoints
	{
		ips := ipsForCreateChange(currentEndpoint.IPs, desiredEndpoint.IPs)

		e := &Endpoint{
			Addresses:        ipsToAddresses(ips),
			IPs:              ips,
			Ports:            currentEndpoint.Ports,
			ResourceVersion:  currentEndpoint.ResourceVersion,
			ServiceName:      currentEndpoint.ServiceName,
			ServiceNamespace: currentEndpoint.ServiceNamespace,
		}

		createChange = r.newK8sEndpoint(e)
	}

	return createChange, nil
}

func ipsForCreateChange(currentIPs []string, desiredIPs []string) []string {
	var ips []string

	for _, ip := range desiredIPs {
		if !containsIP(ips, ip) {
			ips = append(ips, ip)
		}
	}

	if len(currentIPs) == 0 {
		return ips
	}

	return nil
}
