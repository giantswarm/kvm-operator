package key

import (
	"fmt"
	"net"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	MasterID = "master"
	WorkerID = "worker"
	EtcdPort = 443
	// livenessPortBase is a baseline for computing the port for liveness probes.
	livenessPortBase = 23000
	// shutdownDeferrerPortBase is a baseline for computing the port for
	// shutdown-deferrer.
	shutdownDeferrerPortBase = 47000
	// HealthEndpoint is http path for liveness probe.
	HealthEndpoint = "/healthz"
	// ProbeHost host for liveness probe.
	ProbeHost = "127.0.0.1"
	// InitialDelaySeconds is InitialDelaySeconds param in liveness probe config
	InitialDelaySeconds = 240
	// TimeoutSeconds is TimeoutSeconds param in liveness probe config
	TimeoutSeconds = 5
	// PeriodSeconds is PeriodSeconds param in liveness probe config
	PeriodSeconds = 35
	// FailureThreshold is FailureThreshold param in liveness probe config
	FailureThreshold = 4
	// SuccessThreshold is SuccessThreshold param in liveness probe config
	SuccessThreshold = 1

	// Enable k8s-kvm-health check for k8s api.
	CheckK8sApi = "true"

	// Environment variable names for Downward API use (shutdown-deferrer).
	EnvKeyMyPodName      = "MY_POD_NAME"
	EnvKeyMyPodNamespace = "MY_POD_NAMESPACE"

	FlannelEnvPathPrefix = "/run/flannel"
	CoreosImageDir       = "/var/lib/coreos-kvm-images"
	CoreosVersion        = "1911.5.0"

	K8SEndpointUpdaterDocker = "quay.io/giantswarm/k8s-endpoint-updater:590479a6228c2c143695a268bda5382b52f7ffe1"
	K8SKVMDockerImage        = "corest/fedora:last-5"
	K8SKVMHealthDocker       = "quay.io/giantswarm/k8s-kvm-health:1b1983e44775736bb0b9d7a08ca0ba888fda6567"
	ShutdownDeferrerDocker   = "quay.io/giantswarm/shutdown-deferrer:b2ffdb2c4ec93fe6bf2d4af8e55c8a4b11253611"

	// constants for calculation qemu memory overhead.
	baseMasterMemoryOverhead     = "1G"
	baseWorkerMemoryOverheadMB   = 512
	baseWorkerOverheadMultiplier = 2
	baseWorkerOverheadModulator  = 12
	workerIOOverhead             = "512M"

	// DefaultDockerDiskSize defines the space used to partition the docker FS
	// within k8s-kvm. Note we use this only for masters, since the value for the
	// workers can be configured at runtime by the user.
	DefaultDockerDiskSize = "50G"
	// DefaultKubeletDiskSize defines the space used to partition the kubelet FS
	// within k8s-kvm. Note we use this only for masters, since the value for the
	// workers can be configured at runtime by the user.
	DefaultKubeletDiskSize = "5G"
	// DefaultOSDiskSize defines the space used to partition the root FS within
	// k8s-kvm.
	DefaultOSDiskSize = "5G"
)

const (
	AnnotationAPIEndpoint       = "kvm-operator.giantswarm.io/api-endpoint"
	AnnotationEtcdDomain        = "giantswarm.io/etcd-domain"
	AnnotationIp                = "endpoint.kvm.giantswarm.io/ip"
	AnnotationService           = "endpoint.kvm.giantswarm.io/service"
	AnnotationPodDrained        = "endpoint.kvm.giantswarm.io/drained"
	AnnotationPrometheusCluster = "giantswarm.io/prometheus-cluster"
	AnnotationVersionBundle     = "kvm-operator.giantswarm.io/version-bundle"

	LabelApp           = "app"
	LabelCluster       = "giantswarm.io/cluster"
	LabelCustomer      = "customer"
	LabelManagedBy     = "giantswarm.io/managed-by"
	LabelOrganization  = "giantswarm.io/organization"
	LabelVersionBundle = "giantswarm.io/version-bundle"

	LegacyLabelCluster = "cluster"
)

const (
	VersionBundleVersionAnnotation = "giantswarm.io/version-bundle-version"
)

const (
	PodWatcherLabel = "kvm-operator.giantswarm.io/pod-watcher"
)

const (
	OperatorName = "kvm-operator"
)

const (
	PodDeletionGracePeriod = 5 * time.Minute
)

func ClusterAPIEndpoint(customObject v1alpha1.KVMConfig) string {
	return customObject.Spec.Cluster.Kubernetes.API.Domain
}

