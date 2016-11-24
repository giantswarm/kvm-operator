package controller

import (
	"errors"
	"reflect"
	"testing"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/pkg/api/unversioned"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/pkg/runtime"
	"k8s.io/client-go/pkg/util/intstr"
)

type diff struct {
	existingResource runtime.Object
	expectedResource runtime.Object
}

func checkExistingResourcesMatchExpectedResources(clientset kubernetes.Interface, expectedResources []runtime.Object) ([]diff, error) {
	diffs := []diff{}

	for _, expectedResource := range expectedResources {
		var existingResource runtime.Object
		var err error

		switch res := expectedResource.(type) {
		case *v1.ConfigMap:
			existingResource, err = clientset.Core().ConfigMaps(res.Namespace).Get(res.Name)
		case *v1.Service:
			existingResource, err = clientset.Core().Services(res.Namespace).Get(res.Name)
		case *v1beta1.Deployment:
			existingResource, err = clientset.Extensions().Deployments(res.Namespace).Get(res.Name)
		case *v1beta1.Ingress:
			existingResource, err = clientset.Extensions().Ingresses(res.Namespace).Get(res.Name)
		case *v1beta1.Job:
			existingResource, err = clientset.Extensions().Jobs(res.Namespace).Get(res.Name)
		default:
			return []diff{}, errors.New("expected resource was of unknown type")
		}

		if err != nil {
			return []diff{}, err
		}
		if !reflect.DeepEqual(existingResource, expectedResource) {
			diffs = append(diffs, diff{existingResource: existingResource, expectedResource: expectedResource})
		}
	}

	return diffs, nil
}

