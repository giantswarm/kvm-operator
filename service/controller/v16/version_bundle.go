package v16

import (
	"github.com/giantswarm/versionbundle"
)

func VersionBundle() versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{
				Component:   "kubernetes",
				Description: "Updated to 1.11.1.",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "cloudconfig",
				Description: "Removed nginx-ingress-controller related components (will be managed by chart-operator).",
				Kind:        versionbundle.KindRemoved,
			},
			{
				Component:   "kvm-operator",
				Description: "Made worker node docker volume size to be configurable.",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "cloudconfig",
				Description: "Removed CoreDNS related components (now managed by chart-operator).",
				Kind:        versionbundle.KindRemoved,
			},
		},
		Components: []versionbundle.Component{
			{
				Name:    "calico",
				Version: "3.0.5",
			},
			{
				Name:    "containerlinux",
				Version: "1688.5.3",
			},
			{
				Name:    "docker",
				Version: "17.12.1",
			},
			{
				Name:    "etcd",
				Version: "3.3.3",
			},
			{
				Name:    "coredns",
				Version: "1.1.1",
			},
			{
				Name:    "kubernetes",
				Version: "1.11.1",
			},
		},
		Name:    "kvm-operator",
		Version: "3.0.0",
	}
}
