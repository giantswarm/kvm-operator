package template

type KVMConfigE2eChartValues struct {
	ClusterID            string
	HttpNodePort         int
	HttpsNodePort        int
	VNI                  int
	VersionBundleVersion string
}

const ApiextensionsKVMConfigE2EChartValues = `
cluster:
  id:  "{{.ClusterID}}"
baseDomain: "k8s.gastropod.gridscale.kvm.gigantic.io"
sshUser: "test-user"
sshPublicKey: "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAAAYQCurvzg5Ia54kb3NZapA6yP00//+Jt6XJNeC7Seq3TeCqMR9x7Snalj19r0lWok1PkRgDo1PXj+3y53zo/wqBrPqN4cQqp00R06kNfnhAgesaRMvYhuyVRQQbfXV5gQg8M= dummy-key"
encryptionKey: "QitRZGlWeW5WOFo2YmdvMVRwQUQ2UWoxRHZSVEF4MmovajlFb05sT1AzOD0="
versionBundleVersion: "{{.VersionBundleVersion}}"
updateEnabled: true
kvm:
  vni: {{.VNI}}
  ingress:
    httpNodePort: {{.HttpNodePort}}
    httpTargetPort: 30010
    httpsNodePort: {{.HttpsNodePort}}
    httpsTargetPort: 30011
`
