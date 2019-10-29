package deployment

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/kvm-operator/service/controller/v25/key"
)

const (
	// Name is the identifier of the resource.
	Name = "deploymentv25"
)

// Config represents the configuration used to create a new deployment resource.
type Config struct {
	DNSServers string
	K8sClient  kubernetes.Interface
	Logger     micrologger.Logger
	NTPServers string
}

// Resource implements the deployment resource.
type Resource struct {
	dnsServers string
	k8sClient  kubernetes.Interface
	logger     micrologger.Logger
	ntpServers string
}

// New creates a new configured deployment resource.
func New(config Config) (*Resource, error) {
	if config.DNSServers == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.DNSServers must not be empty", config)
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	newResource := &Resource{
		dnsServers: config.DNSServers,
		k8sClient:  config.K8sClient,
		logger:     config.Logger,
		ntpServers: config.NTPServers,
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

func containsDeployment(list []*v1.Deployment, item *v1.Deployment) bool {
	for _, l := range list {
		if l.Name == item.Name {
			return true
		}
	}

	return false
}

func getDeploymentByName(list []*v1.Deployment, name string) (*v1.Deployment, error) {
	for _, l := range list {
		if l.Name == name {
			return l, nil
		}
	}

	return nil, microerror.Mask(notFoundError)
}

func isDeploymentModified(a, b *v1.Deployment) bool {
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

func toDeployments(v interface{}) ([]*v1.Deployment, error) {
	if v == nil {
		return nil, nil
	}

	deployments, ok := v.([]*v1.Deployment)
	if !ok {
		return nil, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", []*v1.Deployment{}, v)
	}

	return deployments, nil
}
