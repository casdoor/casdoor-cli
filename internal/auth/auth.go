// Copyright 2026 The Casdoor Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package auth

import (
	"context"
	"errors"
	"fmt"
	"html"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/casdoor/casdoor-cli/internal/config"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

const DefaultRedirectPort = 9000

// RedirectURL is the callback of the browser login, it has to be in the "Redirect URLs" of
// the Casdoor application.
func RedirectURL(port int) string {
	if port == 0 {
		port = DefaultRedirectPort
	}
	return fmt.Sprintf("http://localhost:%d/callback", port)
}

func oauthConfig(p *config.Profile) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     p.ClientId,
		ClientSecret: p.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:   p.Endpoint + "/login/oauth/authorize",
			TokenURL:  p.Endpoint + "/api/login/oauth/access_token",
			AuthStyle: oauth2.AuthStyleInParams,
		},
		RedirectURL: RedirectURL(p.RedirectPort),
		Scopes:      []string{"openid", "profile", "email", "offline_access"},
	}
}

// checkToken turns the "error: xxx" access token that old Casdoor versions return into an
// error.
func checkToken(token *oauth2.Token, err error) (*oauth2.Token, error) {
	if err != nil {
		return nil, err
	}
	if strings.HasPrefix(token.AccessToken, "error:") {
		return nil, errors.New(strings.TrimSpace(strings.TrimPrefix(token.AccessToken, "error:")))
	}
	if token.AccessToken == "" {
		return nil, errors.New("casdoor returned no access token")
	}
	return token, nil
}

// SaveToken keeps the token of a login in the profile.
func SaveToken(p *config.Profile, token *oauth2.Token) {
	p.AccessToken = token.AccessToken
	p.RefreshToken = token.RefreshToken
	p.Expiry = token.Expiry
}

// BrowserLogin logs in with the authorization code flow and PKCE: Casdoor sends the browser
// back to a server on localhost that receives the code.
func BrowserLogin(ctx context.Context, p *config.Profile, openBrowser func(string) error, out io.Writer) (*oauth2.Token, error) {
	cfg := oauthConfig(p)
	port := p.RedirectPort
	if port == 0 {
		port = DefaultRedirectPort
	}
	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return nil, fmt.Errorf("failed to listen on port %d for the login callback (use --port to pick another one): %w", port, err)
	}

	state := oauth2.GenerateVerifier()
	verifier := oauth2.GenerateVerifier()
	authURL := cfg.AuthCodeURL(state, oauth2.S256ChallengeOption(verifier))

	type result struct {
		code string
		err  error
	}
	results := make(chan result, 1)
	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		res := result{code: q.Get("code")}
		switch {
		case q.Get("state") != state:
			res.err = errors.New("the state of the callback does not match, the login was not started by this CLI")
		case q.Get("error") != "":
			res.err = fmt.Errorf("%s: %s", q.Get("error"), q.Get("error_description"))
		case res.code == "":
			res.err = errors.New("the callback has no code")
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if res.err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "<h3>Login failed</h3><p>%s</p>", html.EscapeString(res.err.Error()))
		} else {
			fmt.Fprint(w, "<h3>Logged in to Casdoor CLI</h3><p>You can close this window and go back to the terminal.</p>")
		}
		select {
		case results <- res:
		default:
		}
	})
	server := &http.Server{Handler: mux, ReadHeaderTimeout: 10 * time.Second}
	go func() { _ = server.Serve(listener) }()
	defer func() { _ = server.Close() }()

	fmt.Fprintf(out, "Opening the browser to log in, or open this URL:\n\n  %s\n\n", authURL)
	if openBrowser != nil {
		if err = openBrowser(authURL); err != nil {
			fmt.Fprintf(out, "Failed to open the browser: %s\n", err)
		}
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	var res result
	select {
	case res = <-results:
	case <-ctx.Done():
		return nil, errors.New("timed out waiting for the login in the browser")
	}
	if res.err != nil {
		return nil, res.err
	}

	return checkToken(cfg.Exchange(ctx, res.code, oauth2.VerifierOption(verifier)))
}

// PasswordLogin uses the password grant, it has to be enabled in the "Grant types" of the
// application.
func PasswordLogin(ctx context.Context, p *config.Profile, username, password string) (*oauth2.Token, error) {
	return checkToken(oauthConfig(p).PasswordCredentialsToken(ctx, username, password))
}

// ClientCredentials checks the client ID and client secret by getting a token with them.
func ClientCredentials(ctx context.Context, p *config.Profile) error {
	cfg := oauthConfig(p)
	cc := clientcredentials.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		TokenURL:     cfg.Endpoint.TokenURL,
		AuthStyle:    oauth2.AuthStyleInParams,
	}
	_, err := checkToken(cc.Token(ctx))
	return err
}

// Refresh returns a valid access token of the profile, refreshing it when it expired. The
// second result tells whether the profile changed and should be saved.
func Refresh(ctx context.Context, p *config.Profile) (bool, error) {
	if p.AccessToken == "" || p.Expiry.IsZero() || time.Until(p.Expiry) > time.Minute {
		return false, nil
	}
	if p.RefreshToken == "" {
		return false, errors.New("the login has expired, run: casdoor login")
	}

	token, err := checkToken(oauthConfig(p).TokenSource(ctx, &oauth2.Token{RefreshToken: p.RefreshToken}).Token())
	if err != nil {
		return false, fmt.Errorf("failed to refresh the login, run: casdoor login: %w", err)
	}
	if token.RefreshToken == "" {
		token.RefreshToken = p.RefreshToken
	}
	SaveToken(p, token)
	return true, nil
}
