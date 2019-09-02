package service

import (
	"github.com/giantswarm/versionbundle"

	v20 "github.com/giantswarm/kvm-operator/service/controller/v20"
	v21 "github.com/giantswarm/kvm-operator/service/controller/v21"
	v22 "github.com/giantswarm/kvm-operator/service/controller/v22"
	v23 "github.com/giantswarm/kvm-operator/service/controller/v23"
	"github.com/giantswarm/kvm-operator/service/controller/v23patch1"
	v24 "github.com/giantswarm/kvm-operator/service/controller/v24"
	v25 "github.com/giantswarm/kvm-operator/service/controller/v25"
)

func NewVersionBundles() []versionbundle.Bundle {
	var versionBundles []versionbundle.Bundle

	versionBundles = append(versionBundles, v20.VersionBundle())
	versionBundles = append(versionBundles, v21.VersionBundle())
	versionBundles = append(versionBundles, v22.VersionBundle())
	versionBundles = append(versionBundles, v23.VersionBundle())
	versionBundles = append(versionBundles, v23patch1.VersionBundle())
	versionBundles = append(versionBundles, v24.VersionBundle())
	versionBundles = append(versionBundles, v25.VersionBundle())

	return versionBundles
}
