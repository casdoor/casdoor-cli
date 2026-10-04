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
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/casdoor/casdoor-cli/internal/config"
)

// fakeTokenServer answers the token API of Casdoor and records the requests.
func fakeTokenServer(t *testing.T, answer func(form url.Values) any) (*httptest.Server, *[]url.Values) {
	var requests []url.Values
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/login/oauth/access_token" {
			t.Errorf("unexpected request %s", r.URL)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}
		requests = append(requests, r.PostForm)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(answer(r.PostForm))
	}))
	t.Cleanup(server.Close)
	return server, &requests
}

func freePort(t *testing.T) int {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = l.Close() }()
	return l.Addr().(*net.TCPAddr).Port
}

func TestBrowserLogin(t *testing.T) {
	server, requests := fakeTokenServer(t, func(form url.Values) any {
		return map[string]any{"access_token": "user-token", "refresh_token": "refresh", "token_type": "Bearer", "expires_in": 3600}
	})
	p := &config.Profile{Endpoint: server.URL, ClientId: "client", RedirectPort: freePort(t)}

	// The "browser" logs in and is sent back to the callback with the code and the state.
	open := func(authURL string) error {
		u, err := url.Parse(authURL)
		if err != nil {
			return err
		}
		q := u.Query()
		if u.Path != "/login/oauth/authorize" || q.Get("code_challenge_method") != "S256" || q.Get("code_challenge") == "" || q.Has("client_secret") {
			t.Errorf("bad authorize URL %s", authURL)
		}
		go func() {
			// A callback with another state is refused, the login goes on.
			resp, err := http.Get(q.Get("redirect_uri") + "?code=stolen&state=other")
			if err == nil {
				_ = resp.Body.Close()
				if resp.StatusCode != http.StatusBadRequest {
					t.Errorf("a callback with another state got %d", resp.StatusCode)
				}
			}
		}()
		return nil
	}

	_, err := BrowserLogin(context.Background(), p, open, io.Discard)
	if err == nil || !strings.Contains(err.Error(), "state") {
		t.Fatalf("a callback with another state should fail the login, got %v", err)
	}
	if len(*requests) != 0 {
		t.Fatal("no code may be exchanged after a wrong state")
	}

	open = func(authURL string) error {
		u, _ := url.Parse(authURL)
		q := u.Query()
		go func() {
			resp, err := http.Get(q.Get("redirect_uri") + "?code=the-code&state=" + url.QueryEscape(q.Get("state")))
			if err == nil {
				_ = resp.Body.Close()
			}
		}()
		return nil
	}
	token, err := BrowserLogin(context.Background(), p, open, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	if token.AccessToken != "user-token" || token.RefreshToken != "refresh" {
		t.Fatalf("got %+v", token)
	}
	form := (*requests)[0]
	if form.Get("grant_type") != "authorization_code" || form.Get("code") != "the-code" || form.Get("code_verifier") == "" {
		t.Fatalf("bad token request %v", form)
	}
}

func TestPasswordLoginError(t *testing.T) {
	// Old Casdoor versions put the error in the access token.
	server, _ := fakeTokenServer(t, func(form url.Values) any {
		return map[string]any{"access_token": "error: invalid username or password", "token_type": "Bearer"}
	})
	p := &config.Profile{Endpoint: server.URL, ClientId: "client", ClientSecret: "secret"}
	if _, err := PasswordLogin(context.Background(), p, "alice", "wrong"); err == nil || err.Error() != "invalid username or password" {
		t.Fatalf("got %v", err)
	}
}

func TestRefresh(t *testing.T) {
	server, requests := fakeTokenServer(t, func(form url.Values) any {
		return map[string]any{"access_token": "new-token", "token_type": "Bearer", "expires_in": 3600}
	})

	valid := &config.Profile{Endpoint: server.URL, AccessToken: "token", RefreshToken: "refresh", Expiry: time.Now().Add(time.Hour)}
	if changed, err := Refresh(context.Background(), valid); changed || err != nil {
		t.Fatalf("a valid token is kept: %v %v", changed, err)
	}

	expired := &config.Profile{Endpoint: server.URL, ClientId: "client", AccessToken: "token", RefreshToken: "refresh", Expiry: time.Now().Add(-time.Hour)}
	changed, err := Refresh(context.Background(), expired)
	if err != nil || !changed {
		t.Fatalf("got %v %v", changed, err)
	}
	if expired.AccessToken != "new-token" || expired.RefreshToken != "refresh" || time.Until(expired.Expiry) < 50*time.Minute {
		t.Fatalf("got %+v", expired)
	}
	if form := (*requests)[0]; form.Get("grant_type") != "refresh_token" || form.Get("refresh_token") != "refresh" {
		t.Fatalf("bad refresh request %v", form)
	}

	noRefresh := &config.Profile{Endpoint: server.URL, AccessToken: "token", Expiry: time.Now().Add(-time.Hour)}
	if _, err = Refresh(context.Background(), noRefresh); err == nil {
		t.Fatal("an expired token without a refresh token needs a new login")
	}
}
