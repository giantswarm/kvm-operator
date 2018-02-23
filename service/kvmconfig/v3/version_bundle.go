package v3

import (
	"time"

	"github.com/giantswarm/versionbundle"
)

func VersionBundle() versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{
				Component:   "kubernetes",
				Description: "Enable encryption at rest",
				Kind:        versionbundle.KindChanged,
			},
		},
		Components: []versionbundle.Component{
			{
				Name:    "calico",
				Version: "2.6.2",
			},
			{
				Name:    "docker",
				Version: "1.12.6",
			},
			{
				Name:    "etcd",
				Version: "3.2.7",
			},
			{
				Name:    "kubedns",
				Version: "1.14.5",
			},
			{
				Name:    "kubernetes",
				Version: "1.8.4",
			},
			{
				Name:    "nginx-ingress-controller",
				Version: "0.9.0",
			},
		},
		Dependencies: []versionbundle.Dependency{},
		Deprecated:   true,
		Name:         "kvm-operator",
		Time:         time.Date(2017, time.December, 19, 10, 00, 0, 0, time.UTC),
		Version:      "1.1.0",
		WIP:          false,
	}
}
