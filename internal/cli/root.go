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

package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/casdoor/casdoor-cli/internal/auth"
	"github.com/casdoor/casdoor-cli/internal/config"
	"github.com/casdoor/casdoor-cli/internal/output"
	"github.com/casdoor/casdoor-go-sdk/casdoorsdk"
	"github.com/spf13/cobra"
)

// Version is set at build time with -ldflags "-X github.com/casdoor/casdoor-cli/internal/cli.Version=..."
var Version = "dev"

type app struct {
	profileName string
	output      string

	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer

	cfg *config.Config
}

func init() {
	casdoorsdk.SetHttpClient(&http.Client{Timeout: 60 * time.Second})
}

// Execute runs the CLI and returns the exit code.
func Execute() int {
	root := newRootCommand(&app{stdin: os.Stdin, stdout: os.Stdout, stderr: os.Stderr})
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return 1
	}
	return 0
}

func newRootCommand(a *app) *cobra.Command {
	root := &cobra.Command{
		Use:           "casdoor",
		Short:         "Command line interface for Casdoor",
		Long:          "Manage Casdoor from the terminal: log in, then list, get, create, update and delete users, applications, roles, permissions and every other Casdoor object.",
		Version:       Version,
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			a.cfg = cfg
			return nil
		},
	}
	root.SetIn(a.stdin)
	root.SetOut(a.stdout)
	root.SetErr(a.stderr)
	root.PersistentFlags().StringVar(&a.profileName, "profile", "", "the profile to use, $CASDOOR_PROFILE or the current profile by default")
	root.PersistentFlags().StringVarP(&a.output, "output", "o", "table", "output format: "+strings.Join(output.Formats, ", "))

	root.AddCommand(newLoginCommand(a), newLogoutCommand(a), newWhoamiCommand(a), newProfileCommand(a),
		newApiCommand(a), newEnforceCommand(a))
	for _, r := range resources {
		root.AddCommand(newResourceCommand(a, r))
	}
	return root
}

// profile returns the profile to use, with a fresh access token.
func (a *app) profile(ctx context.Context) (*config.Profile, error) {
	p := a.cfg.Resolve(a.profileName)
	if err := p.Validate(); err != nil {
		return nil, err
	}

	refreshed, err := auth.Refresh(ctx, p)
	if err != nil {
		return nil, err
	}
	if saved, ok := a.cfg.Profiles[a.cfg.CurrentName(a.profileName)]; ok && refreshed && os.Getenv("CASDOOR_ACCESS_TOKEN") == "" {
		saved.AccessToken, saved.RefreshToken, saved.Expiry = p.AccessToken, p.RefreshToken, p.Expiry
		if err = a.cfg.Save(); err != nil {
			return nil, err
		}
	}
	return p, nil
}

// client calls the API as the logged-in user, or as the application when there is no login.
func (a *app) client(ctx context.Context) (*casdoorsdk.Client, *config.Profile, error) {
	p, err := a.profile(ctx)
	if err != nil {
		return nil, nil, err
	}
	c := casdoorsdk.NewClient(p.Endpoint, p.ClientId, p.ClientSecret, "", p.Organization, p.Application)
	if p.AccessToken != "" {
		c = c.WithAccessToken(p.AccessToken)
	}
	return c, p, nil
}

func apiGet(c *casdoorsdk.Client, action string, query map[string]string) (*casdoorsdk.Response, error) {
	return c.DoGetResponse(c.GetUrl(action, query))
}

func apiPost(c *casdoorsdk.Client, action string, query map[string]string, body any) (*casdoorsdk.Response, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	return c.DoPost(action, query, data, false, false)
}

func (a *app) print(data any, columns []string) error {
	return output.Print(a.stdout, a.output, data, columns)
}
