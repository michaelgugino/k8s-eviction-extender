# k8s-eviction-extender

This project configures a validating webhook to intercept eviction requests
made against pods on the cluster.  These requests are checked against pod
annotations to determine whether or not the pod can be evicted.
