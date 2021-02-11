package project

var (
	description = "The kvm-operator handles Kubernetes clusters running on a Kubernetes cluster."
	gitSHA      = "n/a"
	name        = "kvm-operator"
	source      = "https://github.com/giantswarm/kvm-operator"
	version     = "3.13.1"
)

func Description() string {
	return description
}

func GitSHA() string {
	return gitSHA
}

func Name() string {
	return name
}

func Source() string {
	return source
}

func Version() string {
	return version
}
