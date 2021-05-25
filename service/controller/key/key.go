package key

import (
	"context"
	"fmt"
	"net"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/giantswarm/apiextensions/v3/pkg/apis/provider/v1alpha1"
	releasev1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/k8sclient/v5/pkg/k8sclient"
	k8scloudconfig "github.com/giantswarm/k8scloudconfig/v10/pkg/template"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/tenantcluster/v4/pkg/tenantcluster"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/kvm-operator/pkg/label"
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
	// LivenessProbeInitialDelaySeconds is LivenessProbeInitialDelaySeconds param in liveness probe config.
	LivenessProbeInitialDelaySeconds = 500
	// ReadinessProbeInitialDelaySeconds is ReadinessProbeInitialDelaySeconds param in readiness probe config.
	ReadinessProbeInitialDelaySeconds = 100
	// TimeoutSeconds is TimeoutSeconds param in liveness probe config.
	TimeoutSeconds = 5
	// PeriodSeconds is PeriodSeconds param in liveness probe config.
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

	ContainerLinuxComponentName = "containerlinux"

	FlatcarImageDir = "/var/lib/flatcar-kvm-images"
	FlatcarChannel  = "stable"

	K8SKVMContainerName = "k8s-kvm"

	K8SEndpointUpdaterDocker = "quay.io/giantswarm/k8s-endpoint-updater:0.1.0"
	K8SKVMDockerImage        = "quay.io/giantswarm/k8s-kvm:0.6.2"
	K8SKVMHealthDocker       = "quay.io/giantswarm/k8s-kvm-health:0.1.0"
	ShutdownDeferrerDocker   = "quay.io/giantswarm/shutdown-deferrer:0.1.0"

	// constants for calculation qemu memory overhead.
	baseMasterMemoryOverhead     = "1024M"
	baseWorkerMemoryOverheadMB   = 768
	baseWorkerOverheadMultiplier = 2
	baseWorkerOverheadModulator  = 12
	qemuMemoryIOOverhead         = "512M"

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

	DefaultImagePullProgressDeadline = "1m"
)

const (
	AnnotationAPIEndpoint            = "kvm-operator.giantswarm.io/api-endpoint"
	AnnotationComponentVersionPrefix = "kvm-operator.giantswarm.io/component-version"
	AnnotationEtcdDomain             = "giantswarm.io/etcd-domain"
	AnnotationIp                     = "endpoint.kvm.giantswarm.io/ip"
	AnnotationService                = "endpoint.kvm.giantswarm.io/service"
	AnnotationPodDrained             = "endpoint.kvm.giantswarm.io/drained"
	AnnotationPrometheusCluster      = "giantswarm.io/prometheus-cluster"
	AnnotationVersionBundle          = "kvm-operator.giantswarm.io/version-bundle"

	LabelApp           = "app"
	LabelCluster       = "giantswarm.io/cluster"
	LabelCustomer      = "customer"
	LabelManagedBy     = "giantswarm.io/managed-by"
	LabelOrganization  = "giantswarm.io/organization"
	LabelVersionBundle = "giantswarm.io/version-bundle"

	LegacyLabelCluster = "cluster"
)

const (
	KubernetesNetworkSetupDocker = "0.2.0"
	kubernetesAPIHealthzVersion  = "0.1.1"
)

const (
	VersionBundleVersionAnnotation = "giantswarm.io/version-bundle-version"
	ReleaseVersionAnnotation       = "giantswarm.io/release-version"
)

const (
	PodWatcherLabel = "kvm-operator.giantswarm.io/pod-watcher"
)

const (
	OperatorName = "kvm-operator"
)

const (
	// Keep in sync with https://github.com/giantswarm/node-operator/blob/e8f8250f7d518d7af0acfdb1b0934a1174c22d7d/service/controller/v2/resource/drainer/create.go#L50
	PodDeletionGracePeriod                           = 10 * time.Minute
	WorkloadClusterNodeReady corev1.PodConditionType = "kvm-operator.giantswarm.io/workload-cluster-node-ready"
)

func AllNodes(cr v1alpha1.KVMConfig) []v1alpha1.ClusterNode {
	var results []v1alpha1.ClusterNode

	results = append(results, cr.Spec.Cluster.Masters...)

	results = append(results, cr.Spec.Cluster.Workers...)

	return results
}

func AllocatedNodeIndexes(cr v1alpha1.KVMConfig) []int {
	var results []int

	for _, v := range cr.Status.KVM.NodeIndexes {
		results = append(results, v)
	}

	sort.Ints(results)

	return results
}

func BaseDomain(customObject v1alpha1.KVMConfig) string {
	return strings.TrimPrefix(customObject.Spec.Cluster.Kubernetes.API.Domain, "api.")
}

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

