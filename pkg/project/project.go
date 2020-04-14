package project

var (
	description string = "The kvm-operator handles Kubernetes clusters running on a Kubernetes cluster."
	gitSHA             = "n/a"
	name        string = "kvm-operator"
	source      string = "https://github.com/giantswarm/kvm-operator"
	version            = "3.9.1-dev"
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
