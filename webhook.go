/*
modification history
--------------------
2018/12/21, by lovejoy, create
*/

package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

type admitFunc func(AdmissionReview) *AdmissionResponse

type policy struct {
	User            string `json:"user"`
	NamespacePrefix string `json:"namespace-prefix"`
}

type policyList []*policy

var policys policyList

func namespacePrefixLimit(ar AdmissionReview) *AdmissionResponse {

	reviewResponse := AdmissionResponse{}
	reviewResponse.UID = ar.Request.UID
	user := ar.Request.UserInfo.Username
	if user == "system:unsecured" {
		reviewResponse.Allowed = false
		reviewResponse.Result = &Status{Message: "please use https connection"} //if true this didn't show up
		return &reviewResponse
	}

	allowd := false
	for _, p := range policys {
		if ar.Request.Resource.Resource == "namespaces" {
			if p.User == user && strings.HasPrefix(ar.Request.Object.MetaData.Name, p.NamespacePrefix) {
				allowd = true
				break
			}

		}
		if p.User == user && strings.HasPrefix(ar.Request.Namespace, p.NamespacePrefix) {
			allowd = true
			break
		}

	}

	if user == "admin" || strings.HasPrefix(user, "system:") {
		allowd = true //admin and other system user always allow
	}
	if ar.Request.Resource.Resource != "namespaces" && ar.Request.Namespace == "" {
		allowd = true
	}
	reviewResponse.Allowed = allowd
	if !allowd {
		reviewResponse.Result = &Status{Message: "the user and nampace didn't match any policy"} //if true this didn't show up
	}
	return &reviewResponse
}
func newPolicyListFromFile(path string) (policyList, error) {
	file, err := os.Open(path)
	if err != nil {

		return nil, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	pl := make(policyList, 0)
	for scanner.Scan() {
		b := scanner.Bytes()
		// skip comment lines and blank lines
		trimmed := strings.TrimSpace(string(b))
		if len(trimmed) == 0 || strings.HasPrefix(trimmed, "#") {
			continue
		}
		p := &policy{}
		if err := json.Unmarshal(b, p); err != nil {
			return nil, err
		}
		if p.NamespacePrefix == "" {
			return nil, errors.New("namespace-prefix is empty")
		}
		pl = append(pl, p)
	}
	return pl, nil
}

// Deny all requests made to this function.
func alwaysDeny(ar AdmissionReview) *AdmissionResponse {
	reviewResponse := AdmissionResponse{}
	reviewResponse.UID = ar.Request.UID
	reviewResponse.Allowed = false
	reviewResponse.Result = &Status{Message: "this webhook denies all requests"}
	return &reviewResponse
}

func serve(w http.ResponseWriter, r *http.Request, admit admitFunc) {
	var body []byte
	if r.Body != nil {
		if data, err := ioutil.ReadAll(r.Body); err == nil {
			body = data
		}
	}
	fmt.Printf("handling request: %v", string(body))

	var reviewResponse *AdmissionResponse
	ar := AdmissionReview{}
	if err := json.Unmarshal(body, &ar); err != nil {
		fmt.Println(err)
		reviewResponse = toAdmissionResponse(err)
	} else {
		reviewResponse = admit(ar)
	}
	resp, _ := json.Marshal(reviewResponse)
	fmt.Printf("sending response: %v\n", string(resp))

	response := AdmissionReview{}
	if reviewResponse != nil {
		response.Response = reviewResponse
		response.Response.UID = ar.Request.UID
	}
	resp, err := json.Marshal(response)
	if err != nil {
		fmt.Println(err)
	}
	if _, err := w.Write(resp); err != nil {
		fmt.Println(err)
	}
}
func validate(w http.ResponseWriter, r *http.Request) {
	serve(w, r, namespacePrefixLimit)
}
