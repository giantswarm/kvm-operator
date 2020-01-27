package v27

import (
	"github.com/giantswarm/versionbundle"
)

func VersionBundle() versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{
				Component:   "kvm-operator",
				Description: "Check resource version of control plane endpoint before updating IPs.",
				Kind:        versionbundle.KindChanged,
				URLs: []string{
					"https://github.com/giantswarm/kvm-operator/pull/770",
				},
			},
			{
				Component:   "kubernetes",
				Description: "Add Deny All as default Network Policy in kube-system and giantswarm namespaces.",
				Kind:        versionbundle.KindChanged,
				URLs: []string{
					"https://github.com/giantswarm/k8scloudconfig/pull/609",
				},
			},
			{
				Component:   "calico",
				Description: "Update from v3.9.1 to v3.10.1.",
				Kind:        versionbundle.KindChanged,
				URLs: []string{
					"https://github.com/giantswarm/kvm-operator/pull/785",
				},
			},
			{
				Component:   "containerlinux",
				Description: "Update from v2191.5.0 to v2247.6.0.",
				Kind:        versionbundle.KindChanged,
				URLs: []string{
					"https://github.com/giantswarm/kvm-operator/pull/785",
				},
			},
			{
				Component:   "etcd",
				Description: "Update from v3.3.15 to v3.3.17.",
				Kind:        versionbundle.KindChanged,
				URLs: []string{
					"https://github.com/giantswarm/kvm-operator/pull/785",
				},
			},
			{
				Component:   "kubernetes",
				Description: "Update from v1.15.5 to v1.16.3.",
				Kind:        versionbundle.KindAdded,
				URLs: []string{
					"https://github.com/giantswarm/kvm-operator/pull/785",
				},
			},
		},
		Components: []versionbundle.Component{
			{
				Name:    "calico",
				Version: "3.10.1",
			},
			{
				Name:    "containerlinux",
				Version: "2247.6.0",
			},
			{
				Name:    "docker",
				Version: "18.06.1",
			},
			{
				Name:    "etcd",
				Version: "3.3.17",
			},
			{
				Name:    "kubernetes",
				Version: "1.16.3",
			},
		},
		Name:    "kvm-operator",
		Version: "3.10.0",
	}
}
