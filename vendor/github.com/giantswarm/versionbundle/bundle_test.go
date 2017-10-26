package versionbundle

import (
	"reflect"
	"sort"
	"testing"
	"time"
)

func Test_Bundle_Validate(t *testing.T) {
	testCases := []struct {
		Bundle       Bundle
		ErrorMatcher func(err error) bool
	}{
		// Test 0 ensures that an empty version bundle is not valid.
		{
			Bundle:       Bundle{},
			ErrorMatcher: IsInvalidBundleError,
		},

		// Test 1 is the same as 0 but with an empty list of bundles.
		{
			Bundle: Bundle{
				Changelogs:   []Changelog{},
				Components:   []Component{},
				Dependencies: []Dependency{},
				Deprecated:   false,
				Name:         "",
				Time:         time.Time{},
				Version:      "",
				WIP:          false,
			},
			ErrorMatcher: IsInvalidBundleError,
		},

		// Test 2 ensures a version bundle without changelogs throws an error.
		{
			Bundle: Bundle{
				Changelogs: []Changelog{},
				Components: []Component{
					{
						Name:    "calico",
						Version: "1.1.0",
					},
					{
						Name:    "kube-dns",
						Version: "1.0.0",
					},
				},
				Dependencies: []Dependency{
					{
						Name:    "kubernetes",
						Version: "<= 1.7.x",
					},
				},
				Deprecated: false,
				Name:       "kubernetes-operator",
				Time:       time.Unix(10, 5),
				Version:    "0.1.0",
				WIP:        false,
			},
			ErrorMatcher: IsInvalidBundleError,
		},

		// Test 3 ensures a version bundle without components throws an error.
		{
			Bundle: Bundle{
				Changelogs: []Changelog{
					{
						Component:   "calico",
						Description: "Calico version updated.",
						Kind:        "changed",
					},
					{
						Component:   "kubernetes",
						Description: "Kubernetes version requirements changed due to calico update.",
						Kind:        "changed",
					},
				},
				Components: []Component{},
				Dependencies: []Dependency{
					{
						Name:    "kubernetes",
						Version: "<= 1.7.x",
					},
				},
				Deprecated: false,
				Name:       "kubernetes-operator",
				Time:       time.Unix(10, 5),
				Version:    "0.1.0",
				WIP:        false,
			},
			ErrorMatcher: IsInvalidBundleError,
		},

		// Test 4 ensures a version bundle without dependencies does not throw an
		// error.
		{
			Bundle: Bundle{
				Changelogs: []Changelog{
					{
						Component:   "calico",
						Description: "Calico version updated.",
						Kind:        "changed",
					},
				},
				Components: []Component{
					{
						Name:    "calico",
						Version: "1.1.0",
					},
					{
						Name:    "kube-dns",
						Version: "1.0.0",
					},
				},
				Dependencies: []Dependency{},
				Deprecated:   false,
				Name:         "kubernetes-operator",
				Time:         time.Unix(10, 5),
				Version:      "0.1.0",
				WIP:          false,
			},
			ErrorMatcher: nil,
		},

		// Test 5 ensures a version bundle without time throws an error.
		{
			Bundle: Bundle{
				Changelogs: []Changelog{
					{
						Component:   "calico",
						Description: "Calico version updated.",
						Kind:        "changed",
					},
					{
						Component:   "kubernetes",
						Description: "Kubernetes version requirements changed due to calico update.",
						Kind:        "changed",
					},
				},
				Components: []Component{
					{
						Name:    "calico",
						Version: "1.1.0",
					},
					{
						Name:    "kube-dns",
						Version: "1.0.0",
					},
				},
				Dependencies: []Dependency{
					{
						Name:    "kubernetes",
						Version: "<= 1.7.x",
					},
				},
				Deprecated: false,
				Name:       "kubernetes-operator",
				Time:       time.Time{},
				Version:    "0.1.0",
				WIP:        false,
			},
			ErrorMatcher: IsInvalidBundleError,
		},

		// Test 6 ensures a version bundle without version throws an error.
		{
			Bundle: Bundle{
				Changelogs: []Changelog{
					{
						Component:   "calico",
						Description: "Calico version updated.",
						Kind:        "changed",
					},
					{
						Component:   "kubernetes",
						Description: "Kubernetes version requirements changed due to calico update.",
						Kind:        "changed",
					},
				},
				Components: []Component{
					{
						Name:    "calico",
						Version: "1.1.0",
					},
					{
						Name:    "kube-dns",
						Version: "1.0.0",
					},
				},
				Dependencies: []Dependency{
					{
						Name:    "kubernetes",
						Version: "<= 1.7.x",
					},
				},
				Deprecated: false,
				Name:       "kubernetes-operator",
				Time:       time.Unix(10, 5),
				Version:    "",
				WIP:        false,
			},
			ErrorMatcher: IsInvalidBundleError,
		},

		// Test 7 ensures a deprecated version bundle does not throw an error.
		{
			Bundle: Bundle{
				Changelogs: []Changelog{
					{
						Component:   "calico",
						Description: "Calico version updated.",
						Kind:        "changed",
					},
					{
						Component:   "kubernetes",
						Description: "Kubernetes version requirements changed due to calico update.",
						Kind:        "changed",
					},
				},
				Components: []Component{
					{
						Name:    "calico",
						Version: "1.1.0",
					},
					{
						Name:    "kube-dns",
						Version: "1.0.0",
					},
				},
				Dependencies: []Dependency{
					{
						Name:    "kubernetes",
						Version: "<= 1.7.x",
					},
				},
				Deprecated: true,
				Name:       "kubernetes-operator",
				Time:       time.Unix(10, 5),
				Version:    "0.1.0",
				WIP:        false,
			},
			ErrorMatcher: nil,
		},

		// Test 8 ensures a version bundle with an invalid dependency version format
		// throws an error.
		{
			Bundle: Bundle{
				Changelogs: []Changelog{
					{
						Component:   "calico",
						Description: "Calico version updated.",
						Kind:        "changed",
					},
					{
						Component:   "kubernetes",
						Description: "Kubernetes version requirements changed due to calico update.",
						Kind:        "changed",
					},
				},
				Components: []Component{
					{
						Name:    "calico",
						Version: "1.1.0",
					},
					{
						Name:    "kube-dns",
						Version: "1.0.0",
					},
				},
				Dependencies: []Dependency{
					{
						Name:    "kubernetes",
						Version: "1.7.x",
					},
				},
				Deprecated: true,
				Name:       "kubernetes-operator",
				Time:       time.Unix(10, 5),
				Version:    "0.1.0",
				WIP:        false,
			},
			ErrorMatcher: IsInvalidBundleError,
		},

		// Test 9 ensures an invalid version throws an error.
		{
			Bundle: Bundle{
				Changelogs: []Changelog{
					{
						Component:   "calico",
						Description: "Calico version updated.",
						Kind:        "changed",
					},
					{
						Component:   "kubernetes",
						Description: "Kubernetes version requirements changed due to calico update.",
						Kind:        "changed",
					},
				},
				Components: []Component{
					{
						Name:    "calico",
						Version: "1.1.0",
					},
					{
						Name:    "kube-dns",
						Version: "1.0.0",
					},
				},
				Dependencies: []Dependency{
					{
						Name:    "kubernetes",
						Version: "<= 1.7.x",
					},
				},
				Deprecated: true,
				Name:       "kubernetes-operator",
				Time:       time.Unix(10, 5),
				Version:    "foo",
				WIP:        false,
			},
			ErrorMatcher: IsInvalidBundleError,
		},

		// Test 10 is the same as 9 but with a different version.
		{
			Bundle: Bundle{
				Changelogs: []Changelog{
					{
						Component:   "calico",
						Description: "Calico version updated.",
						Kind:        "changed",
					},
					{
						Component:   "kubernetes",
						Description: "Kubernetes version requirements changed due to calico update.",
						Kind:        "changed",
					},
				},
				Components: []Component{
					{
						Name:    "calico",
						Version: "1.1.0",
					},
					{
						Name:    "kube-dns",
						Version: "1.0.0",
					},
				},
				Dependencies: []Dependency{
					{
						Name:    "kubernetes",
						Version: "<= 1.7.x",
					},
				},
				Deprecated: true,
				Name:       "kubernetes-operator",
				Time:       time.Unix(10, 5),
				Version:    "1.2.3.4",
				WIP:        false,
			},
			ErrorMatcher: IsInvalidBundleError,
		},

		// Test 11 ensures a version bundle being flagged as WIP does not throw an
		// error.
		{
			Bundle: Bundle{
				Changelogs: []Changelog{
					{
						Component:   "calico",
						Description: "Calico version updated.",
						Kind:        "changed",
					},
					{
						Component:   "kubernetes",
						Description: "Kubernetes version requirements changed due to calico update.",
						Kind:        "changed",
					},
				},
				Components: []Component{
					{
						Name:    "calico",
						Version: "1.1.0",
					},
					{
						Name:    "kube-dns",
						Version: "1.0.0",
					},
				},
				Dependencies: []Dependency{
					{
						Name:    "kubernetes",
						Version: "<= 1.7.x",
					},
				},
				Deprecated: false,
				Name:       "kubernetes-operator",
				Time:       time.Unix(10, 5),
				Version:    "0.1.0",
				WIP:        true,
			},
			ErrorMatcher: nil,
		},

		// Test 12 ensures a valid version bundle does not throw an error.
		{
			Bundle: Bundle{
				Changelogs: []Changelog{
					{
						Component:   "calico",
						Description: "Calico version updated.",
						Kind:        "changed",
					},
					{
						Component:   "kubernetes",
						Description: "Kubernetes version requirements changed due to calico update.",
						Kind:        "changed",
					},
				},
				Components: []Component{
					{
						Name:    "calico",
						Version: "1.1.0",
					},
					{
						Name:    "kube-dns",
						Version: "1.0.0",
					},
				},
				Dependencies: []Dependency{
					{
						Name:    "kubernetes",
						Version: "<= 1.7.x",
					},
				},
				Deprecated: false,
				Name:       "kubernetes-operator",
				Time:       time.Unix(10, 5),
				Version:    "0.1.0",
				WIP:        false,
			},
			ErrorMatcher: nil,
		},
	}

	for i, tc := range testCases {
		err := tc.Bundle.Validate()
		if tc.ErrorMatcher != nil {
			if !tc.ErrorMatcher(err) {
				t.Fatalf("test %d expected %#v got %#v", i, true, false)
			}
		} else if err != nil {
			t.Fatalf("test %d expected %#v got %#v", i, nil, err)
		}
	}
}

