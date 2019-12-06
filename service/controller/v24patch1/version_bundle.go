package v24patch1

import (
	"github.com/giantswarm/versionbundle"
)

func VersionBundle() versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{
				Component:   "TODO",
				Description: "TODO",
				Kind:        versionbundle.KindChanged,
				URLs: []string{
					"TODO",
				},
			},
		},
		Components: []versionbundle.Component{
			{
				Name:    "calico",
				Version: "3.8.2",
			},
			{
				Name:    "containerlinux",
				Version: "2135.4.0",
			},
			{
				Name:    "docker",
				Version: "18.06.1",
			},
			{
				Name:    "etcd",
				Version: "3.3.13",
			},
			{
				Name:    "kubernetes",
				Version: "1.14.9",
			},
		},
		Name:    "kvm-operator",
<<<<<<< HEAD
<<<<<<< HEAD
		Version: "3.8.1",
=======
		Version: "3.8.0",
>>>>>>> c4c6c79d... copy v24 to v24patch1
=======
		Version: "3.8.1",
>>>>>>> d6f149c2... wire v24patch1
	}
}
