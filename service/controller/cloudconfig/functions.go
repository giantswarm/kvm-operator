package cloudconfig

import (
	"context"
	"net"
	"strings"
	"sync"

	"github.com/giantswarm/apiextensions/v3/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/apiextensions/v3/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/certs/v3/pkg/certs"
	"github.com/giantswarm/microerror"
	"golang.org/x/sync/errgroup"

	"github.com/giantswarm/kvm-operator/service/controller/key"
)

type certFileMapping map[certs.Cert]func(certs.TLS) []certs.File

var masterCertFiles = certFileMapping{
	certs.APICert:              certs.NewFilesAPI,
	certs.EtcdCert:             certs.NewFilesEtcd,
	certs.ServiceAccountCert:   certs.NewFilesServiceAccount,
	certs.CalicoEtcdClientCert: certs.NewFilesCalicoEtcdClient,
}

var workerCertFiles = certFileMapping{
	certs.WorkerCert:           certs.NewFilesWorker,
	certs.CalicoEtcdClientCert: certs.NewFilesCalicoEtcdClient,
}

func fetchCertFiles(ctx context.Context, searcher certs.Interface, clusterID string, mapping certFileMapping) ([]certs.File, error) {
	group, groupCtx := errgroup.WithContext(ctx)

	var mu sync.Mutex
	var certFiles []certs.File
	for cert, fileMapper := range mapping {
		cert, fileMapper := cert, fileMapper
		group.Go(func() error {
			tls, err := searcher.SearchTLS(groupCtx, clusterID, cert)
			if err != nil {
				return microerror.Mask(err)
			}

			mu.Lock()
			certFiles = append(certFiles, fileMapper(tls)...)
			mu.Unlock()

			return nil
		})
	}

	err := group.Wait()
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return certFiles, nil
}

func clusterToLegacy(c Config, cr v1alpha2.KVMCluster, l string) v1alpha1.KVMConfigSpec {
	return v1alpha1.KVMConfigSpec{
		Cluster: v1alpha1.Cluster{
			Calico: v1alpha1.ClusterCalico{
				CIDR:   c.CalicoCIDR,
				MTU:    c.CalicoMTU,
				Subnet: c.CalicoSubnet,
			},
			Docker: v1alpha1.ClusterDocker{
				Daemon: v1alpha1.ClusterDockerDaemon{
					CIDR: c.DockerDaemonCIDR,
				},
			},
			Etcd: v1alpha1.ClusterEtcd{
				Domain: key.ClusterEtcdDomain(cr),
				Prefix: key.EtcdPrefix,
			},
			Kubernetes: v1alpha1.ClusterKubernetes{
				API: v1alpha1.ClusterKubernetesAPI{
					ClusterIPRange: c.ClusterIPRange,
					Domain:         key.ClusterAPIEndpoint(cr),
					SecurePort:     key.KubernetesSecurePort,
				},
				CloudProvider: key.CloudProvider,
				DNS: v1alpha1.ClusterKubernetesDNS{
					IP: dnsIPFromRange(c.ClusterIPRange),
				},
				Domain: c.ClusterDomain,
				Kubelet: v1alpha1.ClusterKubernetesKubelet{
					Domain: key.ClusterKubeletEndpoint(cr),
					Labels: l,
				},
				NetworkSetup: v1alpha1.ClusterKubernetesNetworkSetup{
					Docker: v1alpha1.ClusterKubernetesNetworkSetupDocker{
						Image: c.NetworkSetupDockerImage,
					},
				},
				SSH: v1alpha1.ClusterKubernetesSSH{
					UserList: stringToUserList(c.SSHUserList),
				},
			},
		},
		KVM: v1alpha1.KVMConfigSpecKVM{
			EndpointUpdater: v1alpha1.KVMConfigSpecKVMEndpointUpdater{},
			K8sKVM:          v1alpha1.KVMConfigSpecKVMK8sKVM{},
			Masters:         nil,
			Network:         v1alpha1.KVMConfigSpecKVMNetwork{},
			NodeController:  v1alpha1.KVMConfigSpecKVMNodeController{},
			PortMappings:    nil,
			Workers:         nil,
		},
	}
}

// dnsIPFromRange takes the cluster IP range and returns the Kube DNS IP we use
// internally. It must be some specific IP, so we chose the last IP octet to be
// 10. The only reason to do this is to have some static value we apply
// everywhere.
func dnsIPFromRange(s string) string {
	ip := ipFromString(s)
	ip[3] = 10
	return ip.String()
}

func ipFromString(cidr string) net.IP {
	ip, _, err := net.ParseCIDR(cidr)
	if err != nil {
		panic(err)
	}

	// Only IPV4 CIDRs are supported.
	ip = ip.To4()
	if ip == nil {
		panic("CIDR must be an IPV4 range")
	}

	// IP must be a network address.
	if ip[3] != 0 {
		panic("CIDR address must be a network address")
	}

	return ip
}

func stringToUserList(s string) []v1alpha1.ClusterKubernetesSSHUser {
	var list []v1alpha1.ClusterKubernetesSSHUser

	for _, user := range strings.Split(s, ",") {
		if user == "" {
			continue
		}

		trimmed := strings.TrimSpace(user)
		split := strings.Split(trimmed, ":")

		if len(split) != 2 {
			panic("SSH user format must be <name>:<public key>")
		}

		u := v1alpha1.ClusterKubernetesSSHUser{
			Name:      split[0],
			PublicKey: split[1],
		}

		list = append(list, u)
	}

	return list
}
