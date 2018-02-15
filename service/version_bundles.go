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
			Deprecated:   true,
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
					Description: "Fixed audit log.",
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
			Deprecated:   true,
			Name:         "kvm-operator",
			Time:         time.Date(2018, time.February, 8, 6, 25, 0, 0, time.UTC),
			Version:      "1.2.0",
			WIP:          false,
		},
		{
			Changelogs: []versionbundle.Changelog{
				{
					Component:   "Kubernetes",
					Description: "Updated to Kubernetes 1.9.2.",
					Kind:        versionbundle.KindChanged,
				},
				{
					Component:   "Kubernetes",
					Description: "Switched to vanilla (previously CoreOS) hyperkube image.",
					Kind:        versionbundle.KindChanged,
				},
				{
					Component:   "Docker",
					Description: "Updated to 17.09.0-ce.",
					Kind:        versionbundle.KindChanged,
				},
				{
					Component:   "Calico",
					Description: "Updated to 3.0.1.",
					Kind:        versionbundle.KindChanged,
				},
				{
					Component:   "CoreDNS",
					Description: "Version 1.0.5 replaces kube-dns.",
					Kind:        versionbundle.KindAdded,
				},
				{
					Component:   "Nginx Ingress Controller",
					Description: "Updated to 0.10.2.",
					Kind:        versionbundle.KindChanged,
				},
				{
					Component:   "cloudconfig",
					Description: "Add OIDC integration for Kubernetes api-server.",
					Kind:        versionbundle.KindAdded,
				},
				{
					Component:   "cloudconfig",
					Description: "Replace systemd units for Kubernetes components with self-hosted pods.",
					Kind:        versionbundle.KindChanged,
				},
				{

					Component:   "containerlinux",
					Description: "Updated Container Linux version to 1576.5.0.",
					Kind:        versionbundle.KindChanged,
				},
			},
			Components: []versionbundle.Component{
				{
					Name:    "calico",
					Version: "3.0.1",
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
					Name:    "coredns",
					Version: "1.0.5",
				},
				{
					Name:    "kubernetes",
					Version: "1.9.2",
				},
				{
					Name:    "nginx-ingress-controller",
					Version: "0.10.2",
				},
			},
			Dependencies: []versionbundle.Dependency{},
			Deprecated:   false,
			Name:         "kvm-operator",
			Time:         time.Date(2018, time.February, 15, 2, 27, 0, 0, time.UTC),
			Version:      "2.0.0",
			WIP:          false,
		},
	}
}
