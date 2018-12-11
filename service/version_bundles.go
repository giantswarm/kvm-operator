package service

import (
	"github.com/giantswarm/versionbundle"

	"github.com/giantswarm/kvm-operator/service/controller/v14patch3"
	"github.com/giantswarm/kvm-operator/service/controller/v15"
	"github.com/giantswarm/kvm-operator/service/controller/v16"
	"github.com/giantswarm/kvm-operator/service/controller/v17"
)

func NewVersionBundles() []versionbundle.Bundle {
	var versionBundles []versionbundle.Bundle

	versionBundles = append(versionBundles, v14patch3.VersionBundle())
	versionBundles = append(versionBundles, v15.VersionBundle())
	versionBundles = append(versionBundles, v16.VersionBundle())
	versionBundles = append(versionBundles, v17.VersionBundle())

	return versionBundles
}