func ClusterAPIEndpointFromPod(pod *corev1.Pod) (string, error) {
	apiEndpoint, ok := pod.GetAnnotations()[AnnotationAPIEndpoint]
	if !ok {
		return "", microerror.Maskf(missingAnnotationError, AnnotationAPIEndpoint)
	}
	if apiEndpoint == "" {
		return "", microerror.Maskf(missingAnnotationError, AnnotationAPIEndpoint)
	}

	return apiEndpoint, nil
}

func ClusterCustomer(customObject v1alpha1.KVMConfig) string {
	return customObject.Spec.Cluster.Customer.ID
}

func ClusterEtcdDomain(customObject v1alpha1.KVMConfig) string {
	return fmt.Sprintf("%s:%d", customObject.Spec.Cluster.Etcd.Domain, EtcdPort)
}

func ClusterID(customObject v1alpha1.KVMConfig) string {
	return customObject.Spec.Cluster.ID
}

func ClusterIDFromPod(pod *corev1.Pod) string {
	l, ok := pod.Labels["cluster"]
	if ok {
		return l
	}

	return "n/a"
}

func ClusterNamespace(customObject v1alpha1.KVMConfig) string {
	return ClusterID(customObject)
}

func ClusterRoleBindingName(customObject v1alpha1.KVMConfig) string {
	return ClusterID(customObject)
}

func ClusterRoleBindingPSPName(customObject v1alpha1.KVMConfig) string {
	return ClusterID(customObject) + "-psp"
}

func ConfigMapName(customObject v1alpha1.KVMConfig, node v1alpha1.ClusterNode, prefix string) string {
	return fmt.Sprintf("%s-%s-%s", prefix, ClusterID(customObject), node.ID)
}

func CPUQuantity(n v1alpha1.KVMConfigSpecKVMNode) (resource.Quantity, error) {
	cpu := strconv.Itoa(n.CPUs)
	q, err := resource.ParseQuantity(cpu)
	if err != nil {
		return resource.Quantity{}, microerror.Mask(err)
	}
	return q, nil
}

func DeploymentName(prefix string, nodeID string) string {
	return fmt.Sprintf("%s-%s", prefix, nodeID)
}

func DockerVolumeSizeFromNode(node v1alpha1.KVMConfigSpecKVMNode) string {
	if node.DockerVolumeSizeGB != 0 {
		return fmt.Sprintf("%dG", node.DockerVolumeSizeGB)
	}

	if node.Disk != 0 {
		return fmt.Sprintf("%sG", strconv.FormatFloat(node.Disk, 'f', 0, 64))
	}

	return DefaultDockerDiskSize
}

func EtcdPVCName(clusterID string, vmNumber string) string {
	return fmt.Sprintf("%s-%s-%s", "pvc-master-etcd", clusterID, vmNumber)
}

func NetworkEnvFilePath(customObject v1alpha1.KVMConfig) string {
	return fmt.Sprintf("%s/networks/%s.env", FlannelEnvPathPrefix, NetworkBridgeName(customObject))
}

func HealthListenAddress(customObject v1alpha1.KVMConfig) string {
	return "http://" + ProbeHost + ":" + strconv.Itoa(int(LivenessPort(customObject)))
}

func IsDeleted(customObject v1alpha1.KVMConfig) bool {
	return customObject.GetDeletionTimestamp() != nil
}

// IsPodDrained checks whether the pod status indicates it got drained. The pod
// status is partially reflected by its annotations. Here we check for the
// annotation that tells us if the pod was already drained or not. In case the
// pod does not have any annotations an unrecoverable error is returned. Such
// situations should actually never happen. If it happens, something really bad
// is going on. This is nothing we can just sort right away in our code.
//
// TODO(xh3b4sd) handle pod status via the runtime object status primitives
// and not via annotations.
func IsPodDrained(pod *corev1.Pod) (bool, error) {
	a := pod.GetAnnotations()
	if a == nil {
		return false, microerror.Mask(missingAnnotationError)
	}
	v, ok := a[AnnotationPodDrained]
	if !ok {
		return false, microerror.Mask(missingAnnotationError)
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return false, microerror.Mask(err)
	}

	return b, nil
}

