package kauth

import (
	"log"
	"net/http"
)

func init() {
	// webhook used as part of the Kubernetes authentication plugin
	handler, err := NewKubernetesValidate()
	if err != nil {
		log.Fatal(err)
	}
	http.Handle("/kubernetes/validate-token", handler)

	// routes used by users to generate tokens from their google credentials
	http.HandleFunc("/auth/start", startAuthentication)
	http.HandleFunc("/auth/complete", completeAuthentication)
}
