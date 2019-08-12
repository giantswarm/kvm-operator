package service

import (
	"github.com/giantswarm/versionbundle"

	"github.com/giantswarm/kvm-operator/service/controller/v20"
	"github.com/giantswarm/kvm-operator/service/controller/v21"
	"github.com/giantswarm/kvm-operator/service/controller/v22"
	"github.com/giantswarm/kvm-operator/service/controller/v23"
	"github.com/giantswarm/kvm-operator/service/controller/v23patch1"
	"github.com/giantswarm/kvm-operator/service/controller/v24"
)

func NewVersionBundles() []versionbundle.Bundle {
	var versionBundles []versionbundle.Bundle

	versionBundles = append(versionBundles, v20.VersionBundle())
	versionBundles = append(versionBundles, v21.VersionBundle())
	versionBundles = append(versionBundles, v22.VersionBundle())
	versionBundles = append(versionBundles, v23.VersionBundle())
	versionBundles = append(versionBundles, v23patch1.VersionBundle())
	versionBundles = append(versionBundles, v24.VersionBundle())

	return versionBundles
}
