package deployment

import releasev1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/release/v1alpha1"

var keyComponents = []string{"calico", "containerlinux", "etcd", "kubernetes"}

func keyReleaseComponentsEqual(a, b *releasev1alpha1.Release) bool {
	for _, keyComponent := range keyComponents {
		if componentVersion(a, keyComponent) != componentVersion(b, keyComponent) {
			return false
		}
	}
	return true
}

func componentVersion(release *releasev1alpha1.Release, component string) string {
	for _, releaseComponent := range release.Spec.Components {
		if releaseComponent.Name == component {
			return releaseComponent.Version
		}
	}

	return ""
}
