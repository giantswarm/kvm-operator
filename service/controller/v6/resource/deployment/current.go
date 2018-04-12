package deployment

import (
	"context"
	"fmt"
	"strings"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/api/extensions/v1beta1"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/kvm-operator/service/controller/v6/key"
	"github.com/giantswarm/kvm-operator/service/metric"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "debug", "looking for a list of deployments in the Kubernetes API")

	var currentDeployments []*v1beta1.Deployment
	{
		namespace := key.ClusterNamespace(customObject)
		deploymentList, err := r.k8sClient.Extensions().Deployments(namespace).List(apismetav1.ListOptions{})
		if err != nil {
			return nil, microerror.Mask(err)
		} else {
			r.logger.LogCtx(ctx, "debug", "found a list of deployments in the Kubernetes API")

			for _, item := range deploymentList.Items {
				d := item
				currentDeployments = append(currentDeployments, &d)
			}

			r.updateVersionBundleVersionGauge(ctx, customObject, metric.VersionBundleVersionGauge, currentDeployments)
		}
	}

	r.logger.LogCtx(ctx, "debug", fmt.Sprintf("found a list of %d deployments in the Kubernetes API", len(currentDeployments)))

	return currentDeployments, nil
}

func (r *Resource) updateVersionBundleVersionGauge(ctx context.Context, customObject v1alpha1.KVMConfig, gauge *prometheus.GaugeVec, deployments []*v1beta1.Deployment) {
	versionCounts := map[string]float64{}

	for _, d := range deployments {
		version, ok := d.Annotations[key.VersionBundleVersionAnnotation]
		if !ok {
			r.logger.LogCtx(ctx, "warning", fmt.Sprintf("cannot update current deployment: annotation '%s' must not be empty", key.VersionBundleVersionAnnotation))
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
			r.logger.LogCtx(ctx, "warning", fmt.Sprintf("cannot update current deployment: invalid version format, expected '<major>.<minor>.<patch>', got '%s'", version))
			continue
		}

		major := split[0]
		minor := split[1]
		patch := split[2]

		gauge.WithLabelValues(major, minor, patch).Set(count)
	}
}
