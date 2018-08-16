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

const FlannelOperatorChartValues = `
clusterName: ${CLUSTER_NAME}
Installation:
  V1:    
    Secret:      
      Registry:
        PullSecret:
          DockerConfigJSON: "{\"auths\":{\"quay.io\":{\"auth\":\"$REGISTRY_PULL_SECRET\"}}}"
`