func Test_Bundles_Copy(t *testing.T) {
	bundles := []Bundle{
		{
			Changelogs: []Changelog{},
			Components: []Component{
				{
					Name:    "calico",
					Version: "1.1.0",
				},
				{
					Name:    "kube-dns",
					Version: "1.0.0",
				},
			},
			Dependencies: []Dependency{
				{
					Name:    "kubernetes",
					Version: "<= 1.7.x",
				},
			},
			Deprecated: false,
			Name:       "kubernetes-operator",
			Time:       time.Unix(10, 5),
			Version:    "0.1.0",
			WIP:        false,
		},
		{
			Changelogs: []Changelog{
				{
					Component:   "calico",
					Description: "Calico version updated.",
					Kind:        "changed",
				},
				{
					Component:   "kubernetes",
					Description: "Kubernetes version requirements changed due to calico update.",
					Kind:        "changed",
				},
			},
			Components: []Component{
				{
					Name:    "calico",
					Version: "1.1.0",
				},
				{
					Name:    "kube-dns",
					Version: "1.0.0",
				},
			},
			Dependencies: []Dependency{
				{
					Name:    "kubernetes",
					Version: "<= 1.7.x",
				},
			},
			Deprecated: false,
			Name:       "kubernetes-operator",
			Time:       time.Unix(20, 10),
			Version:    "0.0.9",
			WIP:        false,
		},
	}

	b1 := ValidateBundles(bundles).Copy()
	b2 := ValidateBundles(bundles).Copy()

	sort.Sort(SortBundlesByTime(b1))
	sort.Sort(SortBundlesByVersion(b2))

	if reflect.DeepEqual(b1, b2) {
		t.Fatalf("expected %#v got %#v", b1, b2)
	}
}

