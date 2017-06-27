package flannel

import (
	"github.com/giantswarm/kvmtpr"
	apiv1 "k8s.io/client-go/pkg/api/v1"
)

func (s *Service) newVolumes(customObject *kvmtpr.CustomObject) []apiv1.Volume {
	return []apiv1.Volume{
		{
			Name: "cgroup",
			VolumeSource: apiv1.VolumeSource{
				HostPath: &apiv1.HostPathVolumeSource{
					Path: "/sys/fs/cgroup",
				},
			},
		},
		{
			Name: "dbus",
			VolumeSource: apiv1.VolumeSource{
				HostPath: &apiv1.HostPathVolumeSource{
					Path: "/var/run/dbus",
				},
			},
		},
		{
			Name: "environment",
			VolumeSource: apiv1.VolumeSource{
				HostPath: &apiv1.HostPathVolumeSource{
					Path: "/etc/environment",
				},
			},
		},
		{
			Name: "etcd-certs",
			VolumeSource: apiv1.VolumeSource{
				HostPath: &apiv1.HostPathVolumeSource{
					Path: "/etc/giantswarm/g8s/ssl/etcd/",
				},
			},
		},
		{
			Name: "etc-systemd",
			VolumeSource: apiv1.VolumeSource{
				HostPath: &apiv1.HostPathVolumeSource{
					Path: "/etc/systemd/",
				},
			},
		},
		{
			Name: "flannel",
			VolumeSource: apiv1.VolumeSource{
				HostPath: &apiv1.HostPathVolumeSource{
					Path: "/run/flannel",
				},
			},
		},
		{
			Name: "ssl",
			VolumeSource: apiv1.VolumeSource{
				HostPath: &apiv1.HostPathVolumeSource{
					Path: "/etc/ssl/certs",
				},
			},
		},
		{
			Name: "systemctl",
			VolumeSource: apiv1.VolumeSource{
				HostPath: &apiv1.HostPathVolumeSource{
					Path: "/usr/bin/systemctl",
				},
			},
		},
		{
			Name: "systemd",
			VolumeSource: apiv1.VolumeSource{
				HostPath: &apiv1.HostPathVolumeSource{
					Path: "/run/systemd",
				},
			},
		},
		{
			Name: "sys-class-net",
			VolumeSource: apiv1.VolumeSource{
				HostPath: &apiv1.HostPathVolumeSource{
					Path: "/sys/class/net/",
				},
			},
		},
	}
}
