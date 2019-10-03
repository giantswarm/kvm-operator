package v25

import (
	"github.com/giantswarm/versionbundle"
)

func VersionBundle() versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{
				Component:   "TODO",
				Description: "TODO",
				Kind:        versionbundle.KindAdded,
			},
		},
		Components: []versionbundle.Component{
			{
				Name:    "calico",
				Version: "3.9.1",
			},
			{
				Name:    "containerlinux",
				Version: "2191.5.0",
			},
			{
				Name:    "docker",
				Version: "18.06.1",
			},
			{
				Name:    "etcd",
				Version: "3.3.15",
			},
			{
				Name:    "kubernetes",
				Version: "1.15.4",
			},
		},
		Name:    "kvm-operator",
		Version: "3.9.0",
	}
}
