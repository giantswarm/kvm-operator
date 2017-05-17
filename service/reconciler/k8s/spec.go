package k8s

import (
	"io"

	"k8s.io/client-go/pkg/runtime"
	"k8s.io/client-go/pkg/watch"
)

type ListDecoder interface {
	Decode(b []byte) (runtime.Object, error)
}

type Resource interface {
	GetForCreate(obj interface{}) (runtime.Object, error)
	GetForDelete(obj interface{}) (runtime.Object, error)
}

type WatchDecoder interface {
	Close()
	Decode() (watch.EventType, runtime.Object, error)
	SetStream(stream io.ReadCloser)
}
