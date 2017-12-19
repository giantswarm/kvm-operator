package service

import (
	"context"
	"reflect"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/kvm-operator/service/resource/configmapv3"
	"github.com/giantswarm/kvm-operator/service/resource/servicev2"
	"github.com/giantswarm/operatorkit/framework"
)

func Test_NewResourceRouter(t *testing.T) {
	resourceV1 := []framework.Resource{
		&servicev2.Resource{},
	}
	resourceV2 := []framework.Resource{
		&servicev2.Resource{},
		&configmapv3.Resource{},
	}

	testCases := []struct {
		ctx                context.Context
		versionedResources map[string][]framework.Resource
		customObject       interface{}
		expectedResources  []framework.Resource
	}{
		// Test 1, get resource
		{
			ctx: context.TODO(),
			versionedResources: map[string][]framework.Resource{
				"1.0.0": resourceV1,
			},
			customObject: &v1alpha1.KVMConfig{
				Spec: v1alpha1.KVMConfigSpec{
					VersionBundle: v1alpha1.KVMConfigSpecVersionBundle{
						Version: "1.0.0",
					},
				},
			},
			expectedResources: resourceV1,
		},
		// Test 2, get correct resources on multiples resources version
		{
			ctx: context.TODO(),
			versionedResources: map[string][]framework.Resource{
				"1.0.0": resourceV1,
				"2.0.0": resourceV2,
			},
			customObject: &v1alpha1.KVMConfig{
				Spec: v1alpha1.KVMConfigSpec{
					VersionBundle: v1alpha1.KVMConfigSpecVersionBundle{
						Version: "2.0.0",
					},
				},
			},
			expectedResources: resourceV2,
		},
		// Test 3, get resources from empty Version Bundle
		{
			ctx: context.TODO(),
			versionedResources: map[string][]framework.Resource{
				"1.0.0": resourceV1,
				"2.0.0": resourceV2,
				"":      resourceV1,
			},
			customObject: &v1alpha1.KVMConfig{
				Spec: v1alpha1.KVMConfigSpec{
					VersionBundle: v1alpha1.KVMConfigSpecVersionBundle{
						Version: "",
					},
				},
			},
			expectedResources: resourceV1,
		},
	}
	for i, tc := range testCases {
		result := newResourceRouter(tc.versionedResources)

		resources, err := result(tc.ctx, tc.customObject)
		if err != nil {
			t.Fatalf("case %d expected %#v got %#v", i+1, nil, err)
		}
		if !reflect.DeepEqual(resources, tc.expectedResources) {
			t.Fatalf("case %d expected %#v got %#v len(%v)", i+1, tc.expectedResources, resources, len(resources))

		}
	}
}
