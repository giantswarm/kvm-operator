package v24

import (
	"github.com/giantswarm/versionbundle"
)

func VersionBundle() versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{
				Component:   "k8scloudconfig",
				Description: "Added configuration necessary for generic rbd storage support.",
				Kind:        versionbundle.KindChanged,
			},
		},
		Components: []versionbundle.Component{
			{
				Name:    "calico",
				Version: "3.7.2",
			},
			{
				Name:    "containerlinux",
				Version: "2079.5.1",
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
				Version: "1.14.3",
			},
		},
		Name:    "kvm-operator",
		Version: "3.8.0",
	}
}
