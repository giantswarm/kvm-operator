package project

import (
	"github.com/giantswarm/versionbundle"
)

func NewVersionBundle() versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{
				Component:   "containerlinux",
				Description: "Deprecate CoreOS and move to Flatcar Linux.",
				Kind:        versionbundle.KindChanged,
				URLs: []string{
					"https://github.com/giantswarm/kvm-operator/pull/825",
					"https://github.com/giantswarm/k8s-kvm/pull/40",
				},
			},
		},
		Components: []versionbundle.Component{
			{
				Name:    "calico",
				Version: "3.10.1",
			},
			{
				Name:    "containerlinux",
				Version: "2345.3.0",
			},
			{
				Name:    "docker",
				Version: "18.06.1",
			},
			{
				Name:    "etcd",
				Version: "3.3.17",
			},
			{
				Name:    "kubernetes",
				Version: "1.16.3",
			},
		},
		Name:    Name(),
		Version: Version(),
	}
}
