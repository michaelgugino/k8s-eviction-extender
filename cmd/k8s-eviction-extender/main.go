package main

import (
  "flag"

  "github.com/michaelgugino/k8s-eviction-extender/pkg/server"
  "k8s.io/klog"

  "k8s.io/client-go/kubernetes"
  "k8s.io/client-go/rest"
)

func main() {

    klog.InitFlags(nil)
    flag.Set("logtostderr", "true")

    // creates the in-cluster config
    config, err := rest.InClusterConfig()
    if err != nil {
        panic(err.Error())
    }
    // creates the clientset
    kclient, err := kubernetes.NewForConfig(config)
    if err != nil {
        panic(err.Error())
    }

    webhookPort := flag.Int("webhook-port", 8443,
		"Webhook Server port, only used when webhook-enabled is true.")

	webhookCert := flag.String("webhook-cert", "/etc/secret-volume/tls.crt",
		"Webhook cert dir, only used when webhook-enabled is true.")

    webhookKey := flag.String("webhook-key", "/etc/secret-volume/tls.key",
		"Webhook cert dir, only used when webhook-enabled is true.")

    flag.Parse()

    server.Serve(*webhookCert, *webhookKey, *webhookPort, kclient)
}
