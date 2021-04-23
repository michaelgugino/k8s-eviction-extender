.PHONY: all
all: build test

.PHONY: build
build: k8s-eviction-extender ## Build binaries

.PHONY: k8s-eviction-extender
k8s-eviction-extender:
	./hack/go-build.sh k8s-eviction-extender
