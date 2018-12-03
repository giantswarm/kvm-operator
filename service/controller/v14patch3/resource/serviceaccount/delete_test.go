package serviceaccount

import (
	"context"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/micrologger/microloggertest"
	apiv1 "k8s.io/api/core/v1"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func Test_Resource_ServiceAccount_newDeleteChange(t *testing.T) {
	testCases := []struct {
		Obj                    interface{}
		Cur                    interface{}
		Des                    interface{}
		ExpectedServiceAccount *apiv1.ServiceAccount
	}{
		{
			Obj: &v1alpha1.KVMConfig{
				Spec: v1alpha1.KVMConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "my-cluster",
					},
				},
			},
			Cur: &apiv1.ServiceAccount{
				TypeMeta: apismetav1.TypeMeta{
					Kind:       "ServiceAccount",
					APIVersion: "v1",
				},
				ObjectMeta: apismetav1.ObjectMeta{
					Name: "al9qy",
					Labels: map[string]string{
						"cluster":  "al9qy",
						"customer": "test-customer",
					},
				},
			},
			Des: &apiv1.ServiceAccount{
				TypeMeta: apismetav1.TypeMeta{
					Kind:       "ServiceAccount",
					APIVersion: "v1",
				},
				ObjectMeta: apismetav1.ObjectMeta{
					Name: "al9qy",
					Labels: map[string]string{
						"cluster":  "al9qy",
						"customer": "test-customer",
					},
				},
			},
			ExpectedServiceAccount: &apiv1.ServiceAccount{
				TypeMeta: apismetav1.TypeMeta{
					Kind:       "ServiceAccount",
					APIVersion: "v1",
				},
				ObjectMeta: apismetav1.ObjectMeta{
					Name: "al9qy",
					Labels: map[string]string{
						"cluster":  "al9qy",
						"customer": "test-customer",
					},
				},
			},
		},

		{
			Obj: &v1alpha1.KVMConfig{
				Spec: v1alpha1.KVMConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "my-cluster",
					},
				},
			},
			Cur: nil,
			Des: &apiv1.ServiceAccount{
				TypeMeta: apismetav1.TypeMeta{
					Kind:       "ServiceAccount",
					APIVersion: "v1",
				},
				ObjectMeta: apismetav1.ObjectMeta{
					Name: "al9qy",
					Labels: map[string]string{
						"cluster":  "al9qy",
						"customer": "test-customer",
					},
				},
			},
			ExpectedServiceAccount: nil,
		},
	}

	var err error
	var newResource *Resource
	{
		resourceConfig := DefaultConfig()
		resourceConfig.K8sClient = fake.NewSimpleClientset()
		resourceConfig.Logger = microloggertest.New()
		newResource, err = New(resourceConfig)
		if err != nil {
			t.Fatal("expected", nil, "got", err)
		}
	}

	for i, tc := range testCases {
		result, err := newResource.newDeleteChange(context.TODO(), tc.Obj, tc.Cur, tc.Des)
		if err != nil {
			t.Fatal("case", i+1, "expected", nil, "got", err)
		}
		if tc.ExpectedServiceAccount == nil {
			if tc.ExpectedServiceAccount != result {
				t.Fatal("case", i+1, "expected", tc.ExpectedServiceAccount, "got", result)
			}
		} else {
			name := result.(*apiv1.ServiceAccount).Name
			if tc.ExpectedServiceAccount.Name != name {
				t.Fatal("case", i+1, "expected", tc.ExpectedServiceAccount.Name, "got", name)
			}
		}
	}
}
