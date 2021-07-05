package deployment

import (
	"context"
	"fmt"
	"strings"

	"github.com/giantswarm/apiextensions/v3/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/prometheus/client_golang/prometheus"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/kvm-operator/v4/service/controller/key"
	"github.com/giantswarm/kvm-operator/v4/service/metric"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	customResource, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.Debugf(ctx, "looking for a list of deployments in the Kubernetes API")

	var currentDeployments []*v1.Deployment
	{
		var deploymentList v1.DeploymentList
		err := r.ctrlClient.List(ctx, &deploymentList, &client.ListOptions{
			Namespace: key.ClusterNamespace(customResource),
			LabelSelector: labels.SelectorFromSet(map[string]string{
				key.LabelManagedBy: key.OperatorName,
			}),
		})
		if err != nil {
			return nil, microerror.Mask(err)
		} else {
			r.logger.Debugf(ctx, "found a list of deployments in the Kubernetes API")

			for _, item := range deploymentList.Items {
				d := item
				currentDeployments = append(currentDeployments, &d)
			}

			r.updateVersionBundleVersionGauge(ctx, customResource, metric.VersionBundleVersionGauge, currentDeployments)
		}
	}

	r.logger.Debugf(ctx, "found a list of %d deployments in the Kubernetes API", len(currentDeployments))

	return currentDeployments, nil
}

func (r *Resource) updateVersionBundleVersionGauge(ctx context.Context, customResource v1alpha1.KVMConfig, gauge *prometheus.GaugeVec, deployments []*v1.Deployment) {
	versionCounts := map[string]float64{}

	for _, d := range deployments {
		version, ok := d.Annotations[key.VersionBundleVersionAnnotation]
		if !ok {
			r.logger.LogCtx(ctx, "level", "warning", "message", fmt.Sprintf("cannot update current deployment: annotation '%s' must not be empty", key.VersionBundleVersionAnnotation))
			continue
		} else {
			count, ok := versionCounts[version]
			if !ok {
				versionCounts[version] = 1
			} else {
				versionCounts[version] = count + 1
			}
		}
	}

	for version, count := range versionCounts {
		split := strings.Split(version, ".")
		if len(split) != 3 {
			r.logger.LogCtx(ctx, "level", "warning", "message", fmt.Sprintf("cannot update current deployment: invalid version format, expected '<major>.<minor>.<patch>', got '%s'", version))
			continue
		}

		major := split[0]
		minor := split[1]
		patch := split[2]

		gauge.WithLabelValues(major, minor, patch).Set(count)
	}
}
