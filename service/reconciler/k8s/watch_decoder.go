package k8s

import (
	"encoding/json"
	"io"

	"github.com/giantswarm/kvmtpr"
	microerror "github.com/giantswarm/microkit/error"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
)

type watchDecoder struct {
	decoder *json.Decoder
	close   func() error
}

func (wd *watchDecoder) Close() {
	err := wd.close()
	if err != nil {
		panic(err)
	}
}

func (wd *watchDecoder) Decode() (watch.EventType, runtime.Object, error) {
	var e struct {
		Type   watch.EventType
		Object kvmtpr.CustomObject
	}

	err := wd.decoder.Decode(&e)
	if err != nil {
		return watch.Error, nil, microerror.MaskAny(err)
	}

	return e.Type, &e.Object, nil
}

func (wd *watchDecoder) SetStream(stream io.ReadCloser) {
	wd.decoder = json.NewDecoder(stream)
	wd.close = stream.Close
}
