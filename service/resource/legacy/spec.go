package legacy

import (
	"k8s.io/apimachinery/pkg/runtime"
)

// Resource implements any Kubernetes resource being reconciled.
type Resource interface {
	// GetForCreate returns the Kubernetes runtime objects for any resource being
	// used in reconciliation loops on create events.
	GetForCreate(obj interface{}) ([]runtime.Object, error)
	// GetForDelete returns the Kubernetes runtime objects for any resource being
	// used in reconciliation loops on delete events.
	GetForDelete(obj interface{}) ([]runtime.Object, error)
}
