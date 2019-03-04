package v19

import (
	"github.com/giantswarm/versionbundle"
)

func VersionBundle() versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{
				Component:   "kvm-operator",
				Description: "Add k8s api and kubelet health check to kvm pod.",
				Kind:        versionbundle.KindAdded,
			},
			{
				Component:   "kvm-operator",
				Description: "Add readiness probe to kvm pod to improve updates.",
				Kind:        versionbundle.KindAdded,
			},
			{
				Component:   "cloudconfig",
				Description: "Switch from cloudinit to ignition.",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "cloudconfig",
				Description: "Update tenant cluster container with Fedora 29.",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "node-operator",
				Description: "Improved node draining during updates and scaling.",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "cloudconfig",
				Description: "Add static iSCSI initiator name for each vm.",
				Kind:        versionbundle.KindAdded,
			},
			{
				Component:   "containerlinux",
				Description: "Update to 1967.5.0 (CVE-2019-5736).",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "kubernetes",
				Description: "Updated kubernetes to 1.13.3.",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "calico",
				Description: "Updated calico to 3.5.1.",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "etcd",
				Description: "Updated calico to 3.3.12.",
				Kind:        versionbundle.KindChanged,
			},
		},
		Components: []versionbundle.Component{
			{
				Name:    "calico",
				Version: "3.5.1",
			},
			{
				Name:    "containerlinux",
				Version: "1967.5.0",
			},
			{
				Name:    "docker",
				Version: "18.06.1",
			},
			{
				Name:    "etcd",
				Version: "3.3.12",
			},
			{
				Name:    "kubernetes",
				Version: "1.13.3",
			},
		},
		Name:    "kvm-operator",
		Version: "3.2.0",
	}
}
