package clusterrolebindingv2

import (
	"reflect"

	"github.com/giantswarm/certs"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/framework"
	apiv1 "k8s.io/api/rbac/v1beta1"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/kvm-operator/service/cloudconfigv2"
)

const (
	// Name is the identifier of the resource.
	Name = "clusterrolebindingv2"

	VersionBundleVersionAnnotation = "giantswarm.io/version-bundle-version"
)

// Config represents the configuration used to create a new config map resource.
type Config struct {
	// Dependencies.
	CertSearcher certs.Interface
	CloudConfig  *cloudconfigv2.CloudConfig
	K8sClient    kubernetes.Interface
	Logger       micrologger.Logger
}

// DefaultConfig provides a default configuration to create a new config map
// resource by best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		CertSearcher: nil,
		CloudConfig:  nil,
		K8sClient:    nil,
		Logger:       nil,
	}
}

// Resource implements the config map resource.
type Resource struct {
	// Dependencies.
	certSearcher certs.Interface
	cloudConfig  *cloudconfigv2.CloudConfig
	k8sClient    kubernetes.Interface
	logger       micrologger.Logger
}

// New creates a new configured config map resource.
func New(config Config) (*Resource, error) {
	// Dependencies.
	if config.CertSearcher == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.CertSearcher must not be empty")
	}
	if config.CloudConfig == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.CloudConfig must not be empty")
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.K8sClient must not be empty")
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}

	newService := &Resource{
		// Dependencies.
		certSearcher: config.CertSearcher,
		cloudConfig:  config.CloudConfig,
		k8sClient:    config.K8sClient,
		logger: config.Logger.With(
			"resource", Name,
		),
	}

	return newService, nil
}

func (r *Resource) Name() string {
	return Name
}

func (r *Resource) Underlying() framework.Resource {
	return r
}

func containsClusterRoleBinding(list []*apiv1.ClusterRoleBinding, item *apiv1.ClusterRoleBinding) bool {
	_, err := getClusterRoleBindingByName(list, item.Name)
	if err != nil {
		return false
	}

	return true
}

func getClusterRoleBindingByName(list []*apiv1.ClusterRoleBinding, name string) (*apiv1.ClusterRoleBinding, error) {
	for _, l := range list {
		if l.Name == name {
			return l, nil
		}
	}

	return nil, microerror.Mask(notFoundError)
}

func isClusterRoleBindingModified(a, b *apiv1.ClusterRoleBinding) bool {
	return !reflect.DeepEqual(a.Subjects, b.Subjects) && !reflect.DeepEqual(a.RoleRef, b.RoleRef)
}

func toClusterRoleBindings(v interface{}) ([]*apiv1.ClusterRoleBinding, error) {
	if v == nil {
		return nil, nil
	}

	clusterRoleBindings, ok := v.([]*apiv1.ClusterRoleBinding)
	if !ok {
		return nil, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", []*apiv1.ClusterRoleBinding{}, v)
	}

	return clusterRoleBindings, nil
}
