package nodeready

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

type annotationPatch struct {
	key   string
	value string
}

func (p annotationPatch) Type() types.PatchType {
	return types.StrategicMergePatchType
}

func (p annotationPatch) Data(_ runtime.Object) ([]byte, error) {
	return []byte(fmt.Sprintf("{\"metadata\":{\"annotations\":{\"%s\":\"%s\"}}}", p.key, p.value)), nil
}
