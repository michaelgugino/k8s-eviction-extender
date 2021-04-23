FROM registry.ci.openshift.org/openshift/release:golang-1.15 AS builder
WORKDIR /go/src/github.com/michaelgugino/k8s-eviction-extender
COPY . .
RUN NO_DOCKER=1 make build

FROM registry.ci.openshift.org/openshift/origin-v4.0:base
COPY --from=builder /go/src/github.com/openshift/machine-api-operator/bin/machine-api-operator .
COPY --from=builder /go/src/github.com/openshift/machine-api-operator/bin/machineset ./machineset-controller
