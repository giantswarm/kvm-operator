package controller

import (
	"errors"
	"reflect"
	"testing"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/pkg/api/unversioned"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/runtime"
)

type diff struct {
	existingResource runtime.Object
	expectedResource runtime.Object
}

func checkExistingResourcesMatchExpectedResources(clientset kubernetes.Interface, expectedResources []runtime.Object) ([]diff, error) {
	diffs := []diff{}

	for _, expectedResource := range expectedResources {
		switch r := expectedResource.(type) {

		case *v1.Service:
			service, err := clientset.Core().Services(r.Namespace).Get(r.Name)
			if err != nil {
				return []diff{}, err
			}
			if !reflect.DeepEqual(service, expectedResource) {
				diffs = append(diffs, diff{existingResource: r, expectedResource: expectedResource})
			}

		default:
			return []diff{}, errors.New("expected resource was of unknown type")
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
			existingResources: []runtime.Object{},
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
						Ports: []v1.ServicePort{
							v1.ServicePort{
								Port: int32(8000),
							},
						},
					},
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
			t.Logf("Existing resources did not match expected resources")
			for _, diff := range diffs {
				t.Logf("Existing: \n %#v \n", diff.existingResource)
				t.Logf("Expected: \n %#v \n", diff.expectedResource)
			}
			t.FailNow()
		}
	}
}
