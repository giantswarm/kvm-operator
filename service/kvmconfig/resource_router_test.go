package kvmconfig

import (
	"context"
	"reflect"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/operatorkit/framework"

	"github.com/giantswarm/kvm-operator/service/kvmconfig/resource/configmapv3"
	"github.com/giantswarm/kvm-operator/service/kvmconfig/resource/servicev2"
)

func Test_Service_newResourceRouter(t *testing.T) {
	resourceV1 := []framework.Resource{
		&servicev2.Resource{},
	}
	resourceV2 := []framework.Resource{
		&servicev2.Resource{},
		&configmapv3.Resource{},
	}

	versionedResources := map[string][]framework.Resource{
		"1.0.0": resourceV1,
		"2.0.0": resourceV2,
		"":      resourceV1,
	}

	testCases := []struct {
		customObject      v1alpha1.KVMConfig
		expectedResources []framework.Resource
		errorMatcher      func(err error) bool
	}{
		// Test 0, get resource
		{
			customObject: v1alpha1.KVMConfig{
				Spec: v1alpha1.KVMConfigSpec{
					VersionBundle: v1alpha1.KVMConfigSpecVersionBundle{
						Version: "1.0.0",
					},
				},
			},
			expectedResources: resourceV1,
			errorMatcher:      nil,
		},
		// Test 1, get correct resources on multiples resources version
		{
			customObject: v1alpha1.KVMConfig{
				Spec: v1alpha1.KVMConfigSpec{
					VersionBundle: v1alpha1.KVMConfigSpecVersionBundle{
						Version: "2.0.0",
					},
				},
			},
			expectedResources: resourceV2,
			errorMatcher:      nil,
		},
		// Test 2, get resources from empty Version Bundle
		{
			customObject: v1alpha1.KVMConfig{
				Spec: v1alpha1.KVMConfigSpec{
					VersionBundle: v1alpha1.KVMConfigSpecVersionBundle{
						Version: "",
					},
				},
			},
			expectedResources: resourceV1,
			errorMatcher:      nil,
		},
		// Test 3, Invalid version returns an error.
		{
			customObject: v1alpha1.KVMConfig{
				Spec: v1alpha1.KVMConfigSpec{
					VersionBundle: v1alpha1.KVMConfigSpecVersionBundle{
						Version: "4.0.0",
					},
				},
			},
			expectedResources: nil,
			errorMatcher:      IsInvalidVersion,
		},
	}
	for i, tc := range testCases {
		result := newResourceRouter(versionedResources)

		resources, err := result(context.TODO(), &tc.customObject)
		if err != nil {
			if tc.errorMatcher == nil {
				t.Fatal("test", i, "expected", nil, "got", "error matcher")
			} else if !tc.errorMatcher(err) {
				t.Fatal("test", i, "expected", true, "got", false)
			}
		} else {
			if !reflect.DeepEqual(tc.expectedResources, resources) {
				t.Fatal("test", i, "expected", tc.expectedResources, "got", resources)
			}
		}
	}
}
