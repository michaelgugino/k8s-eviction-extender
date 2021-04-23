package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"

	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog"
	"k8s.io/client-go/kubernetes"
	// TODO: try this library to see if it generates correct json patch
	// https://github.com/mattbaird/jsonpatch
)

const numPrivNamespaces int = 4
// privileged namespaces we allow; should be regex.
// If you adjust this, be sure to update numPrivNamespaces and associated
// test matrix in server_test.go.
var allowedNameSpaces = [numPrivNamespaces]string {"^kube-*", "^openshift-*", "^default$", "^logging$"}

var regList = compileRegex()

var ex evictionExtender

func compileRegex() []*regexp.Regexp {
	var compiledList = make([]*regexp.Regexp, 0)
	var compiledExp *regexp.Regexp
	for _, exp := range allowedNameSpaces {
		compiledExp = regexp.MustCompile(exp)
		compiledList = append(compiledList, compiledExp)
	}
	return compiledList
}

func checkNamespace(namespace string) bool {
	// Returns true if privileged namespace, false otherwise.
	var isMatch bool
	for _, compiled := range regList {
			isMatch = compiled.MatchString(namespace)
			if isMatch {
				return true
			}
	}
	return false
}

// toAdmissionResponse is a helper function to create an AdmissionResponse
// with an embedded error
func toAdmissionResponse(err error) *admissionv1.AdmissionResponse {
	return &admissionv1.AdmissionResponse{
		Result: &metav1.Status{
			Message: err.Error(),
		},
	}
}

// admitFunc is the type we use for all of our validators and mutators
type admitFunc func(admissionv1.AdmissionReview) *admissionv1.AdmissionResponse

// serve handles the http portion of a request prior to handing to an admit
// function
func serve(w http.ResponseWriter, r *http.Request, admit admitFunc) {
    klog.Errorf("attempting to read body")
	var body []byte
	if r.Body != nil {
		if data, err := ioutil.ReadAll(r.Body); err == nil {
			body = data
		}
	}

	// verify the content type is accurate
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		klog.Errorf("contentType=%s, expect application/json", contentType)
		return
	}

	klog.V(2).Info(fmt.Sprintf("handling request: %s", body))

	deserializer := codecs.UniversalDeserializer()
	obj, gvk, err := deserializer.Decode(body, nil, nil)
	if err != nil {
		msg := fmt.Sprintf("Request could not be decoded: %v", err)
		klog.Error(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}


	var responseObj runtime.Object
	switch *gvk {
	/*
	case v1beta1.SchemeGroupVersion.WithKind("AdmissionReview"):
		requestedAdmissionReview, ok := obj.(*v1beta1.AdmissionReview)
		if !ok {
			klog.Errorf("Expected v1beta1.AdmissionReview but got: %T", obj)
			return
		}
		responseAdmissionReview := &v1beta1.AdmissionReview{}
		responseAdmissionReview.SetGroupVersionKind(*gvk)
		responseAdmissionReview.Response = admit.v1beta1(*requestedAdmissionReview)
		responseAdmissionReview.Response.UID = requestedAdmissionReview.Request.UID
		responseObj = responseAdmissionReview
	*/
	case admissionv1.SchemeGroupVersion.WithKind("AdmissionReview"):
		requestedAdmissionReview, ok := obj.(*admissionv1.AdmissionReview)
		if !ok {
			klog.Errorf("Expected admissionv1.AdmissionReview but got: %T", obj)
			return
		}
		responseAdmissionReview := &admissionv1.AdmissionReview{}
		responseAdmissionReview.SetGroupVersionKind(*gvk)
		responseAdmissionReview.Response = admit(*requestedAdmissionReview)
		responseAdmissionReview.Response.UID = requestedAdmissionReview.Request.UID
		responseObj = responseAdmissionReview
	default:
		msg := fmt.Sprintf("Unsupported group version kind: %v", gvk)
		klog.Error(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	klog.V(2).Info(fmt.Sprintf("sending response: %v", responseObj))
	respBytes, err := json.Marshal(responseObj)
	if err != nil {
		klog.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(respBytes); err != nil {
		klog.Error(err)
	}

}

func serveEvictionCreate(w http.ResponseWriter, r *http.Request) {
	serve(w, r, ex.evictionCreate)
}

func Serve(certfile string, keyfile string, port int, kclient *kubernetes.Clientset) {
	var config = Config{CertFile: certfile, KeyFile: keyfile}

	ex = evictionExtender{kclient: kclient}

	http.HandleFunc("/eviction", serveEvictionCreate)

	server := &http.Server{
		Addr:      fmt.Sprintf(":%v", port),
		TLSConfig: configTLS(config),
	}
	klog.Errorf(fmt.Sprintf("starting server on %v", port))
	server.ListenAndServeTLS("", "")
}
