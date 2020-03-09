package node

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/kvm-operator/service/controller/controllercontext"
	"github.com/giantswarm/kvm-operator/service/controller/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	var release *v1alpha1.Release
	{
		releaseVersion := cr.Labels["release"]
		release, err = r.g8sClient.ReleaseV1alpha1().Releases().Get(releaseVersion, metav1.GetOptions{})
		if err != nil {
			return microerror.Mask(err)
		}
	}

	if release.Spec.Components[0].Name == "kubernetes" {
		cc.Spec.Versions.Kubernetes = release.Spec.Components[0].Version
	}

	return nil
}
