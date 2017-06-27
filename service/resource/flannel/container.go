package flannel

import (
	"github.com/giantswarm/kvm-operator/service/resource"
	"github.com/giantswarm/kvmtpr"
	apiv1 "k8s.io/client-go/pkg/api/v1"
)

func (s *Service) newContainers(customObject *kvmtpr.CustomObject) []apiv1.Container {
	privileged := true

	return []apiv1.Container{
		{
			Name:            "flannel-client",
			Image:           customObject.Spec.KVM.Flannel.Docker.Image,
			ImagePullPolicy: apiv1.PullAlways,
			Command: []string{
				"/bin/sh",
				"-c",
				"/opt/bin/flanneld --etcd-endpoints https://127.0.0.1:2379 --public-ip=$NODE_IP --iface=$NODE_IP --networks=$NETWORK_BRIDGE_NAME -v=0",
			},
			Env: []apiv1.EnvVar{
				{
					Name:  "NETWORK_BRIDGE_NAME",
					Value: resource.NetworkBridgeName(resource.ClusterID(*customObject)),
				},
				{
					Name: "NODE_IP",
					ValueFrom: &apiv1.EnvVarSource{
						FieldRef: &apiv1.ObjectFieldSelector{
							APIVersion: "v1",
							FieldPath:  "spec.nodeName",
						},
					},
				},
			},
			VolumeMounts: []apiv1.VolumeMount{
				{
					Name:      "flannel",
					MountPath: "/run/flannel",
				},
				{
					Name:      "ssl",
					MountPath: "/etc/ssl/certs",
				},
			},
			SecurityContext: &apiv1.SecurityContext{
				Privileged: &privileged,
			},
		},
		{
			Name:            "k8s-network-bridge",
			Image:           customObject.Spec.KVM.Network.Bridge.Docker.Image,
			ImagePullPolicy: apiv1.PullAlways,
			Command: []string{
				"/bin/sh",
				"-c",
				"while [ ! -f ${NETWORK_ENV_FILE_PATH} ]; do echo \"Waiting for ${NETWORK_ENV_FILE_PATH} to be created\"; sleep 1; done; /docker-entrypoint.sh create ${NETWORK_ENV_FILE_PATH} ${NETWORK_BRIDGE_NAME} ${NETWORK_INTERFACE_NAME} ${HOST_PRIVATE_NETWORK}",
			},
			Env: []apiv1.EnvVar{
				{
					Name:  "NETWORK_ENV_FILE_PATH",
					Value: resource.NetworkEnvFilePath(resource.ClusterID(*customObject)),
				},
				{
					Name:  "HOST_PRIVATE_NETWORK",
					Value: customObject.Spec.KVM.Flannel.PrivateNetwork,
				},
				{
					Name:  "NETWORK_BRIDGE_NAME",
					Value: resource.NetworkBridgeName(resource.ClusterID(*customObject)),
				},
				{
					Name:  "NETWORK_DNS_BLOCK",
					Value: resource.NetworkDNSBlock(customObject.Spec.KVM.DNS.Servers),
				},
				{
					Name:  "NETWORK_NTP_BLOCK",
					Value: resource.NetworkNTPBlock(customObject.Spec.KVM.NTP.Servers),
				},
				{
					Name:  "NETWORK_INTERFACE_NAME",
					Value: customObject.Spec.KVM.Flannel.Interface,
				},
			},
			SecurityContext: &apiv1.SecurityContext{
				Privileged: &privileged,
			},
			VolumeMounts: []apiv1.VolumeMount{
				{
					Name:      "cgroup",
					MountPath: "/sys/fs/cgroup",
				},
				{
					Name:      "dbus",
					MountPath: "/var/run/dbus",
				},
				{
					Name:      "environment",
					MountPath: "/etc/environment",
				},
				{
					Name:      "etc-systemd",
					MountPath: "/etc/systemd/",
				},
				{
					Name:      "flannel",
					MountPath: "/run/flannel",
				},
				{
					Name:      "systemctl",
					MountPath: "/usr/bin/systemctl",
				},
				{
					Name:      "systemd",
					MountPath: "/run/systemd",
				},
				{
					Name:      "sys-class-net",
					MountPath: "/sys/class/net/",
				},
			},
		},
	}
}
