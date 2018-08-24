package template

// TODO move this to e2etemplates
// KVMOperatorChartValues values required by kvm-operator-chart, the environment
// variables will be expanded before writing the contents to a file.
var KVMOperatorChartValues = `Installation:
  V1:
    Guest:
      SSH:
        SSOPublicKey: 'test'
      Kubernetes:
        API:
          Auth:
            Provider:
              OIDC:
                ClientID: ""
                IssueURL: ""
                UsernameClaim: ""
                GroupsClaim: ""
      Update:
        Enabled: true
    Secret:
      Registry:
        PullSecret:
          DockerConfigJSON: "{\"auths\":{\"quay.io\":{\"auth\":\"${REGISTRY_PULL_SECRET}\"}}}"
labelSelector: 'clusterID=${CLUSTER_NAME}'
`
