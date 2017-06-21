package worker

import (
	"encoding/json"

	"github.com/giantswarm/kvm-operator/service/resource/flannel"
	microerror "github.com/giantswarm/microkit/error"
	micrologger "github.com/giantswarm/microkit/logger"
	apiv1 "k8s.io/client-go/pkg/api/v1"
	extensionsv1 "k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/pkg/runtime"
)

// Config represents the configuration used to create a new service.
type Config struct {
	// Dependencies.
	Logger  micrologger.Logger
	Flannel *flannel.Service
}

// DefaultConfig provides a default configuration to create a new service by
// best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		Logger:  nil,
		Flannel: nil,
	}
}

// New creates a new configured service.
func New(config Config) (*Service, error) {
	// Dependencies.
	if config.Logger == nil {
		return nil, microerror.MaskAnyf(invalidConfigError, "config.Logger must not be empty")
	}
	if config.Flannel == nil {
		return nil, microerror.MaskAnyf(invalidConfigError, "config.Flannel must not be empty")
	}

	newService := &Service{
		// Dependencies.
		logger:  config.Logger,
		flannel: config.Flannel,
	}

	return newService, nil
}

// Service implements the worker service.
type Service struct {
	// Dependencies.
	logger  micrologger.Logger
	flannel *flannel.Service
}

// GetForCreate returns the Kubernetes runtime object for the worker resource
// being used in reconciliation loops on create events.
func (s *Service) GetForCreate(obj interface{}) ([]runtime.Object, error) {
	ros, err := s.newRuntimeObjects(obj)
	if err != nil {
		return nil, microerror.MaskAny(err)
	}

	return ros, nil
}

// GetForDelete returns the Kubernetes runtime object for the worker resource
// being used in reconciliation loops on delete events. Note that deleting a
// worker resources happens implicity by deleting the namespace the worker
// resource is attached to. This is why we do not return any implementation
// here, but nil. That causes the reconcilliation to ignore the deletion for
// this resource.
func (s *Service) GetForDelete(obj interface{}) ([]runtime.Object, error) {
	return nil, nil
}

func (s *Service) newRuntimeObjects(obj interface{}) ([]runtime.Object, error) {
	var err error
	var runtimeObjects []runtime.Object

	var flannelContainers []apiv1.Container
	{
		flannelContainers, err = s.flannel.Containers(obj)
		if err != nil {
			return nil, microerror.MaskAny(err)
		}
	}

	var flannelInitContainers []apiv1.Container
	{
		flannelInitContainers, err = s.flannel.InitContainers(obj)
		if err != nil {
			return nil, microerror.MaskAny(err)
		}
	}

	var flannelVolumes []apiv1.Volume
	{
		flannelVolumes, err = s.flannel.Volumes(obj)
		if err != nil {
			return nil, microerror.MaskAny(err)
		}
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

	var newDeployments []*extensionsv1.Deployment
	{
		newDeployments, err = s.newDeployments(obj)
		if err != nil {
			return nil, microerror.MaskAny(err)
		}
		for i, _ := range newDeployments {
			ics := flannelInitContainers
			icsBytes, err := json.Marshal(ics)
			if err != nil {
				return nil, microerror.MaskAny(err)
			}
			newDeployments[i].Spec.Template.Annotations["pod.beta.kubernetes.io/init-containers"] = string(icsBytes)
			newDeployments[i].Spec.Template.Annotations["scheduler.alpha.kubernetes.io/affinity"] = podAffinity
			newDeployments[i].Spec.Template.Spec.Containers = append(flannelContainers, newDeployments[i].Spec.Template.Spec.Containers...)
			newDeployments[i].Spec.Template.Spec.Volumes = append(flannelVolumes, newDeployments[i].Spec.Template.Spec.Volumes...)
		}
	}

	var newService *apiv1.Service
	{
		newService, err = s.newService(obj)
		if err != nil {
			return nil, microerror.MaskAny(err)
		}
	}

	runtimeObjects = append(runtimeObjects, newService)
	for _, d := range newDeployments {
		runtimeObjects = append(runtimeObjects, d)
	}

	return runtimeObjects, nil
}
