package resources

import (
	"encoding/json"
	"path/filepath"

	"github.com/giantswarm/clusterspec"

	"k8s.io/client-go/pkg/api"
	apiunversioned "k8s.io/client-go/pkg/api/unversioned"
	apiv1 "k8s.io/client-go/pkg/api/v1"
	extensionsv1 "k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/pkg/runtime"
	"k8s.io/client-go/pkg/util/intstr"
)

type IngressController interface {
	ClusterObj
}

type ingressController struct {
	clusterspec.Cluster
}

const ingressControllerReplicas int32 = 2

func (i *ingressController) generateIngressControllerPodAffinity() (string, error) {
	podAffinity := &api.Affinity{
		PodAntiAffinity: &api.PodAntiAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: []api.PodAffinityTerm{
				{
					LabelSelector: &apiunversioned.LabelSelector{
						MatchExpressions: []apiunversioned.LabelSelectorRequirement{
							{
								Key:      "role",
								Operator: apiunversioned.LabelSelectorOpIn,
								Values:   []string{i.Name + "-ingress-controller"},
							},
						},
					},
					TopologyKey: "kubernetes.io/hostname",
				},
			},
		},
		PodAffinity: &api.PodAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: []api.PodAffinityTerm{
				{
					LabelSelector: &apiunversioned.LabelSelector{
						MatchExpressions: []apiunversioned.LabelSelectorRequirement{
							{
								Key:      "role",
								Operator: apiunversioned.LabelSelectorOpIn,
								Values:   []string{i.Name + "-flannel-client"},
							},
						},
					},
					TopologyKey: "kubernetes.io/hostname",
				},
			},
		},
	}

	bytesPodAffinity, err := json.Marshal(podAffinity)
	if err != nil {
		return "", maskAny(err)
	}

	return string(bytesPodAffinity), nil
}

func (i *ingressController) generateInitIngressControllerContainers() (string, error) {
	privileged := true

	initContainers := []apiv1.Container{
		{
			Name:  "ingress-controller-api-certs",
			Image: "leaseweb-registry.private.giantswarm.io/giantswarm/certctl:" + i.Spec.CertctlVersion,
			ImagePullPolicy: apiv1.PullAlways,
			Command: []string{
				"/bin/sh",
				"-c",
				"/opt/certctl issue --vault-addr=$VAULT_ADDR --vault-token=$VAULT_TOKEN --cluster-id=$CLUSTER_ID --common-name=api.$CUSTOMER_ID.g8s.fra-1.giantswarm.io --ttl=720h --crt-file=/etc/kubernetes/ssl/ingress.pem --key-file=/etc/kubernetes/ssl/ingress-key.pem --ca-file=/etc/kubernetes/ssl/ingress-ca.pem",
			},
			VolumeMounts: []apiv1.VolumeMount{
				{
					Name:      "ssl",
					MountPath: "/etc/ssl/certs/ca-certificates.crt",
				},
				{
					Name:      "ingress-certs",
					MountPath: "/etc/kubernetes/ssl/",
				},
			},
			SecurityContext: &apiv1.SecurityContext{
				Privileged: &privileged,
			},
			Env: []apiv1.EnvVar{
				{
					Name:  "CUSTOMER_ID",
					Value: i.Spec.Customer,
				},
				{
					Name:  "CLUSTER_ID",
					Value: i.Spec.ClusterId,
				},
				{
					Name:  "VAULT_TOKEN",
					Value: i.Spec.Certificates.VaultToken,
				},
				{
					Name:  "VAULT_ADDR",
					Value: i.Spec.GiantnetesConfiguration.VaultAddr,
				},
			},
		},
	}

	bytes, err := json.Marshal(initContainers)
	if err != nil {
		return "", maskAny(err)
	}

	return string(bytes), nil
}

func (i *ingressController) GenerateService() (*apiv1.Service, error) {
	ingressPort := 3080
	pingPort := 80

	service := &apiv1.Service{
		TypeMeta: apiunversioned.TypeMeta{
			Kind:       "service",
			APIVersion: "v1",
		},
		ObjectMeta: apiv1.ObjectMeta{
			Name: i.Spec.ClusterId + "-ingress-controller",
			Labels: map[string]string{
				"cluster-id": i.Spec.ClusterId,
				"app":        "k8s-cluster",
			},
		},
		Spec: apiv1.ServiceSpec{
			Type: apiv1.ServiceType("NodePort"),
			Ports: []apiv1.ServicePort{
				{
					Name:       "http-health",
					Port:       int32(pingPort),
					NodePort:   int32(30100),
					Protocol:   "TCP",
					TargetPort: intstr.FromInt(80),
				},
				{
					Name:       "http-haproxy",
					Port:       int32(ingressPort),
					Protocol:   "TCP",
					TargetPort: intstr.FromInt(3080),
					NodePort:   int32(30101),
				},
			},
			Selector: map[string]string{
				"cluster-id": i.Spec.ClusterId,
				"app":        "k8s-cluster",
				"role":       "ingress-controller",
			},
		},
	}

	return service, nil

}

