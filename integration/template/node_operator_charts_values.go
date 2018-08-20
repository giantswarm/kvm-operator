package template

const NodeOperatorChartValues = `
clusterName: ${CLUSTER_NAME}
Installation:
  V1:    
    Secret:      
      Registry:
        PullSecret:
          DockerConfigJSON: "{\"auths\":{\"quay.io\":{\"auth\":\"$REGISTRY_PULL_SECRET\"}}}"
`