func KubeletVolumeSizeFromNode(node v1alpha1.KVMConfigSpecKVMNode) string {
	// TODO: https://github.com/giantswarm/giantswarm/issues/4105#issuecomment-421772917
	// TODO: for now we use same value as for DockerVolumeSizeFromNode, when we have kubelet size in spec we should use that.

	if node.DockerVolumeSizeGB != 0 {
		return fmt.Sprintf("%dG", node.DockerVolumeSizeGB)
	}

	if node.Disk != 0 {
		return fmt.Sprintf("%sG", strconv.FormatFloat(node.Disk, 'f', 0, 64))
	}

	return DefaultKubeletDiskSize
}

// ArePodContainersTerminated checks ContainerState for all containers present
// in given pod. When all containers are in Terminated state, true is returned.
func ArePodContainersTerminated(pod *corev1.Pod) bool {
	for _, cs := range pod.Status.ContainerStatuses {
		if cs.State.Terminated == nil {
			return false
		}
	}

	return true
}

func LivenessPort(customObject v1alpha1.KVMConfig) int32 {
	return int32(livenessPortBase + customObject.Spec.KVM.Network.Flannel.VNI)
}

func MasterCount(customObject v1alpha1.KVMConfig) int {
	return len(customObject.Spec.KVM.Masters)
}

func MasterHostPathVolumeDir(clusterID string, vmNumber string) string {
	return filepath.Join("/home/core/volumes", clusterID, "k8s-master-vm"+vmNumber)
}

// MemoryQuantity returns a resource.Quantity that represents the memory to be used by the nodes.
// It adds the memory from the node definition parameter to the additional memory calculated on the node role
func MemoryQuantityMaster(n v1alpha1.KVMConfigSpecKVMNode) (resource.Quantity, error) {
	q, err := resource.ParseQuantity(n.Memory)
	if err != nil {
		return resource.Quantity{}, microerror.Maskf(err, "creating Memory quantity from node definition")
	}
	additionalMemory := resource.MustParse(baseMasterMemoryOverhead)
	if err != nil {
		return resource.Quantity{}, microerror.Maskf(err, "creating Memory quantity from addtional memory")
	}
	q.Add(additionalMemory)

	return q, nil
}

// MemoryQuantity returns a resource.Quantity that represents the memory to be used by the nodes.
// It adds the memory from the node definition parameter to the additional memory calculated on the node role
func MemoryQuantityWorker(n v1alpha1.KVMConfigSpecKVMNode) (resource.Quantity, error) {
	mQuantity, err := resource.ParseQuantity(n.Memory)
	if err != nil {
		return resource.Quantity{}, microerror.Maskf(err, "calculating memory overhead multiplier")
	}

	// base worker memory calculated in MB
	q, err := resource.ParseQuantity(fmt.Sprintf("%dM", mQuantity.ScaledValue(resource.Giga)*1024))
	if err != nil {
		return resource.Quantity{}, microerror.Maskf(err, "creating Memory quantity from node definition")
	}
	// IO overhead for qemu is around 512M memory
	ioOverhead := resource.MustParse(workerIOOverhead)
	q.Add(ioOverhead)

	// memory overhead is more complex as it increases with the size of the memory
	// basic calculation is (2 + (memory / 12))*512M
	// examples:
	// Memory under 15G >> overhead 1024M
	// memory between 15 - 30G >> overhead 1536M
	// memory between 30 - 45G >> overhead 2048M
	overheadMultiplier := int(baseWorkerOverheadMultiplier + mQuantity.ScaledValue(resource.Giga)/baseWorkerOverheadModulator)
	workerMemoryOverhead := strconv.Itoa(baseWorkerMemoryOverheadMB*overheadMultiplier) + "M"

	memOverhead, err := resource.ParseQuantity(workerMemoryOverhead)
	if err != nil {
		return resource.Quantity{}, microerror.Maskf(err, "creating Memory quantity from memory overhead")
	}
	q.Add(memOverhead)

	return q, nil
}

func NetworkBridgeName(customObject v1alpha1.KVMConfig) string {
	return fmt.Sprintf("br-%s", ClusterID(customObject))
}

func NetworkTapName(customObject v1alpha1.KVMConfig) string {
	return fmt.Sprintf("tap-%s", ClusterID(customObject))
}

func NetworkDNSBlock(servers []net.IP) string {
	var dnsBlockParts []string

	for _, s := range servers {
		dnsBlockParts = append(dnsBlockParts, fmt.Sprintf("DNS=%s", s.String()))
	}

	dnsBlock := strings.Join(dnsBlockParts, "\n")

	return dnsBlock
}

func NetworkNTPBlock(servers []net.IP) string {
	var ntpBlockParts []string

	for _, s := range servers {
		ntpBlockParts = append(ntpBlockParts, fmt.Sprintf("NTP=%s", s.String()))
	}

	ntpBlock := strings.Join(ntpBlockParts, "\n")

	return ntpBlock
}

