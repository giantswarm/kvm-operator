package service

import (
	"reflect"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// Name is the identifier of the resource.
	Name = "service"
)

// Config represents the configuration used to create a new service resource.
type Config struct {
	// Dependencies.
	CtrlClient    client.Client
	Logger    micrologger.Logger
}

// Resource implements the service resource.
type Resource struct {
	// Dependencies.
	ctrlClient    client.Client
	logger    micrologger.Logger
}

// New creates a new configured service resource.
func New(config Config) (*Resource, error) {
	// Dependencies.
	if config.CtrlClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.CtrlClient must not be empty")
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}

	newService := &Resource{
		// Dependencies.
		ctrlClient: config.CtrlClient,
		logger:    config.Logger,
	}

	return newService, nil
}

func (r *Resource) Name() string {
	return Name
}

func containsService(list []*corev1.Service, item *corev1.Service) bool {
	for _, l := range list {
		if l.Name == item.Name {
			return true
		}
	}

	return false
}

func getServiceByName(list []*corev1.Service, name string) (*corev1.Service, error) {
	for _, l := range list {
		if l.Name == name {
			return l, nil
		}
	}

	return nil, microerror.Mask(notFoundError)
}

func isServiceModified(a, b *corev1.Service) bool {
	if a == nil || b == nil {
		return true
	}
	if !portsEqual(a, b) {
		return true
	}

	if !reflect.DeepEqual(a.Spec.Type, b.Spec.Type) {
		return true
	}

	if !reflect.DeepEqual(a.Labels, b.Labels) {
		return true
	}

	if !reflect.DeepEqual(a.Annotations, b.Annotations) {
		return true
	}

	return false
}

func toServices(v interface{}) ([]*corev1.Service, error) {
	if v == nil {
		return nil, nil
	}

	services, ok := v.([]*corev1.Service)
	if !ok {
		return nil, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", []*corev1.Service{}, v)
	}

	return services, nil
}

// portsEqual is a function that is checking if ports in the service have same important values.
func portsEqual(a, b *corev1.Service) bool {
	if len(a.Spec.Ports) != len(b.Spec.Ports) {
		return false
	}

	for i := 0; i < len(a.Spec.Ports); i++ {
		portA := a.Spec.Ports[i]
		portB := b.Spec.Ports[i]

		if portA.Name != portB.Name {
			return false
		}
		if !reflect.DeepEqual(portA.Port, portB.Port) {
			return false
		}
		if !reflect.DeepEqual(portA.TargetPort, portB.TargetPort) {
			return false
		}
		if !reflect.DeepEqual(portA.Protocol, portB.Protocol) {
			return false
		}
	}
	return true
}
