package deployment

import (
	releasev1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/release/v1alpha1"
	v1 "k8s.io/api/apps/v1"

	"github.com/giantswarm/kvm-operator/service/controller/key"
)

var coreComponents = []string{"calico", "containerlinux", "etcd", "kubernetes"}

// addCoreComponentsAnnotations adds an annotation for each core component to the pod template spec
func addCoreComponentsAnnotations(deployment *v1.Deployment, release releasev1alpha1.Release) {
	for _, component := range coreComponents {
		version := componentVersion(release, component)
		if version == "" {
			// component is not present in the release
			continue
		}
		annotationName := key.AnnotationComponentVersionPrefix + "-" + component
		deployment.Spec.Template.ObjectMeta.Annotations[annotationName] = version
	}
}

func componentVersion(release releasev1alpha1.Release, component string) string {
	for _, releaseComponent := range release.Spec.Components {
		if releaseComponent.Name == component {
			return releaseComponent.Version
		}
	}

	return ""
}
