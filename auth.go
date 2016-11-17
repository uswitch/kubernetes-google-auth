package kauth

import (
	"encoding/json"
	"fmt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/appengine"
	"net/http"
	"net/url"
	"os"
	"time"
)

func scheme(url *url.URL) string {
	if url.Scheme != "" {
		return url.Scheme
	}

	return "http"
}

func redirectURL(r *http.Request) string {
	return fmt.Sprintf("%s://%s/auth/complete", scheme(r.URL), r.Host)
}

const (
	envVarOAuthClient string = "OAUTH2_CLIENT_ID"
	envVarOAuthSecret string = "OAUTH2_CLIENT_SECRET"
	envVarValidDomain string = "VALID_DOMAIN"
)

func oauthConfig(r *http.Request) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     os.Getenv(envVarOAuthClient),
		ClientSecret: os.Getenv(envVarOAuthSecret),
		RedirectURL:  redirectURL(r),
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}
}

const RedirectCookieName = "redirect"

func startAuthentication(w http.ResponseWriter, r *http.Request) {
	csrfToken := "1"
	googleAuthenticateURL := oauthConfig(r).AuthCodeURL(csrfToken)
	if r.FormValue("redirect") != "" {
		http.SetCookie(w, &http.Cookie{Name: RedirectCookieName, Value: r.FormValue("redirect")})
	}
	http.Redirect(w, r, googleAuthenticateURL, http.StatusTemporaryRedirect)
}

type UserInfo struct {
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	EmailVerified bool   `json:"email_verified"`
	Domain        string `json:"hd"`
	ID            string `json:"sub"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	Email         string `json:"email"`
	Locale        string `json:"locale"`
}

func FetchUserInfo(client *http.Client) (*UserInfo, error) {
	response, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(response.Body)
	defer response.Body.Close()

	var u UserInfo
	err = decoder.Decode(&u)
	return &u, err
}

var ErrInvalidDomain = fmt.Errorf("Invalid email domain")

func (u *UserInfo) IsValidDomain() bool {
	return u.Domain == os.Getenv(envVarValidDomain)
}

func redirectFromCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie(RedirectCookieName)
	if err != nil {
		if err == http.ErrNoCookie {
			return "", nil
		} else {
			return "", err
		}
	}
	return cookie.Value, nil
}

func completeAuthentication(w http.ResponseWriter, r *http.Request) {
	oauthCode := r.FormValue("code")
	ctx := appengine.NewContext(r)
	config := oauthConfig(r)
	tok, err := config.Exchange(ctx, oauthCode)
	if err != nil {
		http.Error(w, fmt.Sprintf("couldn't retrieve oauth token: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	client := config.Client(ctx, tok)
	u, err := FetchUserInfo(client)
	if err != nil {
		http.Error(w, fmt.Sprintf("couldn't retrieve user data: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	// Generate a UUID
	t, err := newToken(u)
	if err != nil {
		http.Error(w, fmt.Sprintf("error creating token: %s", err.Error()), http.StatusUnauthorized)
		return
	}
	err = storeToken(ctx, t)
	if err != nil {
		http.Error(w, fmt.Sprintf("error storing token: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	redirect, err := redirectFromCookie(r)
	if err != nil {
		http.Error(w, fmt.Sprintf("error storing token: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	if redirect == "" {
		enc := json.NewEncoder(w)
		enc.Encode(t)
	} else {
		redirectURL, err := url.Parse(redirect)
		if err != nil {
			http.Error(w, fmt.Sprintf("error generating redirect: %s", err.Error()), http.StatusInternalServerError)
			return
		}
		q := redirectURL.Query()
		q.Set("token", t.ID)
		q.Set("email", t.Email)
		redirectURL.RawQuery = q.Encode()
		http.SetCookie(w, &http.Cookie{Name: RedirectCookieName, Expires: time.Now()})

		http.Redirect(w, r, redirectURL.String(), http.StatusTemporaryRedirect)
	}
}
