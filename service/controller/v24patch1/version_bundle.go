package v24patch1

import (
	"github.com/giantswarm/versionbundle"
)

func VersionBundle() versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{
				Component:   "kubernetes",
				Description: "Updated from v1.14.6 to v1.14.9.",
				Kind:        versionbundle.KindChanged,
				URLs: []string{
					"https://github.com/giantswarm/kvm-operator/pull/791",
				},
			},
		},
		Components: []versionbundle.Component{
			{
				Name:    "calico",
				Version: "3.8.2",
			},
			{
				Name:    "containerlinux",
				Version: "2135.4.0",
			},
			{
				Name:    "docker",
				Version: "18.06.1",
			},
			{
				Name:    "etcd",
				Version: "3.3.13",
			},
			{
				Name:    "kubernetes",
				Version: "1.14.9",
			},
		},
		Name:    "kvm-operator",
		Version: "3.8.1",
	}
}
