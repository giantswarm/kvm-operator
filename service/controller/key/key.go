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

	"github.com/giantswarm/apiextensions/v3/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/apiextensions/v3/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/k8sclient/v5/pkg/k8sclient"
	k8scloudconfig "github.com/giantswarm/k8scloudconfig/v10/pkg/template"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/tenantcluster/v4/pkg/tenantcluster"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"

	releasev1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/release/v1alpha1"

	"github.com/giantswarm/kvm-operator/pkg/label"
)

const (
	MasterID = "master"
	WorkerID = "worker"
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

	ContainerLinuxComponentName = "containerlinux"

	FlatcarImageDir = "/var/lib/flatcar-kvm-images"
	FlatcarChannel  = "stable"

	K8SEndpointUpdaterDocker = "quay.io/giantswarm/k8s-endpoint-updater:0.1.0"
	K8SKVMDockerImage        = "quay.io/giantswarm/k8s-kvm:0.3.0-20c5e62ea9f5969f66113e44ffb32bd18db6f4bb"
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

	LabelApp          = "app"
	LabelCluster      = "giantswarm.io/cluster"
	LabelCustomer     = "customer"
	LabelManagedBy    = "giantswarm.io/managed-by"
	LabelOrganization = "giantswarm.io/organization"

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
	PodDeletionGracePeriod = 5 * time.Minute
)

