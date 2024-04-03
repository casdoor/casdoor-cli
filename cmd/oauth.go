package cmd

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/pkg/browser"
	log "github.com/sirupsen/logrus"
	"gitlab.com/sdv9972401/casdoor-cli/models"
	"golang.org/x/oauth2"
	"io"
	"net/http"
	"net/url"
)

// ParseOAuthResponse parses the OAuth response data into a TokenData struct.
//
// Parameter:
//
//	data []byte - the OAuth response data to be parsed
//
// Return:
//
//	*models.TokenData - the parsed TokenData struct
//	error - an error, if any, during the parsing process
func ParseOAuthResponse(data []byte) (*models.TokenData, error) {
	var tokenData models.TokenData
	err := json.Unmarshal(data, &tokenData)
	if err != nil {
		log.Errorf("failed to parse OAuth response: %s", err)
		return nil, err
	}
	return &tokenData, nil
}

// OAuthHandler handles the OAuth process for the given Casdoor configuration.
//
// It takes a CasdoorConfig pointer as a parameter and returns a byte slice and an error.
func OAuthHandler(casdoorConfig *models.CasdoorConfig) ([]byte, error) {
	var dataChan = make(chan []byte)

	ctx := context.Background()

	state, err := randString(16)
	if err != nil {
		log.Fatal(err)
	}

	authURL := fmt.Sprintf("%s/login/oauth/authorize?client_id=%s&redirect_uri=%s&response_type=code&scope=openid&state=%s",
		casdoorConfig.Endpoint, casdoorConfig.ClientID, url.QueryEscape(casdoorConfig.RedirectURI), state)
	log.Debugf("opening authorization URL: %s", authURL)

	log.Debug("awaiting user connection...")
	err = browser.OpenURL(authURL)
	if err != nil {
		log.Fatalf("failed to open browser %s", err)
	}

	provider, err := oidc.NewProvider(ctx, casdoorConfig.Endpoint)
	if err != nil {
		log.Fatal(err)
	}

	oidcConfig := oidc.Config{
		ClientID: casdoorConfig.ClientID,
	}

	verifier := provider.Verifier(&oidcConfig)

	config := oauth2.Config{
		ClientID:     casdoorConfig.ClientID,
		ClientSecret: casdoorConfig.ClientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  casdoorConfig.RedirectURI,
		Scopes:       []string{oidc.ScopeOpenID},
	}

	u, err := url.Parse(config.RedirectURL)
	if err != nil {
		log.Fatal(err)
	}
	port := u.Port()

	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("state") != state {
			http.Error(w, "state did not match", http.StatusBadRequest)
			return
		}

		log.Debug("callback received. Attempting to exchange code...")
		oauth2Token, err := config.Exchange(ctx, r.URL.Query().Get("code"))
		if err != nil {
			http.Error(w, "failed to exchange token: "+err.Error(), http.StatusBadRequest)
			return
		} else {
			log.Debug("token successfully retrieved.")
		}

		rawIDToken, ok := oauth2Token.Extra("id_token").(string)
		if !ok {
			http.Error(w, "fo id_token field in oauth2 token.", http.StatusInternalServerError)
			return
		}

		idToken, err := verifier.Verify(ctx, rawIDToken)
		if err != nil {
			http.Error(w, "failed to verify ID Token: "+err.Error(), http.StatusInternalServerError)
			return
		} else {
			log.Debug("ID Token successfully verified.")
		}

		resp := struct {
			OAuth2Token   *oauth2.Token
			IDTokenClaims *json.RawMessage
		}{oauth2Token, new(json.RawMessage)}

		if err = idToken.Claims(&resp.IDTokenClaims); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		data, err := json.MarshalIndent(resp, "", "    ")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write([]byte("✅ You have been successfully authenticated. You may now close this window."))

		dataChan <- data
	})

	go func() {
		err := http.ListenAndServe(":"+port, nil)
		if err != nil {
			log.Fatalf("Failed to start local server: %v", err)
		}
	}()

	data := <-dataChan

	return data, err
}

// randString generates a random string of specified length.
//
// nByte: the number of bytes for the random string.
// string: the generated random string.
// error: an error, if any.
func randString(nByte int) (string, error) {
	b := make([]byte, nByte)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
