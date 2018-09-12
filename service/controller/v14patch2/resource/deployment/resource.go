package deployment

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/kvm-operator/service/controller/v14patch1/key"
)

const (
	// Name is the identifier of the resource.
	Name = "deploymentv14patch1"
)

// Config represents the configuration used to create a new deployment resource.
type Config struct {
	// Dependencies.
	K8sClient kubernetes.Interface
	Logger    micrologger.Logger
}

// DefaultConfig provides a default configuration to create a new deployment
// resource by best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		K8sClient: nil,
		Logger:    nil,
	}
}

// Resource implements the deployment resource.
type Resource struct {
	// Dependencies.
	k8sClient kubernetes.Interface
	logger    micrologger.Logger
}

// New creates a new configured deployment resource.
func New(config Config) (*Resource, error) {
	// Dependencies.
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.K8sClient must not be empty")
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}

	newResource := &Resource{
		// Dependencies.
		k8sClient: config.K8sClient,
		logger:    config.Logger,
	}

	return newResource, nil
}

func (r *Resource) Name() string {
	return Name
}

func allNumbersEqual(numbers ...int32) bool {
	if len(numbers) == 0 {
		return false
	}

	first := numbers[0]

	for _, n := range numbers {
		if n != first {
			return false
		}
	}

	return true
}

func containsDeployment(list []*v1beta1.Deployment, item *v1beta1.Deployment) bool {
	for _, l := range list {
		if l.Name == item.Name {
			return true
		}
	}

	return false
}

func getDeploymentByName(list []*v1beta1.Deployment, name string) (*v1beta1.Deployment, error) {
	for _, l := range list {
		if l.Name == name {
			return l, nil
		}
	}

	return nil, microerror.Mask(notFoundError)
}

func isDeploymentModified(a, b *v1beta1.Deployment) bool {
	aVersion, ok := a.GetAnnotations()[key.VersionBundleVersionAnnotation]
	if !ok {
		return true
	}
	if aVersion == "" {
		return true
	}

	bVersion, ok := b.GetAnnotations()[key.VersionBundleVersionAnnotation]
	if !ok {
		return true
	}
	if bVersion == "" {
		return true
	}

	if aVersion != bVersion {
		return true
	}

	return false
}

func toDeployments(v interface{}) ([]*v1beta1.Deployment, error) {
	if v == nil {
		return nil, nil
	}

	deployments, ok := v.([]*v1beta1.Deployment)
	if !ok {
		return nil, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", []*v1beta1.Deployment{}, v)
	}

	return deployments, nil
}