func (i *ingressController) GenerateResources() ([]runtime.Object, error) {
	objects := []runtime.Object{}

	deployment, err := i.GenerateDeployment()
	if err != nil {
		return nil, maskAny(err)
	}

	service, err := i.GenerateService()
	if err != nil {
		return nil, maskAny(err)
	}

	objects = append(objects, deployment)
	objects = append(objects, service)

	return objects, nil
}

func (i *ingressController) GenerateDeployment() (*extensionsv1.Deployment, error) {
	privileged := true

	initContainers, err := i.generateInitIngressControllerContainers()
	if err != nil {
		return nil, maskAny(err)
	}

	podAffinity, err := i.generateIngressControllerPodAffinity()
	if err != nil {
		return nil, maskAny(err)
	}

	replicas := int32(ingressControllerReplicas)

	deployment := &extensionsv1.Deployment{
		TypeMeta: apiunversioned.TypeMeta{
			Kind:       "deployment",
			APIVersion: "extensions/v1beta",
		},
		ObjectMeta: apiv1.ObjectMeta{
			Name: i.Spec.ClusterId + "-ingress-controller",
			Labels: map[string]string{
				"cluster-id": i.Spec.ClusterId,
				"app":        "k8s-cluster",
				"role":       "ingress-controller",
			},
		},
		Spec: extensionsv1.DeploymentSpec{
			Strategy: extensionsv1.DeploymentStrategy{
				Type: "RollingUpdate",
			},
			Replicas: &replicas,
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: apiv1.ObjectMeta{
					Name: i.Spec.ClusterId + "-ingress-controller",
					Labels: map[string]string{
						"cluster-id": i.Spec.ClusterId,
						"app":        "k8s-cluster",
						"role":       "ingress-controller",
					},
					Annotations: map[string]string{
						"pod.beta.kubernetes.io/init-containers": initContainers,
						"scheduler.alpha.kubernetes.io/affinity": podAffinity,
					},
				},
				Spec: apiv1.PodSpec{
					HostNetwork: true,
					Volumes: []apiv1.Volume{
						{
							Name: "haproxy",
							VolumeSource: apiv1.VolumeSource{
								HostPath: &apiv1.HostPathVolumeSource{
									Path: filepath.Join("/var/run/", i.Spec.ClusterId, "/haproxy/"),
								},
							},
						},
						{
							Name: "ingress-certs",
							VolumeSource: apiv1.VolumeSource{
								HostPath: &apiv1.HostPathVolumeSource{
									Path: filepath.Join("/etc/kubernetes/", i.Spec.ClusterId, "/", i.Spec.ClusterId, "/ssl/ingress/"),
								},
							},
						},
						{
							Name: "ssl",
							VolumeSource: apiv1.VolumeSource{
								HostPath: &apiv1.HostPathVolumeSource{
									Path: "/etc/ssl/certs/ca-certificates.crt",
								},
							},
						},
					},
					Containers: []apiv1.Container{
						{
							Name:  "ingress-controller-health",
							Image: "leaseweb-registry.private.giantswarm.io/giantswarm/ping:" + i.Spec.PingVersion,
							ImagePullPolicy: apiv1.PullAlways,
							Args: []string{
								"--healthcheck",
								"--cloudflare",
								"--cloudflare-email=$(CLOUDFLARE_EMAIL)",
								"--cloudflare-token=$(CLOUDFLARE_TOKEN)",
								"--cloudflare-ip=$(CLOUDFLARE_IP)",
								"--cloudflare-domain=$(CLOUDFLARE_DOMAIN)",
								"--cloudflare-subdomain=*.$(CUSTOMER_ID).fra-1",
								"--kemp",
								"--kemp-endpoint=$(KEMP_ENDPOINT)",
								"--kemp-password=$(KEMP_PASSWORD)",
								"--kemp-rs-ip=$(HOST_PUBLIC_IP)",
								"--kemp-rs-port=$(KEMP_RS_PORT)",
								"--kemp-rs-port-unique",
								"--kemp-user=$(KEMP_USER)",
								"--kemp-vs-check-port=$(KEMP_VS_CHECK_PORT)",
								"--kemp-vs-ip=$(KEMP_VS_IP)",
								"--kemp-vs-ports=$(KEMP_VS_PORTS)",
								"--kemp-vs-name=$(KEMP_VS_NAME)",
								"--kemp-vs-ssl-acceleration=$(KEMP_VS_SSL_ACCELERATION)",
							},
							Ports: []apiv1.ContainerPort{
								{
									ContainerPort: int32(80),
								},
							},
							Env: []apiv1.EnvVar{
								{
									Name:  "KEMP_PASSWORD",
									Value: i.Spec.IngressController.KempPassword,
								},
								{
									Name:  "KEMP_USER",
									Value: i.Spec.IngressController.KempUser,
								},
								{
									Name:  "KEMP_ENDPOINT",
									Value: i.Spec.IngressController.KempEndpoint,
								},
								{
									Name:  "KEMP_VS_IP",
									Value: i.Spec.IngressController.KempVsIp,
								},
								{
									Name:  "KEMP_VS_PORTS",
									Value: i.Spec.IngressController.KempVsPorts,
								},
								{
									Name:  "KEMP_VS_NAME",
									Value: i.Spec.IngressController.KempVsName,
								},
								{
									Name:  "KEMP_RS_PORT",
									Value: i.Spec.IngressController.KempRsPort,
								},
								{
									Name:  "KEMP_VS_CHECK_PORT",
									Value: i.Spec.IngressController.KempVsCheckPort,
								},
								{
									Name:  "KEMP_VS_SSL_ACCELERATION",
									Value: i.Spec.IngressController.KempVsSslAcceleration,
								},
								{
									Name:  "CLOUDFLARE_EMAIL",
									Value: i.Spec.IngressController.CloudflareEmail,
								},
								{
									Name:  "CLOUDFLARE_TOKEN",
									Value: i.Spec.IngressController.CloudflareToken,
								},
								{
									Name:  "CLOUDFLARE_IP",
									Value: i.Spec.IngressController.CloudflareIp,
								},
								{
									Name:  "CLOUDFLARE_DOMAIN",
									Value: i.Spec.IngressController.CloudflareDomain,
								},
								{
									Name: "HOST_PUBLIC_IP",
									ValueFrom: &apiv1.EnvVarSource{
										FieldRef: &apiv1.ObjectFieldSelector{
											APIVersion: "v1",
											FieldPath:  "spec.nodeName",
										},
									},
								},
								{
									Name:  "CUSTOMER_ID",
									Value: i.Spec.Customer,
								},
							},
						},
						{
							Name:  "ingress-controller-haproxy",
							Image: "leaseweb-registry.private.giantswarm.io/giantswarm/haproxy-etcd-lb:" + i.Spec.IngressControllerVersion,
							ImagePullPolicy: apiv1.PullAlways,
							Args: []string{
								"--backend-store=kubernetes-ingress",
								"--chained-proxies=false",
								"--poll-period=5",
								"--loop-mode=poll",
								"--stats-socket-path=/var/run/haproxy/admin-public.sock",
								"--tls-backends=false",
								"--kubernetes-api=$(K8S_MASTER_DOMAIN_NAME)",
								"--kubernetes-cert-file=/etc/kubernetes/ssl/ingress.pem",
								"--kubernetes-key-file=/etc/kubernetes/ssl/ingress-key.pem",
								"--kubernetes-ca-file=/etc/kubernetes/ssl/ingress-ca.pem",
								"--kubernetes-cluster-id=$(CUSTOMER_ID)",
								"--kubernetes-base-domain=$(KUBERNETES_BASE_DOMAIN)",
								"--v=3",
							},
							Ports: []apiv1.ContainerPort{
								{
									ContainerPort: int32(3080),
								},
							},
							Env: []apiv1.EnvVar{
								{
									Name:  "KUBERNETES_BASE_DOMAIN",
									Value: i.Spec.GiantnetesConfiguration.CloudflareDomain,
								},
								{
									Name:  "K8S_MASTER_DOMAIN_NAME",
									Value: i.Spec.Worker.MasterDomainName,
								},
								{
									Name:  "CUSTOMER_ID",
									Value: i.Spec.Customer,
								},
							},
							VolumeMounts: []apiv1.VolumeMount{
								{
									Name:      "ssl",
									MountPath: "/etc/ssl/certs/ca-certificates.crt",
								},
								{
									Name:      "ingress-certs",
									MountPath: "/etc/kubernetes/ssl/",
								},
								{
									Name:      "haproxy",
									MountPath: "/var/run/haproxy",
								},
							},
							SecurityContext: &apiv1.SecurityContext{
								Privileged: &privileged,
							},
						},
					},
				},
			},
		},
	}

	return deployment, nil
}
