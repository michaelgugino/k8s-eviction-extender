/*
Copyright 2021 Red Hat, Inc. and/or its affiliates

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package server

import (
    "context"

	admissionv1 "k8s.io/api/admission/v1"
    policyv1beta1 "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"

    "k8s.io/client-go/kubernetes"
)

const (
    PreventEvictAnnotation = "k8s-eviction-extender.michaelgugino.github.com/no-evict"
    EvictionRequested = "k8s-eviction-extender.michaelgugino.github.com/evict-requested"
)

type evictionExtender struct {
	kclient *kubernetes.Clientset
}

func (ex evictionExtender) evictionCreate(ar admissionv1.AdmissionReview) *admissionv1.AdmissionResponse {
	klog.Errorf("admitting eviction")

    evictResource := metav1.GroupVersionKind{Group: "policy", Version: "v1beta1", Kind: "Eviction"}
	if ar.Request == nil || ar.Request.RequestKind == nil || *ar.Request.RequestKind != evictResource {
		klog.Errorf("expect requestKind to be %s", evictResource)
		return nil
	}

	var raw []byte
	if ar.Request.Operation == admissionv1.Delete {
		raw = ar.Request.OldObject.Raw
	} else {
		raw = ar.Request.Object.Raw
	}
	evictionRequest := policyv1beta1.Eviction{}
	deserializer := codecs.UniversalDeserializer()
	if _, _, err := deserializer.Decode(raw, nil, &evictionRequest); err != nil {
		klog.Error(err)
		return toAdmissionResponse(err)
	}

    klog.Errorf("Getting pod")

    name := evictionRequest.Name
    namespace := evictionRequest.Namespace

    reviewResponse := admissionv1.AdmissionResponse{}

    pod, err := ex.kclient.CoreV1().Pods(namespace).Get(context.TODO(), name, metav1.GetOptions{})
    if err != nil {
        reviewResponse.Allowed = false
    	reviewResponse.Result = &metav1.Status{
    		Reason: "Unable to get pod",
            Code: 429,
    	}
        return &reviewResponse
    }

    if _, exists := pod.ObjectMeta.Annotations[PreventEvictAnnotation]; exists {
        if _, exists := pod.ObjectMeta.Annotations[EvictionRequested]; !exists {
            p2 := pod.DeepCopy()
            metav1.SetMetaDataAnnotation(&p2.ObjectMeta, EvictionRequested, "")
            _, err := ex.kclient.CoreV1().Pods(namespace).Update(context.TODO(), p2, metav1.UpdateOptions{})
            if err != nil {
                klog.Errorf("Error updating pod: %v", err)
            }
        }
        reviewResponse.Allowed = false
    	reviewResponse.Result = &metav1.Status{
    		Reason: "Eviction not allowed by PreventEvictAnnotation",
            Code: 429,
    	}
        return &reviewResponse
	}

	reviewResponse.Allowed = true

	return &reviewResponse
}