const (
	EtcdPort                     = 2379
	EtcdPrefix                   = "giantswarm.io"
	KubernetesSecurePort         = 443
	KubernetesApiHealthCheckPort = 8089
	CloudProvider                = ""
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

func BaseDomain(cr v1alpha2.KVMCluster) string {
	return cr.Spec.Cluster.DNS.Domain
}

func ClusterAPIEndpoint(cr v1alpha2.KVMCluster) string {
	return fmt.Sprintf("api.%s", BaseDomain(cr))
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

func ClusterCustomer(getter LabelsGetter) string {
	return getter.GetLabels()[label.Organization]
}

func ClusterEtcdDomain(cr v1alpha2.KVMCluster) string {
	return fmt.Sprintf("%s:%d", BaseDomain(cr), EtcdPort)
}

func ClusterID(getter LabelsGetter) string {
	return getter.GetLabels()[label.Cluster]
}

func ClusterNamespace(getter LabelsGetter) string {
	return ClusterID(getter)
}

func ClusterKubeletEndpoint(cluster v1alpha2.KVMCluster) string {
	return fmt.Sprintf("worker.%s", BaseDomain(cluster))
}

func ClusterRoleBindingName(getter LabelsGetter) string {
	return ClusterID(getter)
}

func ClusterRoleBindingPSPName(getter LabelsGetter) string {
	return ClusterID(getter) + "-psp"
}

func ConfigMapName(machine v1alpha2.KVMMachine, prefix string) string {
	return fmt.Sprintf("%s-%s-%s", Role(&machine), ClusterID(&machine), machine.Spec.ProviderID)
}

func Role(getter LabelsGetter) string {
	return getter.GetLabels()[label.Cluster]
}

func ContainerDistro(release releasev1alpha1.Release) (string, error) {
	for _, component := range release.Spec.Components {
		if component.Name == ContainerLinuxComponentName {
			return component.Version, nil
		}
	}

	return "", microerror.Mask(missingVersionError)
}

func CPUQuantity(n v1alpha2.KVMMachineSpecSize) (resource.Quantity, error) {
	cpu := strconv.Itoa(n.CPUs)
	q, err := resource.ParseQuantity(cpu)
	if err != nil {
		return resource.Quantity{}, microerror.Mask(err)
	}
	return q, nil
}

// CreateK8sClientForTenantCluster takes the context of the reconciled object
// and the provided logger and tenant cluster interface and creates a K8s client for the tenant cluster
func CreateK8sClientForTenantCluster(ctx context.Context, cr v1alpha2.KVMCluster, logger micrologger.Logger, tenantCluster tenantcluster.Interface) (kubernetes.Interface, error) {

	// Create a client for the reconciled tenant cluster
	var tcK8sClient kubernetes.Interface
	{
		i := ClusterID(&cr)
		e := ClusterAPIEndpoint(cr)

		restConfig, err := tenantCluster.NewRestConfig(ctx, i, e)
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

		tcK8sClient = k8sClients.K8sClient()
	}

	return tcK8sClient, nil
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

func EtcdPVCName(clusterID string, vmNumber string) string {
	return fmt.Sprintf("%s-%s-%s", "pvc-master-etcd", clusterID, vmNumber)
}

func IscsiInitiatorName(cr v1alpha2.KVMCluster, nodeIndex int, nodeRole string) string {
	return fmt.Sprintf("iqn.2016-04.com.coreos.iscsi:giantswarm-%s-%s-%d", ClusterID(&cr), nodeRole, nodeIndex)
}

func IsDeleted(cr v1.Object) bool {
	return cr.GetDeletionTimestamp() != nil
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

func LivenessPort(cr v1alpha2.KVMCluster) int32 {
	return int32(livenessPortBase + cr.Spec.Provider.FlannelVNI)
}

func MasterCount(cr v1alpha2.KVMCluster) int {
	return 1
}

func OIDCClientID(cluster v1alpha2.KVMCluster) string {
	return cluster.Spec.Cluster.OIDC.ClientID
}
func OIDCIssuerURL(cluster v1alpha2.KVMCluster) string {
	return cluster.Spec.Cluster.OIDC.IssuerURL
}
func OIDCUsernameClaim(cluster v1alpha2.KVMCluster) string {
	return cluster.Spec.Cluster.OIDC.Claims.Username
}
func OIDCGroupsClaim(cluster v1alpha2.KVMCluster) string {
	return cluster.Spec.Cluster.OIDC.Claims.Groups
}

func MasterHostPathVolumeDir(clusterID string, vmNumber string) string {
	return filepath.Join("/home/core/volumes", clusterID, "k8s-master-vm"+vmNumber)
}

// MemoryQuantity returns a resource.Quantity that represents the memory to be used by the nodes.
// It adds the memory from the node definition parameter to the additional memory calculated on the node role
func MemoryQuantityMaster(n v1alpha2.KVMMachineSpecSize) (resource.Quantity, error) {
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
func MemoryQuantityWorker(n v1alpha2.KVMMachineSpecSize) (resource.Quantity, error) {
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

func NetworkBridgeName(cr v1alpha2.KVMCluster) string {
	return fmt.Sprintf("br-%s", ClusterID(&cr))
}

func NetworkTapName(cr v1alpha2.KVMCluster) string {
	return fmt.Sprintf("tap-%s", ClusterID(&cr))
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

func NodeIndex(cr v1alpha2.KVMCluster, nodeID string) (int, bool) {
	idx, present := cr.Status.Provider.NodeIndexes[nodeID]
	return idx, present
}

// NodeInternalIP examines the Status Adresses of a Node
// and returns its InternalIP..
func NodeInternalIP(node corev1.Node) (string, error) {
	for _, a := range node.Status.Addresses {
		if a.Type == corev1.NodeInternalIP {
			return a.Address, nil
		}
	}
	return "", microerror.Maskf(missingNodeInternalIP, "node %s does not have an InternalIP adress in its status", node.Name)
}

func OperatorVersion(cr v1alpha2.KVMCluster) string {
	return cr.GetLabels()[label.OperatorVersion]
}

// PodIsReady examines the Status Conditions of a Pod
// and returns true if the PodReady Condition is true.
func PodIsReady(pod corev1.Pod) bool {
	podReady := false
	for _, c := range pod.Status.Conditions {
		if c.Type == corev1.PodReady && c.Status == corev1.ConditionTrue {
			podReady = true
		}
	}
	return podReady
}

func PortMappings(cr v1alpha2.KVMCluster) []corev1.ServicePort {
	var ports []corev1.ServicePort

	// Compatibility mode, if no port mappings specified.
	if len(cr.Spec.Provider.PortMappings) == 0 {
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

	for _, p := range cr.Spec.Provider.PortMappings {
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

func PVCNames(cr v1alpha2.KVMCluster) []string {
	return []string{EtcdPVCName(ClusterID(&cr), VMNumber(0))}
}

func ReleaseVersion(cr v1alpha2.KVMCluster) string {
	return cr.GetLabels()[label.ReleaseVersion]
}

func ServiceAccountName(cr v1alpha2.KVMCluster) string {
	return ClusterID(&cr)
}

func ShutdownDeferrerListenPort(cr v1alpha2.KVMCluster) int {
	return int(shutdownDeferrerPortBase + cr.Spec.Provider.FlannelVNI)
}

func ShutdownDeferrerListenAddress(cr v1alpha2.KVMCluster) string {
	return "http://" + ProbeHost + ":" + strconv.Itoa(ShutdownDeferrerListenPort(cr))
}

func ShutdownDeferrerPollPath(cr v1alpha2.KVMCluster) string {
	return fmt.Sprintf("%s/v1/defer/", ShutdownDeferrerListenAddress(cr))
}

func StorageType(cr v1alpha2.KVMCluster) string {
	return cr.Spec.Provider.MachineStorageType
}

func ToClusterEndpoint(v interface{}) (string, error) {
	cr, err := ToKVMCluster(v)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return ClusterAPIEndpoint(cr), nil
}

func ToClusterID(v interface{}) (string, error) {
	cr, err := ToKVMCluster(v)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return ClusterID(&cr), nil
}

func ToClusterStatus(v interface{}) (v1alpha1.StatusCluster, error) {
	cr, err := ToKVMConfig(v)
	if err != nil {
		return v1alpha1.StatusCluster{}, microerror.Mask(err)
	}

	return cr.Status.Cluster, nil
}

func ToKVMConfig(v interface{}) (v1alpha1.KVMConfig, error) {
	crPointer, ok := v.(*v1alpha1.KVMConfig)
	if !ok {
		return v1alpha1.KVMConfig{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &v1alpha1.KVMConfig{}, v)
	}
	cr := *crPointer

	return cr, nil
}

func ToKVMCluster(v interface{}) (v1alpha2.KVMCluster, error) {
	crPointer, ok := v.(*v1alpha2.KVMCluster)
	if !ok {
		return v1alpha2.KVMCluster{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &v1alpha2.KVMCluster{}, v)
	}
	cr := *crPointer

	return cr, nil
}

func ToKVMMachine(v interface{}) (v1alpha2.KVMMachine, error) {
	crPointer, ok := v.(*v1alpha2.KVMMachine)
	if !ok {
		return v1alpha2.KVMMachine{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &v1alpha2.KVMMachine{}, v)
	}
	cr := *crPointer

	return cr, nil
}

func ToNodeCount(v interface{}) (int, error) {
	cr, err := ToKVMCluster(v)
	if err != nil {
		return 0, microerror.Mask(err)
	}

	nodeCount := MasterCount(cr) + WorkerCount(cr)

	return nodeCount, nil
}

func ToOperatorVersion(v interface{}) (string, error) {
	cr, err := ToKVMCluster(v)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return OperatorVersion(cr), nil
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

func VMNumber(ID int) string {
	return fmt.Sprintf("%d", ID)
}

func WorkerCount(cr v1alpha2.KVMCluster) int {
	return 2
}
