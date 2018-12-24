/*
modification history
--------------------
2018/12/21, by lovejoy, create
*/

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type admitFunc func(AdmissionReview) *AdmissionResponse

func serve(w http.ResponseWriter, r *http.Request, admit admitFunc) {
	var body []byte
	if r.Body != nil {
		if data, err := ioutil.ReadAll(r.Body); err == nil {
			body = data
		}
	}
	fmt.Printf("handling request: %v\n", string(body))

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
	serve(w, r, alwaysDeny)
}

// Deny all requests made to this function.
func alwaysDeny(ar AdmissionReview) *AdmissionResponse {
	reviewResponse := AdmissionResponse{}
	reviewResponse.UID = ar.Request.UID
	reviewResponse.Allowed = false
	reviewResponse.Result = &Status{Message: "this webhook denies all requests"}
	return &reviewResponse
}
