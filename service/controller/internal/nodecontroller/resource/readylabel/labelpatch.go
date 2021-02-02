package readylabel

import (
	"encoding/json"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

type annotationPatch struct {
	key   string
	value string
}

func (p annotationPatch) Type() types.PatchType {
	return types.JSONPatchType
}

func (p annotationPatch) Data(_ runtime.Object) ([]byte, error) {
	return json.Marshal([]struct {
		Op    string `json:"op"`
		Path  string `json:"path"`
		Value string `json:"value"`
	}{
		{
			Op:    "replace",
			Path:  fmt.Sprintf("/metadata/annotations/%s", p.key),
			Value: p.value,
		},
	})
}
