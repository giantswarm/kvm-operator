package service

import (
	"github.com/giantswarm/versionbundle"

	"github.com/giantswarm/kvm-operator/service/controller/v14patch3"
	"github.com/giantswarm/kvm-operator/service/controller/v14patch4"
	"github.com/giantswarm/kvm-operator/service/controller/v15"
	"github.com/giantswarm/kvm-operator/service/controller/v16"
	"github.com/giantswarm/kvm-operator/service/controller/v17"
	"github.com/giantswarm/kvm-operator/service/controller/v17patch1"
	"github.com/giantswarm/kvm-operator/service/controller/v18"
	"github.com/giantswarm/kvm-operator/service/controller/v19"
)

func NewVersionBundles() []versionbundle.Bundle {
	var versionBundles []versionbundle.Bundle

	versionBundles = append(versionBundles, v14patch3.VersionBundle())
	versionBundles = append(versionBundles, v14patch4.VersionBundle())
	versionBundles = append(versionBundles, v15.VersionBundle())
	versionBundles = append(versionBundles, v16.VersionBundle())
	versionBundles = append(versionBundles, v17.VersionBundle())
	versionBundles = append(versionBundles, v17patch1.VersionBundle())
	versionBundles = append(versionBundles, v18.VersionBundle())
	versionBundles = append(versionBundles, v19.VersionBundle())

	return versionBundles
}