func PortMappings(customObject v1alpha1.KVMConfig) []corev1.ServicePort {
	var ports []corev1.ServicePort

	// Compatibility mode, if no port mappings specified.
	if len(customObject.Spec.KVM.PortMappings) == 0 {
		ports := []corev1.ServicePort{
			{
				Name:       "http",
				Port:       int32(30010),
				TargetPort: intstr.FromInt(30010),
			},
			{
				Name:       "https",
				Port:       int32(30011),
				TargetPort: intstr.FromInt(30011),
			},
		}
		return ports
	}

	for _, p := range customObject.Spec.KVM.PortMappings {
		port := corev1.ServicePort{
			Name:       p.Name,
			NodePort:   int32(p.NodePort),
			Port:       int32(p.TargetPort),
			TargetPort: intstr.FromInt(p.TargetPort),
		}
		ports = append(ports, port)
	}

	return ports
}

func PVCNames(customObject v1alpha1.KVMConfig) []string {
	var names []string

	for i := range customObject.Spec.Cluster.Masters {
		names = append(names, EtcdPVCName(ClusterID(customObject), VMNumber(i)))
	}

	return names
}

func ServiceAccountName(customObject v1alpha1.KVMConfig) string {
	return ClusterID(customObject)
}

func ShutdownDeferrerListenPort(customObject v1alpha1.KVMConfig) int {
	return int(shutdownDeferrerPortBase + customObject.Spec.KVM.Network.Flannel.VNI)
}

func ShutdownDeferrerListenAddress(customObject v1alpha1.KVMConfig) string {
	return "http://" + ProbeHost + ":" + strconv.Itoa(ShutdownDeferrerListenPort(customObject))
}

func ShutdownDeferrerPollPath(customObject v1alpha1.KVMConfig) string {
	return fmt.Sprintf("%s/v1/defer/", ShutdownDeferrerListenAddress(customObject))
}

func StorageType(customObject v1alpha1.KVMConfig) string {
	return customObject.Spec.KVM.K8sKVM.StorageType
}

func ToClusterEndpoint(v interface{}) (string, error) {
	customObject, err := ToCustomObject(v)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return ClusterAPIEndpoint(customObject), nil
}

func ToClusterID(v interface{}) (string, error) {
	customObject, err := ToCustomObject(v)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return ClusterID(customObject), nil
}

func ToClusterStatus(v interface{}) (v1alpha1.StatusCluster, error) {
	customObject, err := ToCustomObject(v)
	if err != nil {
		return v1alpha1.StatusCluster{}, microerror.Mask(err)
	}

	return customObject.Status.Cluster, nil
}

func ToCustomObject(v interface{}) (v1alpha1.KVMConfig, error) {
	customObjectPointer, ok := v.(*v1alpha1.KVMConfig)
	if !ok {
		return v1alpha1.KVMConfig{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &v1alpha1.KVMConfig{}, v)
	}
	customObject := *customObjectPointer

	return customObject, nil
}

func ToNodeCount(v interface{}) (int, error) {
	customObject, err := ToCustomObject(v)
	if err != nil {
		return 0, microerror.Mask(err)
	}

	nodeCount := MasterCount(customObject) + WorkerCount(customObject)

	return nodeCount, nil
}

func ToPod(v interface{}) (*corev1.Pod, error) {
	if v == nil {
		return nil, nil
	}

	pod, ok := v.(*corev1.Pod)
	if !ok {
		return nil, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &corev1.Pod{}, v)
	}

	return pod, nil
}

func ToVersionBundleVersion(v interface{}) (string, error) {
	customObject, err := ToCustomObject(v)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return VersionBundleVersion(customObject), nil
}

func VersionBundleVersion(customObject v1alpha1.KVMConfig) string {
	return customObject.Spec.VersionBundle.Version
}

func VersionBundleVersionFromPod(pod *corev1.Pod) (string, error) {
	a := pod.GetAnnotations()
	if a == nil {
		return "", microerror.Mask(missingAnnotationError)
	}
	v, ok := a[AnnotationVersionBundle]
	if !ok {
		return "", microerror.Mask(missingAnnotationError)
	}

	return v, nil
}

func VMNumber(ID int) string {
	return fmt.Sprintf("%d", ID)
}

func WorkerCount(customObject v1alpha1.KVMConfig) int {
	return len(customObject.Spec.KVM.Workers)
}
