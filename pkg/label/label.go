// Package label contains common Kubernetes object labels. These are defined in
// https://github.com/giantswarm/fmt/blob/master/kubernetes/annotations_and_labels.md.
package label

const (
	// Cluster label is a new style label for ClusterID.
	Cluster = "giantswarm.io/cluster"
	// ManagedBy is set for Kubernetes resources managed by the operator.
	ManagedBy = "giantswarm.io/managed-by"
	// Organization label denotes tenant cluster's organization ID as displayed
	// in the front-end.
	Organization   = "giantswarm.io/organization"
	ReleaseVersion = "release.giantswarm.io/version"
	ControlPlane   = "cluster.x-k8s.io/control-plane"
)

const (
	OperatorVersion = "kvm-operator.giantswarm.io/version"
)
