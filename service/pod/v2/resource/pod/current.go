package pod

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/kvm-operator/service/pod/v2/key"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	currentPod, err := key.ToPod(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	fmt.Printf("\n")
	fmt.Printf("pod resource start: TODO reconciling pod\n")
	fmt.Printf("%#v\n", currentPod)
	fmt.Printf("pod resource end: TODO reconciling pod\n")
	fmt.Printf("\n")

	return nil, nil
}