func TestReconcileResourceState(t *testing.T) {
	namespaceName := "test-namespace"

	tests := []struct {
		existingResources []runtime.Object
		newResources      []runtime.Object
		expectedResources []runtime.Object
	}{
		// Test creating a single Service, with no pre-existing resources.
		{
			// There are no existing resources,
			existingResources: []runtime.Object{},
			// and we want to add this new Service.
			newResources: []runtime.Object{
				&v1.Service{
					TypeMeta: unversioned.TypeMeta{
						Kind:       "Service",
						APIVersion: "v1",
					},
					ObjectMeta: v1.ObjectMeta{
						Name:      "test-service",
						Namespace: namespaceName,
					},
					Spec: v1.ServiceSpec{
						Ports: []v1.ServicePort{
							v1.ServicePort{
								Port: int32(8000),
							},
						},
					},
				},
			},
			// So, we would expect this Service to exist at the end.
			expectedResources: []runtime.Object{
				&v1.Service{
					TypeMeta: unversioned.TypeMeta{
						Kind:       "Service",
						APIVersion: "v1",
					},
					ObjectMeta: v1.ObjectMeta{
						Name:      "test-service",
						Namespace: namespaceName,
						Annotations: map[string]string{
							"test": "{\"kind\":\"Service\",\"apiVersion\":\"v1\",\"metadata\":{\"name\":\"test-service\",\"namespace\":\"test-namespace\",\"creationTimestamp\":null},\"spec\":{\"ports\":[{\"port\":8000,\"targetPort\":0}]},\"status\":{\"loadBalancer\":{}}}\n",
						},
					},
					Spec: v1.ServiceSpec{
						Ports: []v1.ServicePort{
							v1.ServicePort{
								Port: int32(8000),
							},
						},
					},
				},
			},
		},

		// Test updating a Service
		{
			// We have a Service already,
			existingResources: []runtime.Object{
				&v1.Service{
					TypeMeta: unversioned.TypeMeta{
						Kind:       "Service",
						APIVersion: "v1",
					},
					ObjectMeta: v1.ObjectMeta{
						Name:      "test-service",
						Namespace: namespaceName,
						Annotations: map[string]string{
							"test": "{\"kind\":\"Service\",\"apiVersion\":\"v1\",\"metadata\":{\"name\":\"test-service\",\"namespace\":\"test-namespace\",\"creationTimestamp\":null},\"spec\":{\"ports\":[{\"port\":8080,\"targetPort\":0}],\"type\":\"ClusterIP\"},\"status\":{\"loadBalancer\":{}}}\n",
						},
					},
					Spec: v1.ServiceSpec{
						Type: v1.ServiceTypeClusterIP,
						Ports: []v1.ServicePort{
							v1.ServicePort{
								Port: int32(8000),
							},
						},
						ClusterIP: "192.0.2.1", // ClusterIP created by Kubernetes
					},
				},
			},
			// and want to update it with this version.
			newResources: []runtime.Object{
				&v1.Service{
					TypeMeta: unversioned.TypeMeta{
						Kind:       "Service",
						APIVersion: "v1",
					},
					ObjectMeta: v1.ObjectMeta{
						Name:      "test-service",
						Namespace: namespaceName,
					},
					Spec: v1.ServiceSpec{
						Type: v1.ServiceTypeClusterIP,
						Ports: []v1.ServicePort{
							v1.ServicePort{
								Port: int32(8080), // This port has changed
							},
						},
					},
				},
			},
			// So we'd expect this Service spec to exist at the end.
			expectedResources: []runtime.Object{
				&v1.Service{
					TypeMeta: unversioned.TypeMeta{
						Kind:       "Service",
						APIVersion: "v1",
					},
					ObjectMeta: v1.ObjectMeta{
						Name:      "test-service",
						Namespace: namespaceName,
					},
					Spec: v1.ServiceSpec{
						Type: v1.ServiceTypeClusterIP,
						Ports: []v1.ServicePort{
							v1.ServicePort{
								Port: int32(8080),
							},
						},
						ClusterIP: "192.0.2.1", // ClusterIP preserved
					},
				},
			},
		},

		// Test updating an Ingress, without touching the existing ConfigMap.
		{
			// We have an Ingress and a configmap already
			existingResources: []runtime.Object{
				&v1beta1.Ingress{
					TypeMeta: unversioned.TypeMeta{
						Kind:       "Ingress",
						APIVersion: "extensions/v1beta1",
					},
					ObjectMeta: v1.ObjectMeta{
						Name:      "test-ingress",
						Namespace: namespaceName,
					},
					Spec: v1beta1.IngressSpec{
						Backend: &v1beta1.IngressBackend{
							ServiceName: "test-service",
							ServicePort: intstr.FromInt(8000),
						},
					},
				},
				&v1.ConfigMap{
					TypeMeta: unversioned.TypeMeta{
						Kind:       "ConfigMap",
						APIVersion: "v1",
					},
					ObjectMeta: v1.ObjectMeta{
						Name:      "test-configmap",
						Namespace: namespaceName,
					},
					Data: map[string]string{},
				},
			},

			// And we want to change the port for the ingress backend,
			// without affecting the configmap.
			newResources: []runtime.Object{
				&v1beta1.Ingress{
					TypeMeta: unversioned.TypeMeta{
						Kind:       "Ingress",
						APIVersion: "extensions/v1beta1",
					},
					ObjectMeta: v1.ObjectMeta{
						Name:      "test-ingress",
						Namespace: namespaceName,
					},
					Spec: v1beta1.IngressSpec{
						Backend: &v1beta1.IngressBackend{
							ServiceName: "test-service",
							ServicePort: intstr.FromInt(8001),
						},
					},
				},
				&v1.ConfigMap{
					TypeMeta: unversioned.TypeMeta{
						Kind:       "ConfigMap",
						APIVersion: "v1",
					},
					ObjectMeta: v1.ObjectMeta{
						Name:      "test-configmap",
						Namespace: namespaceName,
					},
					Data: map[string]string{},
				},
			},

			// so we expect this to be the actual state
			expectedResources: []runtime.Object{
				&v1beta1.Ingress{
					TypeMeta: unversioned.TypeMeta{
						Kind:       "Ingress",
						APIVersion: "extensions/v1beta1",
					},
					ObjectMeta: v1.ObjectMeta{
						Name:      "test-ingress",
						Namespace: namespaceName,
					},
					Spec: v1beta1.IngressSpec{
						Backend: &v1beta1.IngressBackend{
							ServiceName: "test-service",
							ServicePort: intstr.FromInt(8001),
						},
					},
				},
				&v1.ConfigMap{
					TypeMeta: unversioned.TypeMeta{
						Kind:       "ConfigMap",
						APIVersion: "v1",
					},
					ObjectMeta: v1.ObjectMeta{
						Name:      "test-configmap",
						Namespace: namespaceName,
					},
					Data: map[string]string{},
				},
			},
		},
	}

	for index, test := range tests {
		clientset := fake.NewSimpleClientset(test.existingResources...)

		controller := &controller{clientset: clientset}

		if err := controller.reconcileResourceState(namespaceName, test.newResources); err != nil {
			t.Fatalf("%v: an error occurred reconciling resource state: %v", index, err)
		}

		diffs, err := checkExistingResourcesMatchExpectedResources(clientset, test.expectedResources)
		if err != nil {
			t.Fatalf("%v: an error occurred checking existing resources matching expected resources: %v", index, err)
		}

		if len(diffs) > 0 {
			t.Logf("%v: Existing resources did not match expected resources", index)
			for _, diff := range diffs {
				t.Logf("Existing: \n %#v \n", diff.existingResource)
				t.Logf("Expected: \n %#v \n", diff.expectedResource)
			}
			t.FailNow()
		}
	}
}
