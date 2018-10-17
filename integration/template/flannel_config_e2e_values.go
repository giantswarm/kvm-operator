package template

type FlannelConfigE2eChartValues struct {
	ClusterID string
	Network   string
	VNI       int
}

const ApiextensionsFlannelConfigE2EChartValues = `
clusterName: "{{.ClusterID}}"
versionBundleVersion: "0.2.0"
flannel:
  network: "{{.Network}}"
  vni: {{.VNI}}
`
