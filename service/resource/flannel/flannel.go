package flannel

import (
	"encoding/json"

	microerror "github.com/giantswarm/microkit/error"
	micrologger "github.com/giantswarm/microkit/logger"
	"k8s.io/client-go/pkg/runtime"
)

// Config represents the configuration used to create a new service.
type Config struct {
	// Dependencies.
	Logger micrologger.Logger
}

// DefaultConfig provides a default configuration to create a new service by
// best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		Logger: nil,
	}
}

// New creates a new configured service.
func New(config Config) (*Service, error) {
	// Dependencies.
	if config.Logger == nil {
		return nil, microerror.MaskAnyf(invalidConfigError, "config.Logger must not be empty")
	}

	newService := &Service{
		// Dependencies.
		logger: config.Logger,
	}

	return newService, nil
}

// Service implements the flannel service.
type Service struct {
	// Dependencies.
	logger micrologger.Logger
}

// GetForCreate returns the Kubernetes runtime object for the flannel resource
// being used in reconciliation loops on create events.
func (s *Service) GetForCreate(obj interface{}) (runtime.Object, error) {
	ro, err := s.newRuntimeObject(obj)
	if err != nil {
		return nil, microerror.MaskAny(err)
	}

	return ro, nil
}

// GetForDelete returns the Kubernetes runtime object for the flannel resource
// being used in reconciliation loops on delete events. Note that deleting a
// flannel resources happens implicity by deleting the namespace the flannel
// resource is attached to. This is why we do not return any implementation
// here, but nil. That causes the reconcilliation to ignore the deletion for
// this resource.
func (s *Service) GetForDelete(obj interface{}) (runtime.Object, error) {
	return nil, nil
}

func (s *Service) newRuntimeObject(obj interface{}) (runtime.Object, error) {
	var initContainers string
	{
		ics, err := s.newInitContainers(obj)
		if err != nil {
			return nil, microerror.MaskAny(err)
		}
		marshalled, err := json.Marshal(ics)
		if err != nil {
			return nil, microerror.MaskAny(err)
		}
		initContainers = string(marshalled)
	}

	var podAffinity string
	{
		pa, err := s.newPodAfinity(obj)
		if err != nil {
			return nil, microerror.MaskAny(err)
		}
		marshalled, err := json.Marshal(pa)
		if err != nil {
			return nil, microerror.MaskAny(err)
		}
		podAffinity = string(marshalled)
	}

	deployment, err := s.newDeployment(obj)
	if err != nil {
		return nil, microerror.MaskAny(err)
	}
	deployment.Spec.Template.Annotations["pod.beta.kubernetes.io/init-containers"] = initContainers
	deployment.Spec.Template.Annotations["scheduler.alpha.kubernetes.io/affinity"] = podAffinity

	return deployment, nil
}
