package podcondition

import (
	"fmt"

	"github.com/giantswarm/microerror"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/json"
)

type podConditionPatch struct {
	v1.PodCondition
}

func (p podConditionPatch) Type() types.PatchType {
	return types.StrategicMergePatchType
}

func (p podConditionPatch) Data(_ runtime.Object) ([]byte, error) {
	conditionJSON, err := json.Marshal(p)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	return []byte(fmt.Sprintf("{\"status\":{\"conditions\":[%s]}}", conditionJSON)), nil
}
