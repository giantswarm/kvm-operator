package endpoint

import (
	"context"

	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"

<<<<<<< HEAD
<<<<<<< HEAD
	"github.com/giantswarm/kvm-operator/service/controller/v24patch1/key"
=======
	"github.com/giantswarm/kvm-operator/service/controller/v24/key"
>>>>>>> c4c6c79d... copy v24 to v24patch1
=======
	"github.com/giantswarm/kvm-operator/service/controller/v24patch1/key"
>>>>>>> d6f149c2... wire v24patch1
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	pod, err := toPod(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	endpointIP, serviceName, err := getAnnotations(*pod, IPAnnotation, ServiceAnnotation)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredEndpoint := &Endpoint{
		IPs: []string{
			endpointIP,
		},
		ServiceName:      serviceName,
		ServiceNamespace: pod.GetNamespace(),
	}

	podIsReady := false
	for _, condition := range pod.Status.Conditions {
		if condition.Type == corev1.PodReady && condition.Status == corev1.ConditionTrue {
			podIsReady = true
			break
		}
	}

	// If the pod is not ready we should not add the ip to the endpoint list.
	if !podIsReady && serviceName == key.WorkerID {
		desiredEndpoint.IPs = []string{}
	}

	return desiredEndpoint, nil
}
