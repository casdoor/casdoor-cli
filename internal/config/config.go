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
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const DefaultProfile = "default"

// Profile is a Casdoor server and the credentials the CLI uses for it.
type Profile struct {
	Endpoint     string `json:"endpoint"`
	ClientId     string `json:"clientId"`
	ClientSecret string `json:"clientSecret,omitempty"`
	Organization string `json:"organization"`
	Application  string `json:"application"`
	RedirectPort int    `json:"redirectPort,omitempty"`

	// The token of the logged-in user, empty when the CLI acts as the application itself
	// with the client ID and client secret.
	AccessToken  string    `json:"accessToken,omitempty"`
	RefreshToken string    `json:"refreshToken,omitempty"`
	Expiry       time.Time `json:"expiry,omitzero"`
}

type Config struct {
	Current  string              `json:"current"`
	Profiles map[string]*Profile `json:"profiles"`

	path string
}

// Path returns the config file, $CASDOOR_CONFIG or casdoor/config.json in the user's config
// directory (e.g. ~/.config on Linux).
func Path() (string, error) {
	if path := os.Getenv("CASDOOR_CONFIG"); path != "" {
		return path, nil
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "casdoor", "config.json"), nil
}

func Load() (*Config, error) {
	path, err := Path()
	if err != nil {
		return nil, err
	}
	return LoadFile(path)
}

// LoadFile reads the config, a missing file is an empty config.
func LoadFile(path string) (*Config, error) {
	c := &Config{Profiles: map[string]*Profile{}, path: path}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return c, nil
	}
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal(data, c); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", path, err)
	}
	if c.Profiles == nil {
		c.Profiles = map[string]*Profile{}
	}
	return c, nil
}

// Save writes the config, only the user can read it since it holds secrets and tokens.
func (c *Config) Save() error {
	if err := os.MkdirAll(filepath.Dir(c.path), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	tmp := c.path + ".tmp"
	if err = os.WriteFile(tmp, append(data, '\n'), 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, c.path)
}

func (c *Config) File() string {
	return c.path
}

// CurrentName is the profile to use: the given one, $CASDOOR_PROFILE, the current one or
// "default".
func (c *Config) CurrentName(name string) string {
	if name != "" {
		return name
	}
	if name = os.Getenv("CASDOOR_PROFILE"); name != "" {
		return name
	}
	if c.Current != "" {
		return c.Current
	}
	return DefaultProfile
}

func (c *Config) Names() []string {
	names := make([]string, 0, len(c.Profiles))
	for name := range c.Profiles {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Resolve returns the profile to use with the environment variables applied, so the CLI also
// works without a config file, e.g. in CI.
func (c *Config) Resolve(name string) *Profile {
	p := Profile{}
	if saved, ok := c.Profiles[c.CurrentName(name)]; ok {
		p = *saved
	}

	override := func(field *string, env string) {
		if v := os.Getenv(env); v != "" {
			*field = v
		}
	}
	override(&p.Endpoint, "CASDOOR_ENDPOINT")
	override(&p.ClientId, "CASDOOR_CLIENT_ID")
	override(&p.ClientSecret, "CASDOOR_CLIENT_SECRET")
	override(&p.Organization, "CASDOOR_ORGANIZATION")
	override(&p.Application, "CASDOOR_APPLICATION")
	if v := os.Getenv("CASDOOR_ACCESS_TOKEN"); v != "" {
		p.AccessToken = v
		p.RefreshToken = ""
		p.Expiry = time.Time{}
	}
	p.Endpoint = strings.TrimRight(p.Endpoint, "/")
	return &p
}

func (p *Profile) Validate() error {
	var missing []string
	if p.Endpoint == "" {
		missing = append(missing, "endpoint (CASDOOR_ENDPOINT)")
	}
	if p.Organization == "" {
		missing = append(missing, "organization (CASDOOR_ORGANIZATION)")
	}
	if p.AccessToken == "" && (p.ClientId == "" || p.ClientSecret == "") {
		missing = append(missing, "a login (casdoor login) or client ID and client secret (CASDOOR_CLIENT_ID, CASDOOR_CLIENT_SECRET)")
	}
	if len(missing) > 0 {
		return fmt.Errorf("not configured, missing %s", strings.Join(missing, ", "))
	}
	return nil
}
