package clusterrolebinding

import (
	"reflect"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	apiv1 "k8s.io/api/rbac/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// Name is the identifier of the resource.
	Name = "clusterrolebinding"
)

// Config represents the configuration used to create a new config map resource.
type Config struct {
	CtrlClient client.Client
	Logger     micrologger.Logger

	ClusterRoleGeneral string
	ClusterRolePSP     string
}

// Resource implements the config map resource.
type Resource struct {
	ctrlClient client.Client
	logger     micrologger.Logger

	clusterRoleGeneral string
	clusterRolePSP     string
}

// New creates a new configured config map resource.
func New(config Config) (*Resource, error) {
	if config.CtrlClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CtrlClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.ClusterRoleGeneral == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.ClusterRoleGeneral must not be empty", config)
	}
	if config.ClusterRolePSP == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.ClusterRolePSP must not be empty", config)
	}

	newService := &Resource{
		ctrlClient: config.CtrlClient,
		logger:     config.Logger,

		clusterRoleGeneral: config.ClusterRoleGeneral,
		clusterRolePSP:     config.ClusterRolePSP,
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
