package service

import (
	"github.com/giantswarm/versionbundle"

	"github.com/giantswarm/kvm-operator/service/controller/v11"
	"github.com/giantswarm/kvm-operator/service/controller/v12"
	"github.com/giantswarm/kvm-operator/service/controller/v13"
	"github.com/giantswarm/kvm-operator/service/controller/v14"
	"github.com/giantswarm/kvm-operator/service/controller/v14patch1"
	"github.com/giantswarm/kvm-operator/service/controller/v14patch2"
	"github.com/giantswarm/kvm-operator/service/controller/v14patch3"
	"github.com/giantswarm/kvm-operator/service/controller/v15"
	"github.com/giantswarm/kvm-operator/service/controller/v16"
	"github.com/giantswarm/kvm-operator/service/controller/v17"
	"github.com/giantswarm/kvm-operator/service/controller/v2"
	"github.com/giantswarm/kvm-operator/service/controller/v4"
)

func NewVersionBundles() []versionbundle.Bundle {
	var versionBundles []versionbundle.Bundle

	versionBundles = append(versionBundles, v2.VersionBundles()...) // NOTE this is special because it was created during the introduction of version bundles.
	versionBundles = append(versionBundles, v4.VersionBundle())
	versionBundles = append(versionBundles, v11.VersionBundle())
	versionBundles = append(versionBundles, v12.VersionBundle())
	versionBundles = append(versionBundles, v13.VersionBundle())
	versionBundles = append(versionBundles, v14.VersionBundle())
	versionBundles = append(versionBundles, v14patch1.VersionBundle())
	versionBundles = append(versionBundles, v14patch2.VersionBundle())
	versionBundles = append(versionBundles, v14patch3.VersionBundle())
	versionBundles = append(versionBundles, v15.VersionBundle())
	versionBundles = append(versionBundles, v16.VersionBundle())
	versionBundles = append(versionBundles, v17.VersionBundle())

	return versionBundles
}
