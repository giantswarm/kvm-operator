package pvc

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// Name is the identifier of the resource.
	Name = "pvc"
	// StorageClass is the storage class annotation persistent volume claims are
	// configured with.
	StorageClass = "g8s-storage"
)

// Config represents the configuration used to create a new PVC resource.
type Config struct {
	// Dependencies.
	CtrlClient client.Client
	Logger     micrologger.Logger
}

// Resource implements the PVC resource.
type Resource struct {
	// Dependencies.
	ctrlClient client.Client
	logger     micrologger.Logger
}

// New creates a new configured PVC resource.
func New(config Config) (*Resource, error) {
	// Dependencies.
	if config.CtrlClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.CtrlClient must not be empty")
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}

	newResource := &Resource{
		// Dependencies.
		ctrlClient: config.CtrlClient,
		logger:     config.Logger,
	}

	return newResource, nil
}

func (r *Resource) Name() string {
	return Name
}

func containsPVC(list []*corev1.PersistentVolumeClaim, item *corev1.PersistentVolumeClaim) bool {
	for _, l := range list {
		if l.Name == item.Name {
			return true
		}
	}

	return false
}

func toPVCs(v interface{}) ([]*corev1.PersistentVolumeClaim, error) {
	if v == nil {
		return nil, nil
	}

	PVCs, ok := v.([]*corev1.PersistentVolumeClaim)
	if !ok {
		return nil, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", []*corev1.PersistentVolumeClaim{}, v)
	}

	return PVCs, nil
}
