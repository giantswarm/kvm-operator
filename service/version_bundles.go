package service

import (
	"github.com/giantswarm/versionbundle"

	"github.com/giantswarm/kvm-operator/service/kvmconfig/v10"
	"github.com/giantswarm/kvm-operator/service/kvmconfig/v2"
	"github.com/giantswarm/kvm-operator/service/kvmconfig/v3"
	"github.com/giantswarm/kvm-operator/service/kvmconfig/v4"
	"github.com/giantswarm/kvm-operator/service/kvmconfig/v5"
	"github.com/giantswarm/kvm-operator/service/kvmconfig/v6"
	"github.com/giantswarm/kvm-operator/service/kvmconfig/v7"
	"github.com/giantswarm/kvm-operator/service/kvmconfig/v8"
	"github.com/giantswarm/kvm-operator/service/kvmconfig/v9"
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

	return versionBundles
}
