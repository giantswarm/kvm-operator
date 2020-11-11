package deployment

import (
	"github.com/giantswarm/apiextensions/v3/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/tenantcluster/v4/pkg/tenantcluster"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/kvm-operator/service/controller/key"
)

const (
	// Name is the identifier of the resource.
	Name = "deployment"
)

// Config represents the configuration used to create a new deployment resource.
type Config struct {
	DNSServers    string
	G8sClient     versioned.Interface
	K8sClient     kubernetes.Interface
	Logger        micrologger.Logger
	NTPServers    string
	TenantCluster tenantcluster.Interface
}

// Resource implements the deployment resource.
type Resource struct {
	dnsServers    string
	g8sClient     versioned.Interface
	k8sClient     kubernetes.Interface
	logger        micrologger.Logger
	ntpServers    string
	tenantCluster tenantcluster.Interface
}

// New creates a new configured deployment resource.
func New(config Config) (*Resource, error) {
	if config.DNSServers == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.DNSServers must not be empty", config)
	}
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.TenantCluster == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.TenantCluster must not be empty", config)
	}

	newResource := &Resource{
		dnsServers:    config.DNSServers,
		g8sClient:     config.G8sClient,
		k8sClient:     config.K8sClient,
		logger:        config.Logger,
		ntpServers:    config.NTPServers,
		tenantCluster: config.TenantCluster,
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

func isAnnotationModified(a, b *v1.Deployment, annotation string) bool {
	aVersion := a.GetAnnotations()[annotation]
	if aVersion == "" {
		return true
	}

	bVersion := b.GetAnnotations()[annotation]
	if bVersion == "" {
		return true
	}

	return aVersion != bVersion
}

func isDeploymentModified(a, b *v1.Deployment) bool {
	if isAnnotationModified(a, b, key.VersionBundleVersionAnnotation) {
		return true
	}

	if isAnnotationModified(a, b, key.ReleaseVersionAnnotation) {
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