func ClusterNamespace(customObject v1alpha1.KVMConfig) string {
	return ClusterID(customObject)
}

func ClusterRoleBindingName(customObject v1alpha1.KVMConfig) string {
	return ClusterID(customObject)
}

func ClusterRoleBindingPSPName(customObject v1alpha1.KVMConfig) string {
	return ClusterID(customObject) + "-psp"
}

func ConfigMapName(cr v1alpha1.KVMConfig, node v1alpha1.ClusterNode, prefix string) string {
	return fmt.Sprintf("%s-%s-%s", prefix, ClusterID(cr), node.ID)
}

func ContainerDistro(release *releasev1alpha1.Release) (string, error) {
	for _, component := range release.Spec.Components {
		if component.Name == ContainerLinuxComponentName {
			return component.Version, nil
		}
	}

	return "", microerror.Mask(missingVersionError)
}

func CPUQuantity(n v1alpha1.KVMConfigSpecKVMNode) (resource.Quantity, error) {
	cpu := strconv.Itoa(n.CPUs)
	q, err := resource.ParseQuantity(cpu)
	if err != nil {
		return resource.Quantity{}, microerror.Mask(err)
	}
	return q, nil
}

// CreateK8sClientForWorkloadCluster takes the context of the reconciled object
// and the provided logger and tenant cluster interface and creates a K8s client for the workload cluster
func CreateK8sClientForWorkloadCluster(ctx context.Context, cluster v1alpha1.KVMConfig, logger micrologger.Logger, workloadCluster tenantcluster.Interface) (*k8sclient.Clients, error) {
	i := ClusterID(cluster)
	e := ClusterAPIEndpoint(cluster)

	restConfig, err := workloadCluster.NewRestConfig(ctx, i, e)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	clientsConfig := k8sclient.ClientsConfig{
		Logger:     logger,
		RestConfig: restConfig,
	}
	k8sClients, err := k8sclient.NewClients(clientsConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return k8sClients, nil
}

func DeploymentName(prefix string, nodeID string) string {
	return fmt.Sprintf("%s-%s", prefix, nodeID)
}

func DefaultVersions() k8scloudconfig.Versions {
	return k8scloudconfig.Versions{
		KubernetesAPIHealthz:         kubernetesAPIHealthzVersion,
		KubernetesNetworkSetupDocker: KubernetesNetworkSetupDocker,
	}
}

func DockerVolumeSizeFromNode(node v1alpha1.KVMConfigSpecKVMNode) string {
	if node.DockerVolumeSizeGB != 0 {
		return fmt.Sprintf("%dG", node.DockerVolumeSizeGB)
	}

	if node.Disk.Value != 0 {
		return fmt.Sprintf("%sG", strconv.FormatFloat(node.Disk.Value, 'f', 0, 64))
	}

	return DefaultDockerDiskSize
}

func EtcdPVCName(clusterID string, vmNumber string) string {
	return fmt.Sprintf("%s-%s-%s", "pvc-master-etcd", clusterID, vmNumber)
}

func LocalPVCName(clusterID string, mountTag string) string {
	return fmt.Sprintf("%s-%s-%s", "local-pv", clusterID, mountTag)
}

// FindNodeCondition returns the condition of the given type from the node. The second return value indicates if the condition was found.
func FindNodeCondition(node corev1.Node, conditionType corev1.NodeConditionType) (corev1.NodeCondition, bool) {
	for _, condition := range node.Status.Conditions {
		if condition.Type == conditionType {
			return condition, true
		}
	}
	return corev1.NodeCondition{}, false
}

// FindNodeCondition returns the condition of the given type from the pod. The second return value indicates if the condition was found.
func FindPodCondition(pod corev1.Pod, conditionType corev1.PodConditionType) (corev1.PodCondition, bool) {
	for _, condition := range pod.Status.Conditions {
		if condition.Type == conditionType {
			return condition, true
		}
	}
	return corev1.PodCondition{}, false
}

func HealthListenAddress(customObject v1alpha1.KVMConfig) string {
	return "http://" + ProbeHost + ":" + strconv.Itoa(int(LivenessPort(customObject)))
}

func IscsiInitiatorName(customObject v1alpha1.KVMConfig, nodeIndex int, nodeRole string) string {
	return fmt.Sprintf("iqn.2016-04.com.coreos.iscsi:giantswarm-%s-%s-%d", ClusterID(customObject), nodeRole, nodeIndex)
}

func IsDeleted(object v1.Object) bool {
	return object.GetDeletionTimestamp() != nil
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
func IsPodDrained(pod corev1.Pod) (bool, error) {
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

	if node.Disk.Value != 0 {
		return fmt.Sprintf("%sG", strconv.FormatFloat(node.Disk.Value, 'f', 0, 64))
	}

	return DefaultKubeletDiskSize
}

func HostVolumesToEnvVar(hostVolumes []v1alpha1.KVMConfigSpecKVMNodeHostVolumes) corev1.EnvVar {
	var lastElemIndex = len(hostVolumes) - 1

	hostVolumesEnvVar := corev1.EnvVar{
		Name:  "HOST_DATA_VOLUME_PATHS",
		Value: "",
	}

	for idx, hostVolume := range hostVolumes {
		hostVolumesEnvVar.Value += fmt.Sprintf("%s:%s", hostVolume.MountTag, hostVolume.HostPath)

		if idx != lastElemIndex {
			hostVolumesEnvVar.Value += ","
		}
	}

	return hostVolumesEnvVar
}

func HostVolumesToVolumeMounts(hostVolumes []v1alpha1.KVMConfigSpecKVMNodeHostVolumes) []corev1.VolumeMount {
	var volumeMounts []corev1.VolumeMount

	for _, hostVolume := range hostVolumes {
		vm := corev1.VolumeMount{
			Name:      hostVolume.MountTag,
			MountPath: hostVolume.HostPath,
		}

		volumeMounts = append(volumeMounts, vm)
	}

	return volumeMounts
}

func HostVolumesToVolumes(hostVolumes []v1alpha1.KVMConfigSpecKVMNodeHostVolumes) []corev1.Volume {
	var volumes []corev1.Volume

	for _, hostVolume := range hostVolumes {
		v := corev1.Volume{
			Name: hostVolume.MountTag,
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: hostVolume.HostPath,
				},
			},
		}

		volumes = append(volumes, v)
	}

	return volumes
}

