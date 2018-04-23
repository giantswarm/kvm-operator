package v11

import (
	"time"

	"github.com/giantswarm/versionbundle"
)

func VersionBundle() versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{
				Component:   "kubernetes",
				Description: "Updated to 1.10.1.",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "cloudconfig",
				Description: "Updated kube-state-metrics to version 1.3.1.",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "cloudconfig",
				Description: "Changed kubelet bind mount mode from shared to rshared.",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "cloudconfig",
				Description: "Disabled etcd3-defragmentation service in favor systemd timer.",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "cloudconfig",
				Description: "Added /lib/modules mount for kubelet.",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "cloudconfig",
				Description: "Updated CoreDNS to 1.1.1.",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "cloudconfig",
				Description: "Updated Calico to 3.0.5.",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "cloudconfig",
				Description: "Updated Etcd to 3.3.3.",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "cloudconfig",
				Description: "Removed docker flag --disable-legacy-registry.",
				Kind:        versionbundle.KindRemoved,
			},
			{
				Component:   "cloudconfig",
				Description: "Removed calico-ipip-pinger.",
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
				Version: "1.10.1",
			},
			{
				Name:    "nginx-ingress-controller",
				Version: "0.12.0",
			},
		},
		Dependencies: []versionbundle.Dependency{},
		Deprecated:   false,
		Name:         "kvm-operator",
		Time:         time.Date(2018, time.April, 04, 17, 32, 0, 0, time.UTC),
		Version:      "2.2.0",
		WIP:          false,
	}
}
