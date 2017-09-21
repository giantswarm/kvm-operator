// Package configmapnamescontext stores and accesses the config map names in
// context.Context.
package configmapnamescontext

import (
	"context"

	servicekey "github.com/giantswarm/kvm-operator/service/key"
	"github.com/giantswarm/microerror"
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

func NewChannel(obj interface{}) (chan string, error) {
	customObject, err := servicekey.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	ch := make(chan string, len(servicekey.ConfigMapNames(customObject)))

	return ch, nil
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
