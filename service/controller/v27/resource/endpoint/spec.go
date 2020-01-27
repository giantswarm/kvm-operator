package endpoint

import (
	corev1 "k8s.io/api/core/v1"
)

// TODO Addresses and IPs are basically the same and we only need Addresses in
// fact. The whole thing could be refactored and simplified a little bit later.
//
//     https://github.com/giantswarm/giantswarm/issues/5154
//
type Endpoint struct {
	Addresses []corev1.EndpointAddress
	Ports     []corev1.EndpointPort
	// IPs contains string representations of IP addresses. When the current state
	// is computed the list contains all IPs of the endpoint belonging to the
	// reconciled pod. When the desired state is computed the list contains only a
	// single IP, which is the IP of the reconciled pod.
	IPs              []string
	ResourceVersion  string
	ServiceName      string
	ServiceNamespace string
}
