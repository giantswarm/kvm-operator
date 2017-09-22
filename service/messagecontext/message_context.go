// Package messagecontext stores and accesses the message struct in
// context.Context.
package messagecontext

import (
	"context"
)

// key is an unexported type for keys defined in this package. This prevents
// collisions with keys defined in other packages.
type key string

// messageKey is the key for message struct values in context.Context. Clients
// use messagecontext.NewContext and messagecontext.FromContext instead of using
// this key directly.
var messageKey key = "message"

type Message struct {
	ConfigMapNames []string
}

func NewMessage() *Message {
	return &Message{}
}

// NewContext returns a new context.Context that carries value v.
func NewContext(ctx context.Context, v *Message) context.Context {
	if v == nil {
		return ctx
	}

	return context.WithValue(ctx, messageKey, v)
}

// FromContext returns the message struct, if any.
func FromContext(ctx context.Context) (*Message, bool) {
	v, ok := ctx.Value(messageKey).(*Message)
	return v, ok
}
