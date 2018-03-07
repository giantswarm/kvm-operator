package v8

import (
	"time"

	"github.com/giantswarm/versionbundle"
)

func VersionBundle() versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{
				Component:   "kvm-node-controller",
				Description: "Updated KVM node controller with pod status bugfix.",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "Calico",
				Description: "Updated to 3.0.2.",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "kubelet",
				Description: "Tune kubelet flags for protecting key units (kubelet and container runtime) from workload overloads.",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "etcd",
				Description: "Updated to 3.3.1.",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "qemu",
				Description: "Fixed formula for calculating qemu memory overhead.",
				Kind:        versionbundle.KindFixed,
			},
			{
				Component:   "monitoring",
				Description: "Added configuration for monitoring endpoint IP addresses.",
				Kind:        versionbundle.KindAdded,
			},
			{
				Component:   "cloudconfig",
				Description: "Enable aggregation layer to be able to extend kubernetes API.",
				Kind:        versionbundle.KindChanged,
			},
		},
		Components: []versionbundle.Component{
			{
				Name:    "calico",
				Version: "3.0.2",
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
				Version: "3.3.1",
			},
			{
				Name:    "coredns",
				Version: "1.0.6",
			},
			{
				Name:    "kubernetes",
				Version: "1.9.2",
			},
			{
				Name:    "nginx-ingress-controller",
				Version: "0.11.0",
			},
		},
		Dependencies: []versionbundle.Dependency{},
		Deprecated:   false,
		Name:         "kvm-operator",
		Time:         time.Date(2018, time.March, 1, 2, 57, 0, 0, time.UTC),
		Version:      "2.1.1",
		WIP:          true,
	}
}
