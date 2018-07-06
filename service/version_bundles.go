package service

import (
	"github.com/giantswarm/versionbundle"

	"github.com/giantswarm/kvm-operator/service/controller/v10"
	"github.com/giantswarm/kvm-operator/service/controller/v11"
	"github.com/giantswarm/kvm-operator/service/controller/v12"
	"github.com/giantswarm/kvm-operator/service/controller/v13"
	"github.com/giantswarm/kvm-operator/service/controller/v14"
	"github.com/giantswarm/kvm-operator/service/controller/v2"
	"github.com/giantswarm/kvm-operator/service/controller/v3"
	"github.com/giantswarm/kvm-operator/service/controller/v4"
	"github.com/giantswarm/kvm-operator/service/controller/v5"
	"github.com/giantswarm/kvm-operator/service/controller/v6"
	"github.com/giantswarm/kvm-operator/service/controller/v7"
	"github.com/giantswarm/kvm-operator/service/controller/v8"
	"github.com/giantswarm/kvm-operator/service/controller/v9"
)

func NewVersionBundles() []versionbundle.Bundle {
	var versionBundles []versionbundle.Bundle

	versionBundles = append(versionBundles, v2.VersionBundles()...) // NOTE this is special because it was created during the introduction of version bundles.
	versionBundles = append(versionBundles, v3.VersionBundle())
	versionBundles = append(versionBundles, v4.VersionBundle())
	versionBundles = append(versionBundles, v5.VersionBundle())
	versionBundles = append(versionBundles, v6.VersionBundle())
	versionBundles = append(versionBundles, v7.VersionBundle())
	versionBundles = append(versionBundles, v8.VersionBundle())
	versionBundles = append(versionBundles, v9.VersionBundle())
	versionBundles = append(versionBundles, v10.VersionBundle())
	versionBundles = append(versionBundles, v11.VersionBundle())
	versionBundles = append(versionBundles, v12.VersionBundle())
	versionBundles = append(versionBundles, v13.VersionBundle())
	versionBundles = append(versionBundles, v14.VersionBundle())

	return versionBundles
}
