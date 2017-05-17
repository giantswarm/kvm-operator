package service

import (
	"encoding/json"

	"github.com/giantswarm/kvmtpr"
	microerror "github.com/giantswarm/microkit/error"
	"k8s.io/client-go/pkg/runtime"
)

type listDecoder struct{}

func (ld *listDecoder) Decode(b []byte) (runtime.Object, error) {
	var l kvmtpr.List

	err := json.Unmarshal(b, &l)
	if err != nil {
		return nil, microerror.MaskAny(err)
	}

	return &l, nil
}
