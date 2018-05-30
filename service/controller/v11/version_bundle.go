package v11

import (
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
				Description: "Fixed node-exporter running in container by adjusting host mounts.",
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
				Description: "Added trusted certificate CNs to aggregation API allowed names.",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "cloudconfig",
				Description: "Disabled SSL passthrough in nginx-ingress-controller.",
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
			{
				Component:   "kvm-operator",
				Description: "Added node draining support.",
				Kind:        versionbundle.KindAdded,
			},
			{
				Component:   "cloudconfig",
				Description: "Fixed etcd mount path.",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "cloudconfig",
				Description: "Removed explicit hostname labeling for kubelet.",
				Kind:        versionbundle.KindChanged,
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
		Name:    "kvm-operator",
		Version: "2.2.0",
	}
}
