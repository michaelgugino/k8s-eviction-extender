# k8s-eviction-extender

This project configures a validating webhook to intercept eviction requests
made against pods on the cluster.  These requests are checked against pod
annotations to determine whether or not the pod can be evicted.

To prevent eviction, add the annotation to the pod:
`k8s-eviction-extender.michaelgugino.github.com/no-evict`

No eviction will take place while this annotation is present.

When a pod eviction request is created, the annotation
`k8s-eviction-extender.michaelgugino.github.com/evict-requested`
will be added.  Another component should watch pods for this annotation and
remove the prevent annotation when appropriate.

# Install

## Warning

This webhook will mutate pods by adding an annotation.  The webhook will be
exposed open to requests from any pods (and possibly elsewhere) running on the
cluster by default.  Please follow the steps here to secure the webhook server:

https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/

## Install

Tested with OpenShift 4.7 and k8s 1.20.

When installing on OpenShift, you can kubectl apply the assets directory of
this project.

When installing on kubernetes, you will need to provide your own TLS cert key
pair and insert an appropriate CA into the validatingwebhook configuration.
