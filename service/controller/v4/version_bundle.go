package v4

import (
	"github.com/giantswarm/versionbundle"
)

func VersionBundle() versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{

				Component:   "containerlinux",
				Description: "Updated containerlinux version to 1576.5.0.",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "kubernetes",
				Description: "Fixed audit log.",
				Kind:        "fixed",
			},
		},
		Components: []versionbundle.Component{
			{
				Name:    "calico",
				Version: "2.6.2",
			},
			{
				Name:    "containerlinux",
				Version: "1576.5.0",
			},
			{
				Name:    "docker",
				Version: "17.09.0",
			},
			{
				Name:    "etcd",
				Version: "3.2.7",
			},
			{
				Name:    "kubedns",
				Version: "1.14.5",
			},
			{
				Name:    "kubernetes",
				Version: "1.8.4",
			},
			{
				Name:    "nginx-ingress-controller",
				Version: "0.9.0",
			},
		},
		Name:    "kvm-operator",
		Version: "1.2.0",
	}
}
