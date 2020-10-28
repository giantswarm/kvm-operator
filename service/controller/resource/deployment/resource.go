package deployment

import (
	"strings"

	releasev1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/tenantcluster"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func (r *Resource) getRelease(releaseName string) (*releasev1alpha1.Release, error) {
	if !strings.HasPrefix(releaseName, "v") {
		releaseName = "v" + releaseName
	}

	release, err := r.g8sClient.ReleaseV1alpha1().Releases().Get(releaseName, metav1.GetOptions{})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return release, nil
}

func (r *Resource) isDeploymentModified(a, b *v1.Deployment) (bool, error) {
	versionA, versionB, present := getAnnotationsIfPresent(a, b, key.VersionBundleVersionAnnotation)
	if !present || (versionA != versionB) {
		return true, nil
	}

	releaseA, releaseB, present := getAnnotationsIfPresent(a, b, key.ReleaseVersionAnnotation)
	if !present {
		return true, nil
	}

	if releaseA != releaseB {
		componentsChanged, err := r.releaseComponentsChanged(releaseA, releaseB)
		if err != nil {
			return false, err
		}

		if componentsChanged {
			return true, nil
		}
	}

	return false, nil
}

func (r *Resource) releaseComponentsChanged(a, b string) (bool, error) {
	aRelease, err := r.getRelease(a)
	if err != nil {
		return false, microerror.Mask(err)
	}

	bRelease, err := r.getRelease(b)
	if err != nil {
		return false, microerror.Mask(err)
	}

	if !keyReleaseComponentsEqual(aRelease, bRelease) {
		return true, nil
	}

	return false, nil
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

func getAnnotationsIfPresent(a, b *v1.Deployment, annotation string) (string, string, bool) {
	aVersion := a.GetAnnotations()[annotation]
	bVersion := b.GetAnnotations()[annotation]
	if aVersion == "" || bVersion == "" {
		return aVersion, bVersion, false
	}

	return aVersion, bVersion, true
}

func getDeploymentByName(list []*v1.Deployment, name string) (*v1.Deployment, error) {
	for _, l := range list {
		if l.Name == name {
			return l, nil
		}
	}

	return nil, microerror.Mask(notFoundError)
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
