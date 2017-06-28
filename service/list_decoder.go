package service

import (
	"encoding/json"

	"k8s.io/apimachinery/pkg/runtime"

	"github.com/giantswarm/kvmtpr"
	microerror "github.com/giantswarm/microkit/error"
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
