VERSION := $(shell cat pkg/project/project.go | grep version | tr -s ' ' | cut -d ' ' -f 3 | sed 's/"//g')
VERSION_HYPHENATED := $(shell cat pkg/project/project.go | grep version | tr -s ' ' | cut -d ' ' -f 3 | sed 's/"//g;s/\./-/g')

.PHONY: okteto-up
## okteto-up: start okteto against the current kubeconfig context
okteto-up:
	cat okteto.yml | sed 's/$$VERSION/$(VERSION)/' > .okteto.yml
	kubectl patch psp -n giantswarm kvm-operator-$(VERSION_HYPHENATED)-psp -p '{"spec":{"runAsGroup":{"ranges":null,"rule":"RunAsAny"},"runAsUser":{"rule":"RunAsAny"},"volumes":["secret","configMap","hostPath","persistentVolumeClaim","emptyDir"]}}'
	okteto up -f .okteto.yml

.PHONY: okteto-down
## okteto-down: stop and clean up an okteto session started by okteto-up
okteto-down:
	okteto down -f .okteto.yml
	kubectl patch psp -n giantswarm kvm-operator-$(VERSION_HYPHENATED)-psp -p '{"spec":{"runAsGroup":{"ranges": [{"max":65535, "min":1}],"rule":"MustRunAs"},"runAsUser":{"rule":"MustRunAsNonRoot"},"volumes":["secret","configMap"]}}'
	rm .okteto.yml
