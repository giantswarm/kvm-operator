package pvc

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/framework"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	// Name is the identifier of the resource.
	Name = "pvcv5"
	// StorageClass is the storage class annotation persistent volume claims are
	// configured with.
	StorageClass = "g8s-storage"
)

// Config represents the configuration used to create a new PVC resource.
type Config struct {
	// Dependencies.
	K8sClient kubernetes.Interface
	Logger    micrologger.Logger
}

// DefaultConfig provides a default configuration to create a new PVC
// resource by best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		K8sClient: nil,
		Logger:    nil,
	}
}

// Resource implements the PVC resource.
type Resource struct {
	// Dependencies.
	k8sClient kubernetes.Interface
	logger    micrologger.Logger
}

// New creates a new configured PVC resource.
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
		logger: config.Logger.With(
			"resource", Name,
		),
	}

	return newResource, nil
}

func (r *Resource) Name() string {
	return Name
}

func (r *Resource) Underlying() framework.Resource {
	return r
}

func containsPVC(list []*apiv1.PersistentVolumeClaim, item *apiv1.PersistentVolumeClaim) bool {
	for _, l := range list {
		if l.Name == item.Name {
			return true
		}
	}

	return false
}

func toPVCs(v interface{}) ([]*apiv1.PersistentVolumeClaim, error) {
	if v == nil {
		return nil, nil
	}

	PVCs, ok := v.([]*apiv1.PersistentVolumeClaim)
	if !ok {
		return nil, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", []*apiv1.PersistentVolumeClaim{}, v)
	}

	return PVCs, nil
}
