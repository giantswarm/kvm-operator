package versionbundle

import (
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
