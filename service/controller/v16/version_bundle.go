package v16

import (
	"github.com/giantswarm/versionbundle"
)

func VersionBundle() versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{
				Component:   "cloudconfig",
				Description: "Updated Calico to 3.2.3.",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "cloudconfig",
				Description: "Updated Calico manifest with resource limits.",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "cloudconfig",
				Description: "Enabled admission plugins: DefaultTolerationSeconds, MutatingAdmissionWebhook, ValidatingAdmissionWebhook.",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "cloudconfig",
				Description: "Use patched GiantSwarm build of Kubernetes (hyperkube:v1.11.1-cec4fb8023db783fbf26fb056bf6c76abfcd96cf-giantswarm).",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "TODO",
				Description: "TODO",
				Kind:        versionbundle.KindChanged,
			},
		},
		Components: []versionbundle.Component{
			{
				Name:    "calico",
				Version: "3.2.3",
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
				Version: "3.3.8",
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
		Version: "3.0.1",
	}
}
