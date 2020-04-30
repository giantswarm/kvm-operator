package project

import (
	"github.com/giantswarm/versionbundle"
)

func NewVersionBundle() versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{
				Component:   "containerlinux",
				Description: "Deprecate CoreOS and move to Flatcar Linux.",
				Kind:        versionbundle.KindChanged,
				URLs: []string{
					"https://github.com/giantswarm/kvm-operator/pull/825",
				},
			},
		},
		Name:    Name(),
		Version: Version(),
	}
}
