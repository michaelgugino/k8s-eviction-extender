package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"

	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	klog.Errorf(fmt.Sprintf("handling request: %s", body))
    klog.Errorf("attempting to read header")
	// verify the content type is accurate
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		klog.Errorf("contentType=%s, expect application/json", contentType)
		return
	}

	klog.V(2).Info(fmt.Sprintf("handling request: %s", body))

    klog.Errorf("allocate ar")
	// The AdmissionReview that was sent to the webhook
	requestedAdmissionReview := admissionv1.AdmissionReview{}

	// The AdmissionReview that will be returned
	responseAdmissionReview := admissionv1.AdmissionReview{}
    klog.Errorf("attempting deserialize")
	deserializer := codecs.UniversalDeserializer()
	if _, _, err := deserializer.Decode(body, nil, &requestedAdmissionReview); err != nil {
		klog.Error(err)
		responseAdmissionReview.Response = toAdmissionResponse(err)
	} else {
		// pass to admitFunc
        klog.Errorf("attempting to admit")
		responseAdmissionReview.Response = admit(requestedAdmissionReview)
	}
    klog.Errorf("attempting to return uid")
	// Return the same UID
	responseAdmissionReview.Response.UID = requestedAdmissionReview.Request.UID

	klog.V(2).Info(fmt.Sprintf("sending response: %v", responseAdmissionReview.Response))

	respBytes, err := json.Marshal(responseAdmissionReview)
	if err != nil {
		klog.Error(err)
	}
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
