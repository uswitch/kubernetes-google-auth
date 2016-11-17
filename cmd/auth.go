package main

import (
	"fmt"
	"github.com/skratchdot/open-golang/open"
	"net/http"
	"net/url"
)

const StartPath = "/auth/start"

func startAuthenticationURL(baseURL *url.URL) *url.URL {
	baseURL.Path = StartPath
	return baseURL
}

type AuthenticatedInfo struct {
	Email string
	Token string
}

func startServer(port string) <-chan *AuthenticatedInfo {
	resultCh := make(chan *AuthenticatedInfo)

	http.HandleFunc("/authed", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Finished authentication. You can close this browser.")
		auth := &AuthenticatedInfo{
			Token: r.FormValue("token"),
			Email: r.FormValue("email"),
		}
		resultCh <- auth
	})

	binding := fmt.Sprintf("127.0.0.1:%s", port)
	go http.ListenAndServe(binding, nil)

	return resultCh
}

func authedURL(port string) string {
	return fmt.Sprintf("http://localhost:%s/authed", port)
}

func ExecAuth(clusterName string, baseURL *url.URL, localPort string) error {
	resultCh := startServer(localPort)

	startURL := startAuthenticationURL(baseURL)
	q := startURL.Query()
	q.Set("redirect", authedURL(localPort))
	startURL.RawQuery = q.Encode()

	err := open.Run(startURL.String())
	if err != nil {
		return err
	}

	auth := <-resultCh

	config, err := ReadKubeConfig(defaultConfigPath())
	config.UpdateUser(auth.Email, auth.Token)
	context := config.UpdateContextUser(auth.Email, clusterName)
	config.CurrentContext = context.Name

	err = config.Save()
	if err != nil {
		return err
	}

	fmt.Println("Saved new credentials.")
	return nil
}
