// Package configmapnamescontext stores and accesses the config map names in
// context.Context.
package configmapnamescontext

import (
	"context"

	"github.com/giantswarm/kvmtpr"

	servicekey "github.com/giantswarm/kvm-operator/service/key"
)

const (
	PrefixMaster = "master"
	PrefixWorker = "worker"
)

// key is an unexported type for keys defined in this package. This prevents
// collisions with keys defined in other packages.
type key string

// configMapNamesKey is the key for config map names values in context.Context.
// Clients use configmapnamescontext.NewContext and
// configmapnamescontext.FromContext instead of using this key directly.
var configMapNamesKey key = "configmapnames"

func NewChannel(obj interface{}) chan string {
	return make(chan string, len(getConfigMapNames(toCustomObject(obj))))
}

// NewContext returns a new context.Context that carries value v.
func NewContext(ctx context.Context, v chan string) context.Context {
	if v == nil {
		return ctx
	}

	return context.WithValue(ctx, configMapNamesKey, v)
}

// FromContext returns the config map names channel, if any.
func FromContext(ctx context.Context) (chan string, bool) {
	v, ok := ctx.Value(configMapNamesKey).(chan string)
	return v, ok
}

func getConfigMapNames(customObject kvmtpr.CustomObject) []string {
	var names []string

	for _, node := range customObject.Spec.Cluster.Masters {
		name := servicekey.ConfigMapName(customObject, node, PrefixMaster)
		names = append(names, name)
	}

	for _, node := range customObject.Spec.Cluster.Workers {
		name := servicekey.ConfigMapName(customObject, node, PrefixWorker)
		names = append(names, name)
	}

	return names
}

func toCustomObject(v interface{}) kvmtpr.CustomObject {
	customObjectPointer, ok := v.(*kvmtpr.CustomObject)
	if !ok {
		return kvmtpr.CustomObject{}
	}
	customObject := *customObjectPointer

	return customObject
}
