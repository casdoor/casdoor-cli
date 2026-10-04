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

package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func clearEnv(t *testing.T) {
	for _, env := range []string{"CASDOOR_PROFILE", "CASDOOR_ENDPOINT", "CASDOOR_CLIENT_ID", "CASDOOR_CLIENT_SECRET",
		"CASDOOR_ORGANIZATION", "CASDOOR_APPLICATION", "CASDOOR_ACCESS_TOKEN"} {
		t.Setenv(env, "")
	}
}

func TestSaveAndLoad(t *testing.T) {
	clearEnv(t)
	path := filepath.Join(t.TempDir(), "casdoor", "config.json")
	c, err := LoadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(c.Profiles) != 0 {
		t.Fatal("a missing file is an empty config")
	}

	expiry := time.Date(2030, 1, 2, 3, 4, 5, 0, time.UTC)
	c.Profiles["prod"] = &Profile{Endpoint: "https://door.example.com", ClientId: "id", Organization: "org", AccessToken: "token", Expiry: expiry}
	c.Current = "prod"
	if err = c.Save(); err != nil {
		t.Fatal(err)
	}

	if runtime.GOOS != "windows" {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode().Perm() != 0o600 {
			t.Errorf("the config holds tokens, its mode is %v", info.Mode().Perm())
		}
	}

	loaded, err := LoadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	p := loaded.Resolve("")
	if p.Endpoint != "https://door.example.com" || p.AccessToken != "token" || !p.Expiry.Equal(expiry) {
		t.Fatalf("got %+v", p)
	}
	if err = p.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestResolve(t *testing.T) {
	clearEnv(t)
	c := &Config{Current: "a", Profiles: map[string]*Profile{
		"a": {Endpoint: "https://a.example.com/", Organization: "org-a", AccessToken: "token-a", RefreshToken: "refresh-a"},
		"b": {Endpoint: "https://b.example.com", Organization: "org-b"},
	}}

	if p := c.Resolve(""); p.Organization != "org-a" || p.Endpoint != "https://a.example.com" {
		t.Errorf("current profile: %+v", p)
	}
	if p := c.Resolve("b"); p.Organization != "org-b" {
		t.Errorf("--profile: %+v", p)
	}
	t.Setenv("CASDOOR_PROFILE", "b")
	if p := c.Resolve(""); p.Organization != "org-b" {
		t.Errorf("$CASDOOR_PROFILE: %+v", p)
	}

	t.Setenv("CASDOOR_PROFILE", "")
	t.Setenv("CASDOOR_ORGANIZATION", "org-env")
	t.Setenv("CASDOOR_ACCESS_TOKEN", "token-env")
	p := c.Resolve("")
	if p.Organization != "org-env" || p.AccessToken != "token-env" || p.RefreshToken != "" {
		t.Errorf("environment: %+v", p)
	}
	if c.Profiles["a"].AccessToken != "token-a" {
		t.Error("Resolve must not change the saved profile")
	}
}

func TestValidate(t *testing.T) {
	cases := []struct {
		profile Profile
		ok      bool
	}{
		{Profile{}, false},
		{Profile{Endpoint: "e", Organization: "o"}, false},
		{Profile{Endpoint: "e", Organization: "o", ClientId: "id"}, false},
		{Profile{Endpoint: "e", Organization: "o", ClientId: "id", ClientSecret: "secret"}, true},
		{Profile{Endpoint: "e", Organization: "o", AccessToken: "token"}, true},
	}
	for _, c := range cases {
		if err := c.profile.Validate(); (err == nil) != c.ok {
			t.Errorf("%+v: %v", c.profile, err)
		}
	}
}
