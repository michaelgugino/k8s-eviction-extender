FROM docker.io/golang:1.16 as builder
WORKDIR /go/src/github.com/michaelgugino/k8s-eviction-extender
COPY . .
RUN NO_DOCKER=1 make build

# FROM registry.ci.openshift.org/openshift/origin-v4.0:base

FROM gcr.io/distroless/static:nonroot
WORKDIR /

COPY --from=builder /go/src/github.com/michaelgugino/k8s-eviction-extender/bin/k8s-eviction-extender .
