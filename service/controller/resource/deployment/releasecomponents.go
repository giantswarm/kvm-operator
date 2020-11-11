package deployment

import (
	"github.com/giantswarm/kvm-operator/service/controller/key"

	releasev1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/release/v1alpha1"
)

var coreComponents = []string{"calico", "containerlinux", "etcd", "kubernetes"}

func addCoreComponentsAnnotations(annotations map[string]string, release *releasev1alpha1.Release) {
	for _, component := range coreComponents {
		version := componentVersion(release, component)
		if version == "" {
			continue
		}
		annotationName := key.AnnotationComponentVersionPrefix + "-" + component
		annotations[annotationName] = version
	}
}

func componentVersion(release *releasev1alpha1.Release, component string) string {
	for _, releaseComponent := range release.Spec.Components {
		if releaseComponent.Name == component {
			return releaseComponent.Version
		}
	}

	return ""
}
