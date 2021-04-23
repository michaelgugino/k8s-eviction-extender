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
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"

    "k8s.io/client-go/kubernetes"
)

type evictionExtender struct {
	kclient *kubernetes.Clientset
}

func (ex evictionExtender) evictionCreate(ar admissionv1.AdmissionReview) *admissionv1.AdmissionResponse {
	klog.Errorf("admitting eviction")

    /*
	routeresource := metav1.GroupVersionResource{Group: "route.openshift.io", Version: "v1", Resource: "routes"}
	if ar.Request.Resource != routeresource {
		klog.Errorf("expect resource to be %s, found %v", routeresource, ar.Request.Resource)
		return nil
	}
	reviewResponse := v1beta1.AdmissionResponse{}

	reviewResponse.Allowed = false

	raw := ar.Request.Object.Raw
	route := routeapi.Route{}
	err := json.Unmarshal(raw, &route)
	if err != nil {
		klog.Error(err)
		return toAdmissionResponse(err)
	}
    */
    reviewResponse := admissionv1.AdmissionResponse{}
	reviewResponse.Allowed = false
	reviewResponse.Result = &metav1.Status{
		Reason: "Eviction not allowed",
        Code: 429,
	}
	return &reviewResponse
}
