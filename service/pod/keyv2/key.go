package keyv2

import (
	"github.com/giantswarm/microerror"
	apiv1 "k8s.io/api/core/v1"
)

const (
	PodWatcherLabel = "giantswarm.io/pod-watcher"
)

func ToPod(v interface{}) (*apiv1.Pod, error) {
	if v == nil {
		return nil, nil
	}

	pod, ok := v.(*apiv1.Pod)
	if !ok {
		return nil, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &apiv1.Pod{}, v)
	}

	return pod, nil
}
