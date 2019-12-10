package v26

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
		},
		Components: []versionbundle.Component{
			{
				Name:    "calico",
				Version: "3.9.1",
			},
			{
				Name:    "containerlinux",
				Version: "2191.5.0",
			},
			{
				Name:    "docker",
				Version: "18.06.1",
			},
			{
				Name:    "etcd",
				Version: "3.3.15",
			},
			{
				Name:    "kubernetes",
				Version: "1.15.5",
			},
		},
		Name:    "kvm-operator",
		Version: "3.10.0",
	}
}