func Test_Bundles_Validate(t *testing.T) {
	testCases := []struct {
		Bundles      []Bundle
		ErrorMatcher func(err error) bool
	}{
		// Test 0 ensures that a nil list is not valid.
		{
			Bundles:      nil,
			ErrorMatcher: IsInvalidBundlesError,
		},

		// Test 1 is the same as 0 but with an empty list of bundles.
		{
			Bundles:      []Bundle{},
			ErrorMatcher: IsInvalidBundlesError,
		},

		// Test 2 ensures validation of a list of version bundles where any version
		// bundle has no changelogs throws an error.
		{
			Bundles: []Bundle{
				{
					Changelogs: []Changelog{},
					Components: []Component{
						{
							Name:    "calico",
							Version: "1.1.0",
						},
						{
							Name:    "kube-dns",
							Version: "1.0.0",
						},
					},
					Dependencies: []Dependency{
						{
							Name:    "kubernetes",
							Version: "<= 1.7.x",
						},
					},
					Deprecated: false,
					Name:       "kubernetes-operator",
					Time:       time.Unix(10, 5),
					Version:    "0.1.0",
					WIP:        false,
				},
			},
			ErrorMatcher: IsInvalidBundlesError,
		},

		// Test 3 is the same as 2 but with multiple version bundles.
		{
			Bundles: []Bundle{
				{
					Changelogs: []Changelog{},
					Components: []Component{
						{
							Name:    "calico",
							Version: "1.1.0",
						},
						{
							Name:    "kube-dns",
							Version: "1.0.0",
						},
					},
					Dependencies: []Dependency{
						{
							Name:    "kubernetes",
							Version: "<= 1.7.x",
						},
					},
					Deprecated: false,
					Name:       "kubernetes-operator",
					Time:       time.Unix(10, 5),
					Version:    "0.1.0",
					WIP:        false,
				},
				{
					Changelogs: []Changelog{
						{
							Component:   "calico",
							Description: "Calico version updated.",
							Kind:        "changed",
						},
						{
							Component:   "kubernetes",
							Description: "Kubernetes version requirements changed due to calico update.",
							Kind:        "changed",
						},
					},
					Components: []Component{
						{
							Name:    "calico",
							Version: "1.1.0",
						},
						{
							Name:    "kube-dns",
							Version: "1.0.0",
						},
					},
					Dependencies: []Dependency{
						{
							Name:    "kubernetes",
							Version: "<= 1.7.x",
						},
					},
					Deprecated: false,
					Name:       "kubernetes-operator",
					Time:       time.Unix(10, 5),
					Version:    "0.2.0",
					WIP:        false,
				},
			},
			ErrorMatcher: IsInvalidBundlesError,
		},

		// Test 4 ensures validation of a list of version bundles where any version
		// bundle has no components throws an error.
		{
			Bundles: []Bundle{
				{
					Changelogs: []Changelog{
						{
							Component:   "calico",
							Description: "Calico version updated.",
							Kind:        "changed",
						},
						{
							Component:   "kubernetes",
							Description: "Kubernetes version requirements changed due to calico update.",
							Kind:        "changed",
						},
					},
					Components: []Component{},
					Dependencies: []Dependency{
						{
							Name:    "kubernetes",
							Version: "<= 1.7.x",
						},
					},
					Deprecated: false,
					Name:       "kubernetes-operator",
					Time:       time.Unix(10, 5),
					Version:    "0.1.0",
					WIP:        false,
				},
			},
			ErrorMatcher: IsInvalidBundlesError,
		},

		// Test 5 is the same as 4 but with multiple version bundles.
		{
			Bundles: []Bundle{
				{
					Changelogs: []Changelog{
						{
							Component:   "calico",
							Description: "Calico version updated.",
							Kind:        "changed",
						},
						{
							Component:   "kubernetes",
							Description: "Kubernetes version requirements changed due to calico update.",
							Kind:        "changed",
						},
					},
					Components: []Component{},
					Dependencies: []Dependency{
						{
							Name:    "kubernetes",
							Version: "<= 1.7.x",
						},
					},
					Deprecated: false,
					Name:       "kubernetes-operator",
					Time:       time.Unix(10, 5),
					Version:    "0.1.0",
					WIP:        false,
				},
				{
					Changelogs: []Changelog{
						{
							Component:   "calico",
							Description: "Calico version updated.",
							Kind:        "changed",
						},
						{
							Component:   "kubernetes",
							Description: "Kubernetes version requirements changed due to calico update.",
							Kind:        "changed",
						},
					},
					Components: []Component{
						{
							Name:    "calico",
							Version: "1.1.0",
						},
						{
							Name:    "kube-dns",
							Version: "1.0.0",
						},
					},
					Dependencies: []Dependency{
						{
							Name:    "kubernetes",
							Version: "<= 1.7.x",
						},
					},
					Deprecated: false,
					Name:       "kubernetes-operator",
					Time:       time.Unix(10, 5),
					Version:    "0.2.0",
					WIP:        false,
				},
			},
			ErrorMatcher: IsInvalidBundlesError,
		},

		// Test 6 ensures validation of a list of version bundles where any version
		// bundle has no dependency does not throw an error.
		{
			Bundles: []Bundle{
				{
					Changelogs: []Changelog{
						{
							Component:   "calico",
							Description: "Calico version updated.",
							Kind:        "changed",
						},
					},
					Components: []Component{
						{
							Name:    "calico",
							Version: "1.1.0",
						},
						{
							Name:    "kube-dns",
							Version: "1.0.0",
						},
					},
					Dependencies: []Dependency{},
					Deprecated:   false,
					Name:         "kubernetes-operator",
					Time:         time.Unix(10, 5),
					Version:      "0.1.0",
					WIP:          false,
				},
			},
			ErrorMatcher: nil,
		},

		// Test 7 is the same as 6 but with multiple version bundles.
		{
			Bundles: []Bundle{
				{
					Changelogs: []Changelog{
						{
							Component:   "calico",
							Description: "Calico version updated.",
							Kind:        "changed",
						},
					},
					Components: []Component{
						{
							Name:    "calico",
							Version: "1.1.0",
						},
						{
							Name:    "kube-dns",
							Version: "1.0.0",
						},
					},
					Dependencies: []Dependency{},
					Deprecated:   false,
					Name:         "kubernetes-operator",
					Time:         time.Unix(10, 5),
					Version:      "0.1.0",
					WIP:          false,
				},
				{
					Changelogs: []Changelog{
						{
							Component:   "calico",
							Description: "Calico version updated.",
							Kind:        "changed",
						},
					},
					Components: []Component{
						{
							Name:    "calico",
							Version: "1.1.0",
						},
						{
							Name:    "kube-dns",
							Version: "1.0.0",
						},
					},
					Dependencies: []Dependency{
						{
							Name:    "kubernetes",
							Version: "<= 1.7.x",
						},
					},
					Deprecated: false,
					Name:       "kubernetes-operator",
					Time:       time.Unix(10, 5),
					Version:    "0.2.0",
					WIP:        false,
				},
			},
			ErrorMatcher: nil,
		},

		// Test 8 ensures validation of a list of version bundles not having the
		// same name throws an error.
		{
			Bundles: []Bundle{
				{
					Changelogs: []Changelog{
						{
							Component:   "calico",
							Description: "Calico version updated.",
							Kind:        "changed",
						},
					},
					Components: []Component{
						{
							Name:    "calico",
							Version: "1.1.0",
						},
						{
							Name:    "kube-dns",
							Version: "1.0.0",
						},
					},
					Dependencies: []Dependency{},
					Deprecated:   false,
					Name:         "kubernetes-operator",
					Time:         time.Unix(10, 5),
					Version:      "0.1.0",
					WIP:          false,
				},
				{
					Changelogs: []Changelog{
						{
							Component:   "calico",
							Description: "Calico version updated.",
							Kind:        "changed",
						},
					},
					Components: []Component{
						{
							Name:    "calico",
							Version: "1.1.0",
						},
						{
							Name:    "kube-dns",
							Version: "1.0.0",
						},
					},
					Dependencies: []Dependency{
						{
							Name:    "kubernetes",
							Version: "<= 1.7.x",
						},
					},
					Deprecated: false,
					Name:       "ingress-operator",
					Time:       time.Unix(10, 5),
					Version:    "0.2.0",
					WIP:        false,
				},
			},
			ErrorMatcher: IsInvalidBundlesError,
		},

		// Test 9 ensures validation of a list of version bundles having duplicated
		// version bundles throws an error.
		{
			Bundles: []Bundle{
				{
					Changelogs: []Changelog{
						{
							Component:   "calico",
							Description: "Calico version updated.",
							Kind:        "changed",
						},
					},
					Components: []Component{
						{
							Name:    "calico",
							Version: "1.1.0",
						},
						{
							Name:    "kube-dns",
							Version: "1.0.0",
						},
					},
					Dependencies: []Dependency{},
					Deprecated:   false,
					Name:         "kubernetes-operator",
					Time:         time.Unix(10, 5),
					Version:      "0.1.0",
					WIP:          false,
				},
				{
					Changelogs: []Changelog{
						{
							Component:   "calico",
							Description: "Calico version updated.",
							Kind:        "changed",
						},
					},
					Components: []Component{
						{
							Name:    "calico",
							Version: "1.1.0",
						},
						{
							Name:    "kube-dns",
							Version: "1.0.0",
						},
					},
					Dependencies: []Dependency{},
					Deprecated:   false,
					Name:         "kubernetes-operator",
					Time:         time.Unix(10, 5),
					Version:      "0.1.0",
					WIP:          false,
				},
			},
			ErrorMatcher: IsInvalidBundlesError,
		},

		// Test 10 ensures validation of a list of version bundles having the same
		// version throws an error.
		{
			Bundles: []Bundle{
				{
					Changelogs: []Changelog{
						{
							Component:   "calico",
							Description: "Calico version updated.",
							Kind:        "changed",
						},
					},
					Components: []Component{
						{
							Name:    "calico",
							Version: "1.1.0",
						},
						{
							Name:    "kube-dns",
							Version: "1.0.0",
						},
					},
					Dependencies: []Dependency{},
					Deprecated:   false,
					Name:         "kubernetes-operator",
					Time:         time.Unix(10, 5),
					Version:      "0.1.0",
					WIP:          false,
				},
				{
					Changelogs: []Changelog{
						{
							Component:   "kube-dns",
							Description: "Kube-DNS version updated.",
							Kind:        "changed",
						},
					},
					Components: []Component{
						{
							Name:    "calico",
							Version: "1.1.0",
						},
						{
							Name:    "kube-dns",
							Version: "1.1.0",
						},
					},
					Dependencies: []Dependency{},
					Deprecated:   false,
					Name:         "kubernetes-operator",
					Time:         time.Unix(20, 10),
					Version:      "0.1.0",
					WIP:          false,
				},
			},
			ErrorMatcher: IsInvalidBundlesError,
		},

		// Test 11 ensures validation of a list of version bundles in which a newer
		// version bundle (time) has a lower version number throws an error.
		{
			Bundles: []Bundle{
				{
					Changelogs: []Changelog{
						{
							Component:   "calico",
							Description: "Calico version updated.",
							Kind:        "changed",
						},
					},
					Components: []Component{
						{
							Name:    "calico",
							Version: "1.1.0",
						},
						{
							Name:    "kube-dns",
							Version: "1.0.0",
						},
					},
					Dependencies: []Dependency{},
					Deprecated:   false,
					Name:         "kubernetes-operator",
					Time:         time.Unix(10, 5),
					Version:      "0.1.0",
					WIP:          false,
				},
				{
					Changelogs: []Changelog{
						{
							Component:   "kube-dns",
							Description: "Kube-DNS version updated.",
							Kind:        "changed",
						},
					},
					Components: []Component{
						{
							Name:    "calico",
							Version: "1.1.0",
						},
						{
							Name:    "kube-dns",
							Version: "1.1.0",
						},
					},
					Dependencies: []Dependency{},
					Deprecated:   false,
					Name:         "kubernetes-operator",
					Time:         time.Unix(20, 10),
					Version:      "0.0.9",
					WIP:          false,
				},
			},
			ErrorMatcher: IsInvalidBundlesError,
		},
	}

	for i, tc := range testCases {
		err := ValidateBundles(tc.Bundles).Validate()
		if tc.ErrorMatcher != nil {
			if !tc.ErrorMatcher(err) {
				t.Fatalf("test %d expected %#v got %#v", i, true, false)
			}
		} else if err != nil {
			t.Fatalf("test %d expected %#v got %#v", i, nil, err)
		}
	}
}

