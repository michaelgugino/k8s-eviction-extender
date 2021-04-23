/*
Copyright 2019 Red Hat, Inc. and/or its affiliates

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
	"fmt"
	"testing"
)

type regexTestCase struct {
	testString string
	shouldPass	[numPrivNamespaces]bool
}

var regexTestCases = [...]regexTestCase {
	// ^kube-*
	regexTestCase {testString: "kube-x", shouldPass: [numPrivNamespaces]bool{true, false, false, false}},
	regexTestCase {testString: "testkube-x", shouldPass: [numPrivNamespaces]bool{false, false, false, false}},
	// ^openshift-*
	regexTestCase {testString: "openshift-x", shouldPass: [numPrivNamespaces]bool{false, true, false, false}},
	regexTestCase {testString: "testopenshift-x", shouldPass: [numPrivNamespaces]bool{false, false, false, false}},
	// ^default$
	regexTestCase {testString: "default", shouldPass: [numPrivNamespaces]bool{false, false, true, false}},
	regexTestCase {testString: "defaultx", shouldPass: [numPrivNamespaces]bool{false, false, false, false}},
	// ^logging$
	regexTestCase {testString: "logging", shouldPass: [numPrivNamespaces]bool{false, false, false, true}},
	regexTestCase {testString: "xlogging", shouldPass: [numPrivNamespaces]bool{false, false, false, false}},
}

func TestRegex(t *testing.T) {
	var isMatch bool
	for i, compiled := range regList {
		for _, tCase := range regexTestCases {
			isMatch = compiled.MatchString(tCase.testString)
			if !(isMatch == tCase.shouldPass[i]) {
				t.Fatal(fmt.Sprintf("regex failed. match is %v ; string: %v regex: %v ; expected %v ", isMatch, tCase.testString, allowedNameSpaces[i], tCase.shouldPass[i]))
			}
		}
	}
}

func (rtc *regexTestCase) checkTestIsPriv() bool {
	for _, val := range rtc.shouldPass {
		if val {
			return true
		}
	}
	return false
}

func TestCheckNamespace(t *testing.T) {
	var isPriv bool
	var expectedPriv bool
	for _, tCase := range regexTestCases {
		isPriv = checkNamespace(tCase.testString)
		expectedPriv = tCase.checkTestIsPriv()
		if isPriv != expectedPriv {
			t.Fatal(fmt.Sprintf("Namespace privileged check failed. Found is %v ; string: %v; expected %v ", isPriv, tCase.testString, expectedPriv))
		}
	}
}
