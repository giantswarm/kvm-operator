package k8s

import (
	"io"

	"k8s.io/client-go/pkg/runtime"
	"k8s.io/client-go/pkg/watch"
)

// ListDecoder implements customized decoding of list responses received in
// Kubernetes client list-watches. The implementation has to ensure proper
// decoding of custom types.
type ListDecoder interface {
	Decode(b []byte) (runtime.Object, error)
}

// Resource implements any Kubernetes resource being reconciled.
type Resource interface {
	// GetForCreate returns the Kubernetes runtime objects for any resource being
	// used in reconciliation loops on create events.
	GetForCreate(obj interface{}) ([]runtime.Object, error)
	// GetForDelete returns the Kubernetes runtime objects for any resource being
	// used in reconciliation loops on delete events.
	GetForDelete(obj interface{}) ([]runtime.Object, error)
}

// WatchDecoder implements general decoding of watch streams created in
// Kubernetes client list-watches.
type WatchDecoder interface {
	// Close closes the configured stream.
	Close()
	// Decode decodes the received stream body.
	Decode() (watch.EventType, runtime.Object, error)
	// SetStream configures stream relevant settings. It must be called before
	// decoding.
	SetStream(stream io.ReadCloser)
}
