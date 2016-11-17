package kauth

import (
	"encoding/json"
	"fmt"
	"google.golang.org/appengine"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
)

type TokenSpec struct {
	Token string `json:"token"`
}

type TokenReview struct {
	APIVersion string     `json:"apiVersion"`
	Kind       string     `json:"kind"`
	Spec       *TokenSpec `json:spec`
}

type TokenReviewResponse struct {
	APIVersion string             `json:"apiVersion"`
	Kind       string             `json:"kind"`
	Status     *TokenReviewStatus `json:"status"`
}

type TokenReviewStatus struct {
	Authenticated bool  `json:"authenticated"`
	User          *User `json:"user"`
}

type User struct {
	Username string              `json:"username"`
	Uid      string              `json:"uid"`
	Groups   []string            `json:"groups"`
	Extra    map[string][]string `json:"extra"`
}

func newTokenReviewResponse(user *User) *TokenReviewResponse {
	return &TokenReviewResponse{
		APIVersion: "authentication.k8s.io/v1beta1",
		Kind:       "TokenReview",
		Status: &TokenReviewStatus{
			Authenticated: true,
			User:          user,
		},
	}
}

func validUser(w http.ResponseWriter, user *User) {
	tokenResponse := newTokenReviewResponse(user)
	encoder := json.NewEncoder(w)
	err := encoder.Encode(tokenResponse)
	if err != nil {
		http.Error(w, fmt.Sprintf("couldn't encode user: %s", err.Error()), http.StatusInternalServerError)
		return
	}
}

func userFromToken(token *Token, groups []string) *User {
	return &User{
		Username: token.Email,
		Uid:      token.UserID,
		Groups:   groups,
		Extra:    map[string][]string{},
	}
}

type AuthFailedStatus struct {
	Authenticated bool `json:"authenticated"`
}

type AuthFailed struct {
	APIVersion string            `json:"apiVersion"`
	Kind       string            `json:"kind"`
	Status     *AuthFailedStatus `json:"status"`
}

func authFailed(w http.ResponseWriter) error {
	failed := &AuthFailed{
		APIVersion: "authentication.k8s.io/v1beta1",
		Kind:       "TokenReview",
		Status: &AuthFailedStatus{
			Authenticated: false,
		},
	}

	enc := json.NewEncoder(w)
	err := enc.Encode(failed)
	if err != nil {
		return err
	}
	return nil
}

type KubernetesValidate struct {
	groups map[string][]string
}

func NewKubernetesValidate() (*KubernetesValidate, error) {
	bytes, err := ioutil.ReadFile("groups.yaml")
	if err != nil {
		return nil, err
	}
	var groups map[string][]string
	err = yaml.Unmarshal(bytes, &groups)
	if err != nil {
		return nil, err
	}

	return &KubernetesValidate{groups}, nil
}

func (k *KubernetesValidate) userGroups(email string) []string {
	groups := k.groups[email]
	if len(groups) == 0 {
		return []string{"user"}
	}
	return groups
}

func (k *KubernetesValidate) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	var t TokenReview
	err := decoder.Decode(&t)
	if err != nil {
		http.Error(w, fmt.Sprintf("couldn't decode request: %s", err.Error()), http.StatusBadRequest)
		return
	}

	ctx := appengine.NewContext(r)
	token, err := getToken(ctx, t.Spec.Token)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		authFailed(w)
		return
	}

	if token == nil {
		w.WriteHeader(http.StatusUnauthorized)
		authFailed(w)
		return
	}

	if token.IsExpired() {
		w.WriteHeader(http.StatusUnauthorized)
		authFailed(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	validUser(w, userFromToken(token, k.userGroups(token.Email)))
}
