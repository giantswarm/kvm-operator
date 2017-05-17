package k8s

import (
	"io"

	"k8s.io/client-go/pkg/runtime"
	"k8s.io/client-go/pkg/watch"
)

type ListDecoder interface {
	Decode(b []byte) (object runtime.Object, err error)
}

type Resource interface {
	GetForCreate(obj interface{}) (runtime.Object, error)
	GetForDelete(obj interface{}) (runtime.Object, error)
}

type WatchDecoder interface {
	Close()
	Decode() (action watch.EventType, object runtime.Object, err error)
	SetStream(stream io.ReadCloser)
}