// AnyPodContainerRunning checks ContainerState for all containers present
// in given pod. If any container is in Running state, true is returned.
func AnyPodContainerRunning(pod corev1.Pod) bool {
	for _, cs := range pod.Status.ContainerStatuses {
		if cs.State.Running != nil {
			return true
		}
	}

	for _, cs := range pod.Status.InitContainerStatuses {
		if cs.State.Running != nil {
			return true
		}
	}

	for _, cs := range pod.Status.EphemeralContainerStatuses {
		if cs.State.Running != nil {
			return true
		}
	}

	return false
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
		return resource.Quantity{}, microerror.Maskf(invalidMemoryConfigurationError, "error creating Memory quantity from node definition: %s", err)
	}
	additionalMemory := resource.MustParse(baseMasterMemoryOverhead)
	q.Add(additionalMemory)

	// IO overhead for qemu is around 512M memory
	ioOverhead := resource.MustParse(qemuMemoryIOOverhead)
	q.Add(ioOverhead)

	return q, nil
}

// MemoryQuantity returns a resource.Quantity that represents the memory to be used by the nodes.
// It adds the memory from the node definition parameter to the additional memory calculated on the node role
func MemoryQuantityWorker(n v1alpha1.KVMConfigSpecKVMNode) (resource.Quantity, error) {
	mQuantity, err := resource.ParseQuantity(n.Memory)
	if err != nil {
		return resource.Quantity{}, microerror.Maskf(invalidMemoryConfigurationError, "error calculating memory overhead multiplier: %s", err)
	}

	// base worker memory calculated in MB
	q, err := resource.ParseQuantity(fmt.Sprintf("%dM", mQuantity.ScaledValue(resource.Giga)*1024))
	if err != nil {
		return resource.Quantity{}, microerror.Maskf(invalidMemoryConfigurationError, "error creating Memory quantity from node definition: %s", err)
	}
	// IO overhead for qemu is around 512M memory
	ioOverhead := resource.MustParse(qemuMemoryIOOverhead)
	q.Add(ioOverhead)

	// memory overhead is more complex as it increases with the size of the memory
	// basic calculation is (2 + (memory / 12))*768M
	// examples:
	// Memory under 12G >> overhead 1536M
	// memory between 12 - 24G >> overhead 2048M
	// memory between 24 - 36G >> overhead 2816M
	// memory between 24 - 36G >> overhead 3584M
	// memory between 36 - 48G >> overhead 4352M
	// memory between 48 - 60G >> overhead 5120M
	// memory between 60 - 72G >> overhead 5888M
	// memory between 72 - 84G >> overhead 6656M
	// memory between 84 - 96G >> overhead 7424M
	// memory between 96 - 108G >> overhead 8192M
	// memory between 108 - 120G >> overhead 8960M

	overheadMultiplier := int(baseWorkerOverheadMultiplier + mQuantity.ScaledValue(resource.Giga)/baseWorkerOverheadModulator)
	workerMemoryOverhead := strconv.Itoa(baseWorkerMemoryOverheadMB*overheadMultiplier) + "M"

	memOverhead, err := resource.ParseQuantity(workerMemoryOverhead)
	if err != nil {
		return resource.Quantity{}, microerror.Maskf(invalidMemoryConfigurationError, "error creating Memory quantity from memory overhead: %s", err)
	}
	q.Add(memOverhead)

	return q, nil
}

