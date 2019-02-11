package v18

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
				Component:   "cloudconfig",
				Description: "Bring back missing kubelet resources reservation for systemd and container runtime.",
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
				Version: "1911.5.0",
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
				Version: "1.12.3",
			},
		},
		Name:    "kvm-operator",
		Version: "3.2.0",
	}
}
