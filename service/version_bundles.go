package service

import (
	"time"

	"github.com/giantswarm/versionbundle"
)

func NewVersionBundles() []versionbundle.Bundle {
	return []versionbundle.Bundle{
		{
			Changelogs: []versionbundle.Changelog{
				{
					Component:   "calico",
					Description: "Calico version updated.",
					Kind:        versionbundle.KindChanged,
				},
				{
					Component:   "docker",
					Description: "Docker version updated.",
					Kind:        versionbundle.KindChanged,
				},
				{
					Component:   "etcd",
					Description: "Etcd version updated.",
					Kind:        versionbundle.KindChanged,
				},
				{
					Component:   "kubedns",
					Description: "KubeDNS version updated.",
					Kind:        versionbundle.KindChanged,
				},
				{
					Component:   "kubernetes",
					Description: "Kubernetes version updated.",
					Kind:        versionbundle.KindChanged,
				},
				{
					Component:   "nginx-ingress-controller",
					Description: "Nginx-ingress-controller version updated.",
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
					Version: "1.8.1",
				},
				{
					Name:    "nginx-ingress-controller",
					Version: "0.9.0",
				},
			},
			Dependencies: []versionbundle.Dependency{},
			Deprecated:   true,
			Name:         "kvm-operator",
			Time:         time.Date(2017, time.October, 26, 16, 38, 0, 0, time.UTC),
			Version:      "0.1.0",
			WIP:          false,
		},
		{
			Changelogs: []versionbundle.Changelog{
				{
					Component:   "kubernetes",
					Description: "Updated to kubernetes 1.8.4. Fixes a goroutine leak in the k8s api.",
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
			Time:         time.Date(2017, time.December, 12, 10, 00, 0, 0, time.UTC),
			Version:      "1.0.0",
			WIP:          false,
		},
		{
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
			Deprecated:   false,
			Name:         "kvm-operator",
			Time:         time.Date(2017, time.December, 19, 10, 00, 0, 0, time.UTC),
			Version:      "1.1.0",
			WIP:          false,
		},
		{
			Changelogs: []versionbundle.Changelog{
				{

					Component:   "containerlinux",
					Description: "Updated containerlinux version to 1576.5.0.",
					Kind:        versionbundle.KindChanged,
				},
				{
					Component:   "kubernetes",
					Description: "Fix audit log.",
					Kind:        "fixed",
				},
			},
			Components: []versionbundle.Component{
				{
					Name:    "calico",
					Version: "2.6.2",
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
			Deprecated:   false,
			Name:         "kvm-operator",
			Time:         time.Date(2018, time.February, 8, 6, 25, 0, 0, time.UTC),
			Version:      "1.1.1",
			WIP:          true,
		},
	}
}