func NetworkBridgeName(customObject v1alpha1.KVMConfig) string {
	return fmt.Sprintf("br-%s", ClusterID(customObject))
}

func NetworkEnvFilePath(customObject v1alpha1.KVMConfig) string {
	return fmt.Sprintf("%s/networks/%s.env", FlannelEnvPathPrefix, NetworkBridgeName(customObject))
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

// NodeIsReady examines the Status Conditions of a Node
// and returns true if the NodeReady Condition is true.
func NodeIsReady(node corev1.Node) bool {
	condition, _ := FindNodeCondition(node, corev1.NodeReady)
	return condition.Status == corev1.ConditionTrue
}

// NodeIsUnschedulable examines a Node and returns true if the Node is marked Unschedulable or has a NoSchedule/NoExecute taint.
// Ignores the default NoSchedule effect for master nodes.
func NodeIsUnschedulable(node corev1.Node) bool {
	if node.Spec.Unschedulable {
		return true
	}

	for _, t := range node.Spec.Taints {
		if (t.Effect == corev1.TaintEffectNoSchedule && t.Key != "node-role.kubernetes.io/master") ||
			t.Effect == corev1.TaintEffectNoExecute {
			return true
		}
	}
	return false
}

func NodeIndex(cr v1alpha1.KVMConfig, nodeID string) (int, bool) {
	idx, present := cr.Status.KVM.NodeIndexes[nodeID]
	return idx, present
}

// NodeInternalIP examines the Status Addresses of a Node
// and returns its InternalIP..
func NodeInternalIP(node corev1.Node) (string, error) {
	for _, a := range node.Status.Addresses {
		if a.Type == corev1.NodeInternalIP {
			return a.Address, nil
		}
	}
	return "", microerror.Maskf(missingNodeInternalIP, "node %s does not have an InternalIP address in its status", node.Name)
}

func NodePodObjectKey(cluster v1alpha1.KVMConfig, node corev1.Node) client.ObjectKey {
	return client.ObjectKey{
		Name:      node.Name,
		Namespace: ClusterNamespace(cluster),
	}
}

func OperatorVersion(cr v1alpha1.KVMConfig) string {
	return cr.GetLabels()[label.OperatorVersion]
}

// PodIsReady examines the Status Conditions of a Pod
// and returns true if the PodReady Condition is true.
func PodIsReady(pod corev1.Pod) bool {
	condition, _ := FindPodCondition(pod, corev1.PodReady)
	return condition.Status == corev1.ConditionTrue
}

// PodNodeIsReady examines the Status Conditions of a Pod
// and returns true if the WorkloadClusterNodeReady Condition is true.
func PodNodeIsReady(pod corev1.Pod) bool {
	condition, _ := FindPodCondition(pod, WorkloadClusterNodeReady)
	return condition.Status == corev1.ConditionTrue
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

func ReleaseVersion(cr v1alpha1.KVMConfig) string {
	return cr.GetLabels()[label.ReleaseVersion]
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

func ToNode(v interface{}) (corev1.Node, error) {
	if v == nil {
		return corev1.Node{}, nil
	}

	node, ok := v.(*corev1.Node)
	if !ok {
		return corev1.Node{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", corev1.Node{}, v)
	}

	return *node.DeepCopy(), nil
}

func ToNodeCount(v interface{}) (int, error) {
	customObject, err := ToCustomObject(v)
	if err != nil {
		return 0, microerror.Mask(err)
	}

	nodeCount := MasterCount(customObject) + WorkerCount(customObject)

	return nodeCount, nil
}

func ToOperatorVersion(v interface{}) (string, error) {
	customObject, err := ToCustomObject(v)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return OperatorVersion(customObject), nil
}

func ToPod(v interface{}) (corev1.Pod, error) {
	if v == nil {
		return corev1.Pod{}, nil
	}

	pod, ok := v.(*corev1.Pod)
	if !ok {
		return corev1.Pod{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &corev1.Pod{}, v)
	}

	return *pod.DeepCopy(), nil
}

func VMNumber(ID int) string {
	return fmt.Sprintf("%d", ID)
}

func WorkerCount(customObject v1alpha1.KVMConfig) int {
	return len(customObject.Spec.KVM.Workers)
}
