package v19

import (
	"github.com/giantswarm/versionbundle"
)

func VersionBundle() versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{
				Component:   "kvm-operator",
				Description: "Changed iSCSI initiator name to use assigned node index instead of node ID.",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "kubernetes",
				Description: "Update to 1.13.4 (CVE-2019-1002100).",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "containerlinux",
				Description: "Update to 2023.4.0. Fixes CVE-2019-8912",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "kvm-operator",
				Description: "Fixed closing http idle connection in kvm health check.",
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
				Version: "2023.4.0",
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
				Version: "1.13.4",
			},
		},
		Name:    "kvm-operator",
		Version: "3.3.0",
	}
}
