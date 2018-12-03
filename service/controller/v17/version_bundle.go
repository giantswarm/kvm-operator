package v17

import (
	"github.com/giantswarm/versionbundle"
)

func VersionBundle() versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{
				Component:   "kvm-operator",
				Description: "Add separate volume for /var/lib/kubelet and allow specifying its size.",
				Kind:        versionbundle.KindAdded,
			},
		},
		Components: []versionbundle.Component{
			{
				Name:    "calico",
				Version: "3.2.3",
			},
			{
				Name:    "containerlinux",
				Version: "1855.5.0",
			},
			{
				Name:    "docker",
				Version: "18.06.1",
			},
			{
				Name:    "etcd",
				Version: "3.3.9",
			},
			{
				Name:    "kubernetes",
				Version: "1.12.2",
			},
		},
		Name:    "kvm-operator",
		Version: "3.1.1",
	}
}
