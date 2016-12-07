package controller

import (
	"encoding/json"

	"github.com/giantswarm/cluster-controller/resources"

	"k8s.io/client-go/pkg/runtime"
	"k8s.io/client-go/pkg/watch"
)

type clusterDecoder struct {
	decoder *json.Decoder
	close   func() error
}

func (d *clusterDecoder) Close() {
	d.close()
}

func (d *clusterDecoder) Decode() (action watch.EventType, object runtime.Object, err error) {
	var e struct {
		Type   watch.EventType
		Object resources.Cluster
	}
	if err := d.decoder.Decode(&e); err != nil {
		return watch.Error, nil, err
	}
	return e.Type, &e.Object, nil
}
