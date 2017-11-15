package deployment

import (
	"context"
	"fmt"
	"strings"

	"github.com/giantswarm/microerror"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"

	"github.com/giantswarm/kvm-operator/service/key"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.Log("cluster", key.ClusterID(customObject), "debug", "looking for a list of deployments in the Kubernetes API")

	var currentDeployments []*v1beta1.Deployment
	{
		namespace := key.ClusterNamespace(customObject)
		deploymentList, err := r.k8sClient.Extensions().Deployments(namespace).List(apismetav1.ListOptions{})
		if err != nil {
			return nil, microerror.Mask(err)
		} else {
			r.logger.Log("cluster", key.ClusterID(customObject), "debug", "found a list of deployments in the Kubernetes API")

			for _, item := range deploymentList.Items {
				d := item

				major, minor, patch, err := getVersionBundleVersionInfos(d.Labels)
				if err != nil {
					r.logger.Log("cluster", key.ClusterID(customObject), "warning", fmt.Sprintf("cannot to update current version bundle version metric for guest cluster: %#v", err))
				} else {
					updateVersionBundleVersionMetric(major, minor, patch)
				}

				currentDeployments = append(currentDeployments, &d)
			}
		}
	}

	r.logger.Log("cluster", key.ClusterID(customObject), "debug", fmt.Sprintf("found a list of %d deployments in the Kubernetes API", len(currentDeployments)))

	return currentDeployments, nil
}

func getVersionBundleVersionInfos(labels map[string]string) (string, string, string, error) {
	version, ok := labels[VersionBundleVersionLabel]
	if !ok {
		return "", "", "", microerror.Maskf(executionFailedError, "label '%s' must not be empty", VersionBundleVersionLabel)
	}

	split := strings.Split(version, ".")
	if len(split) != 3 {
		return "", "", "", microerror.Maskf(executionFailedError, "invalid version format, expected '<major>.<minor>.<patch>', got '%s'", version)
	}

	major := split[0]
	minor := split[1]
	patch := split[2]

	return major, minor, patch, nil
}
