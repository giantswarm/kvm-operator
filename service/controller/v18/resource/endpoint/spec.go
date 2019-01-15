package endpoint

import (
	corev1 "k8s.io/api/core/v1"
)

// TODO Addresses and IPs are basically the same and we only need Addresses in
// fact. The whole thing could be refactored and simplified a little bit later.
type Endpoint struct {
	Addresses        []corev1.EndpointAddress
	Ports            []corev1.EndpointPort
	IPs              []string
	ServiceName      string
	ServiceNamespace string
}
