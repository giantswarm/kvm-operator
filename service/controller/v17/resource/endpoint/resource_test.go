package endpoint

import (
	"reflect"
	"testing"
)

func Test_Resource_Endpoint_cutIPs(t *testing.T) {
	testCases := []struct {
		name         string
		base         []string
		cutset       []string
		expectedSet  []string
		errorMatcher func(error) bool
	}{
		{
			name: "case 0: Base has one more address than the cutset",
			base: []string{
				"1.1.1.1",
				"0.0.0.0",
			},
			cutset: []string{
				"1.1.1.1",
			},
			expectedSet: []string{
				"0.0.0.0",
			},
			errorMatcher: nil,
		},
		{
			name: "case 1: Base and cutset have the same addresses",
			base: []string{
				"1.1.1.1",
			},
			cutset: []string{
				"1.1.1.1",
			},
			expectedSet:  []string{},
			errorMatcher: nil,
		},
		{
			name: "case 2: Cutset has one more address than base",
			base: []string{
				"1.1.1.1",
			},
			cutset: []string{
				"1.1.1.1",
				"0.0.0.0",
			},
			expectedSet:  []string{},
			errorMatcher: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			resultSet := cutIPs(tc.base, tc.cutset)

			if !reflect.DeepEqual(resultSet, tc.expectedSet) {
				t.Fatalf("resultSet == %q, want %q", resultSet, tc.expectedSet)
			}
		})
	}
}