func Test_AggregatedBundles_Validate(t *testing.T) {
	testCases := []struct {
		AggregatedBundles [][]Bundle
		ErrorMatcher      func(err error) bool
	}{
		// Test 0 ensures that validating an aggregated list of version bundles
		// having different lengths throws an error.
		{
			AggregatedBundles: [][]Bundle{
				{
					{
						Changelogs: []Changelog{
							{
								Component:   "calico",
								Description: "Calico version updated.",
								Kind:        "changed",
							},
							{
								Component:   "kubernetes",
								Description: "Kubernetes version requirements changed due to calico update.",
								Kind:        "changed",
							},
						},
						Components: []Component{
							{
								Name:    "calico",
								Version: "1.1.0",
							},
							{
								Name:    "kube-dns",
								Version: "1.0.0",
							},
						},
						Dependencies: []Dependency{
							{
								Name:    "kubernetes",
								Version: "<= 1.7.x",
							},
						},
						Deprecated: false,
						Name:       "kubernetes-operator",
						Time:       time.Unix(10, 5),
						Version:    "0.1.0",
						WIP:        false,
					},
					{
						Changelogs: []Changelog{
							{
								Component:   "etcd",
								Description: "Etcd version updated.",
								Kind:        "changed",
							},
							{
								Component:   "kubernetes",
								Description: "Kubernetes version updated.",
								Kind:        "changed",
							},
						},
						Components: []Component{
							{
								Name:    "etcd",
								Version: "3.2.0",
							},
							{
								Name:    "kubernetes",
								Version: "1.7.1",
							},
						},
						Dependencies: []Dependency{},
						Name:         "cloud-config-operator",
						Deprecated:   false,
						Time:         time.Unix(20, 15),
						Version:      "0.2.0",
						WIP:          false,
					},
				},
				{
					{
						Changelogs: []Changelog{
							{
								Component:   "etcd",
								Description: "Etcd version updated.",
								Kind:        "changed",
							},
							{
								Component:   "kubernetes",
								Description: "Kubernetes version updated.",
								Kind:        "changed",
							},
						},
						Components: []Component{
							{
								Name:    "etcd",
								Version: "3.2.0",
							},
							{
								Name:    "kubernetes",
								Version: "1.7.1",
							},
						},
						Dependencies: []Dependency{},
						Deprecated:   false,
						Name:         "cloud-config-operator",
						Time:         time.Unix(20, 15),
						Version:      "0.2.0",
						WIP:          false,
					},
				},
			},
			ErrorMatcher: IsInvalidAggregatedBundlesError,
		},

		// Test 1 ensures that validating an aggregated list of version bundles
		// having the same version bundles throws an error.
		{
			AggregatedBundles: [][]Bundle{
				{
					{
						Changelogs: []Changelog{
							{
								Component:   "calico",
								Description: "Calico version updated.",
								Kind:        "changed",
							},
							{
								Component:   "kubernetes",
								Description: "Kubernetes version requirements changed due to calico update.",
								Kind:        "changed",
							},
						},
						Components: []Component{
							{
								Name:    "calico",
								Version: "1.1.0",
							},
							{
								Name:    "kube-dns",
								Version: "1.0.0",
							},
						},
						Dependencies: []Dependency{
							{
								Name:    "kubernetes",
								Version: "<= 1.7.x",
							},
						},
						Deprecated: false,
						Name:       "kubernetes-operator",
						Time:       time.Unix(10, 5),
						Version:    "0.1.0",
						WIP:        false,
					},
					{
						Changelogs: []Changelog{
							{
								Component:   "etcd",
								Description: "Etcd version updated.",
								Kind:        "changed",
							},
							{
								Component:   "kubernetes",
								Description: "Kubernetes version updated.",
								Kind:        "changed",
							},
						},
						Components: []Component{
							{
								Name:    "etcd",
								Version: "3.2.0",
							},
							{
								Name:    "kubernetes",
								Version: "1.7.1",
							},
						},
						Dependencies: []Dependency{},
						Deprecated:   false,
						Name:         "cloud-config-operator",
						Time:         time.Unix(20, 15),
						Version:      "0.2.0",
						WIP:          false,
					},
				},
				{
					{
						Changelogs: []Changelog{
							{
								Component:   "calico",
								Description: "Calico version updated.",
								Kind:        "changed",
							},
							{
								Component:   "kubernetes",
								Description: "Kubernetes version requirements changed due to calico update.",
								Kind:        "changed",
							},
						},
						Components: []Component{
							{
								Name:    "calico",
								Version: "1.1.0",
							},
							{
								Name:    "kube-dns",
								Version: "1.0.0",
							},
						},
						Dependencies: []Dependency{
							{
								Name:    "kubernetes",
								Version: "<= 1.7.x",
							},
						},
						Deprecated: false,
						Name:       "kubernetes-operator",
						Time:       time.Unix(10, 5),
						Version:    "0.1.0",
						WIP:        false,
					},
					{
						Changelogs: []Changelog{
							{
								Component:   "etcd",
								Description: "Etcd version updated.",
								Kind:        "changed",
							},
							{
								Component:   "kubernetes",
								Description: "Kubernetes version updated.",
								Kind:        "changed",
							},
						},
						Components: []Component{
							{
								Name:    "etcd",
								Version: "3.2.0",
							},
							{
								Name:    "kubernetes",
								Version: "1.7.1",
							},
						},
						Dependencies: []Dependency{},
						Deprecated:   false,
						Name:         "cloud-config-operator",
						Time:         time.Unix(20, 15),
						Version:      "0.2.0",
						WIP:          false,
					},
				},
			},
			ErrorMatcher: IsInvalidAggregatedBundlesError,
		},
	}

	for i, tc := range testCases {
		err := ValidateAggregatedBundles(tc.AggregatedBundles).Validate()
		if tc.ErrorMatcher != nil {
			if !tc.ErrorMatcher(err) {
				t.Fatalf("test %d expected %#v got %#v", i, true, false)
			}
		} else if err != nil {
			t.Fatalf("test %d expected %#v got %#v", i, nil, err)
		}
	}
}
