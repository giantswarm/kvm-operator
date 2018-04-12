package clusterrolebinding

import (
	"reflect"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	apiv1 "k8s.io/api/rbac/v1beta1"
	"k8s.io/client-go/kubernetes"
)

const (
	// Name is the identifier of the resource.
	Name = "clusterrolebindingv9"
)

// Config represents the configuration used to create a new config map resource.
type Config struct {
	K8sClient kubernetes.Interface
	Logger    micrologger.Logger
}

// Resource implements the config map resource.
type Resource struct {
	k8sClient kubernetes.Interface
	logger    micrologger.Logger
}

// New creates a new configured config map resource.
func New(config Config) (*Resource, error) {
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.K8sClient must not be empty")
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}

	newService := &Resource{
		k8sClient: config.K8sClient,
		logger:    config.Logger,
	}

	return newService, nil
}

func (r *Resource) Name() string {
	return Name
}

func containsClusterRoleBinding(list []*apiv1.ClusterRoleBinding, item *apiv1.ClusterRoleBinding) bool {
	_, err := getClusterRoleBindingByName(list, item.Name)
	if IsNotFound(err) {
		return false
	} else if err != nil {
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

	return nil, microerror.Maskf(notFoundError, "cluster role binding '%s' not found", name)
}

func isClusterRoleBindingModified(a, b *apiv1.ClusterRoleBinding) bool {
	return !reflect.DeepEqual(a.Subjects, b.Subjects) || !reflect.DeepEqual(a.RoleRef, b.RoleRef)
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
