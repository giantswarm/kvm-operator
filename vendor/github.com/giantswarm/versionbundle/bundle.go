package versionbundle

import (
	"encoding/json"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/giantswarm/microerror"
)

// Bundle represents a single version bundle exposed by an authority. An
// authority might exposes mutliple version bundles using the Capability
// structure. Version bundles are aggregated into a merged structure represented
// by the Aggregation structure. Also see the Aggregate function.
type Bundle struct {
	// Changelogs describe what changes are introduced by the version bundle. Each
	// version bundle must have at least one changelog entry.
	//
	// NOTE that once this property is set it must never change again.
	Changelogs []Changelog `json:"changelogs" yaml:"changelogs"`
	// Components describe the components an authority exposes. Functionality of
	// components listed here is guaranteed to be implemented in the according
	// versions.
	//
	// NOTE that once this property is set it must never change again.
	Components []Component `json:"components" yaml:"components"`
	// Dependencies describe which components other authorities expose have to be
	// available to be able to guarantee functionality this authority implements.
	//
	// NOTE that once this property is set it must never change again.
	Dependencies []Dependency `json:"dependency" yaml:"dependency"`
	// Deprecated defines a version bundle to be deprecated. Deprecated version
	// bundles are not intended to be mainatined anymore. Further usage of a
	// deprecated version bundle should be omitted.
	Deprecated bool `json:"deprecated" yaml:"deprecated"`
	// Name is the name of the authority exposing the version bundle.
	//
	// NOTE that once this property is set it must never change again.
	Name string `json:"name" yaml:"name"`
	// Time describes the time this version bundle got introduced.
	//
	// NOTE that once this property is set it must never change again.
	Time time.Time `json:"time" yaml:"time"`
	// Version describes the version of the version bundle. Versions of version
	// bundles must be semver versions. Versions must not be duplicated. Versions
	// should be incremented gradually.
	//
	// NOTE that once this property is set it must never change again.
	Version string `json:"version" yaml:"version"`
	// WIP describes if a version bundle is being developed. Usage of a version
	// bundle still being developed should be omitted.
	WIP bool `json:"wip" yaml:"wip"`
}

func (b Bundle) Validate() error {
	if len(b.Changelogs) == 0 {
		return microerror.Maskf(invalidBundleError, "changelogs must not be empty")
	}
	for _, c := range b.Changelogs {
		err := c.Validate()
		if err != nil {
			return microerror.Maskf(invalidBundleError, err.Error())
		}
	}

	if len(b.Components) == 0 {
		return microerror.Maskf(invalidBundleError, "components must not be empty")
	}
	for _, c := range b.Components {
		err := c.Validate()
		if err != nil {
			return microerror.Maskf(invalidBundleError, err.Error())
		}
	}

	for _, d := range b.Dependencies {
		err := d.Validate()
		if err != nil {
			return microerror.Maskf(invalidBundleError, err.Error())
		}
	}

	var emptyTime time.Time
	if b.Time == emptyTime {
		return microerror.Maskf(invalidBundleError, "time must not be empty")
	}

	if b.Name == "" {
		return microerror.Maskf(invalidBundleError, "name must not be empty")
	}

	versionSplit := strings.Split(b.Version, ".")
	if len(versionSplit) != 3 {
		return microerror.Maskf(invalidBundleError, "version format must be '<major>.<minor>.<patch>'")
	}

	if !isPositiveNumber(versionSplit[0]) {
		return microerror.Maskf(invalidBundleError, "major version must be positive number")
	}

	if !isPositiveNumber(versionSplit[1]) {
		return microerror.Maskf(invalidBundleError, "minor version must be positive number")
	}

	if !isPositiveNumber(versionSplit[2]) {
		return microerror.Maskf(invalidBundleError, "patch version must be positive number")
	}

	return nil
}

type SortBundlesByName []Bundle

func (b SortBundlesByName) Len() int           { return len(b) }
func (b SortBundlesByName) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b SortBundlesByName) Less(i, j int) bool { return b[i].Name < b[j].Name }

type SortBundlesByVersion []Bundle

func (b SortBundlesByVersion) Len() int           { return len(b) }
func (b SortBundlesByVersion) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b SortBundlesByVersion) Less(i, j int) bool { return b[i].Version < b[j].Version }

type SortBundlesByTime []Bundle

func (b SortBundlesByTime) Len() int           { return len(b) }
func (b SortBundlesByTime) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b SortBundlesByTime) Less(i, j int) bool { return b[i].Time.UnixNano() < b[j].Time.UnixNano() }

// ValidateBundles is a plain validation type for a list of version bundles. A
// list of version bundles is exposed by authorities. Lists of version bundles
// of multiple authorities are aggregated and grouped to reflect distributions.
type ValidateBundles []Bundle

func (b ValidateBundles) Copy() ValidateBundles {
	raw, err := json.Marshal(b)
	if err != nil {
		panic(err)
	}

	var copy ValidateBundles
	err = json.Unmarshal(raw, &copy)
	if err != nil {
		panic(err)
	}

	return copy
}

func (b ValidateBundles) Validate() error {
	if len(b) == 0 {
		return microerror.Maskf(invalidBundlesError, "version bundles must not be empty")
	}

	if b.hasDuplicatedVersions() {
		return microerror.Maskf(invalidBundlesError, "version bundle versions must be unique")
	}

	b1 := b.Copy()
	b2 := b.Copy()
	sort.Sort(SortBundlesByTime(b1))
	sort.Sort(SortBundlesByVersion(b2))
	if !reflect.DeepEqual(b1, b2) {
		return microerror.Maskf(invalidBundlesError, "version bundle versions must always increment")
	}

	for _, bundle := range b {
		err := bundle.Validate()
		if err != nil {
			return microerror.Maskf(invalidBundlesError, err.Error())
		}
	}

	bundleName := b[0].Name
	for _, bundle := range b {
		if bundle.Name != bundleName {
			return microerror.Maskf(invalidBundlesError, "name must be the same for all version bundles")
		}
	}

	return nil
}

func (b ValidateBundles) hasDuplicatedVersions() bool {
	for _, b1 := range b {
		var seen int

		for _, b2 := range b {
			if b1.Version == b2.Version {
				seen++

				if seen >= 2 {
					return true
				}
			}
		}
	}

	return false
}

// ValidateAggregatedBundles is a plain validation type for aggregated lists of
// version bundles. Lists of version bundles reflect distributions.
type ValidateAggregatedBundles [][]Bundle

func (b ValidateAggregatedBundles) Validate() error {
	if len(b) != 0 {
		l := len(b[0])
		for _, group := range b {
			if l != len(group) {
				return microerror.Maskf(invalidAggregatedBundlesError, "number of version bundles within aggregated version bundles must be equal")
			}
		}
	}

	if b.hasDuplicatedAggregatedBundles() {
		return microerror.Maskf(invalidAggregatedBundlesError, "version bundles within aggregated version bundles must be unique")
	}

	return nil
}

func (b ValidateAggregatedBundles) hasDuplicatedAggregatedBundles() bool {
	for _, b1 := range b {
		var seen int

		for _, b2 := range b {
			if reflect.DeepEqual(b1, b2) {
				seen++

				if seen >= 2 {
					return true
				}
			}
		}
	}

	return false
}
