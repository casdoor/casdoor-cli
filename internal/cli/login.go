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
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/casdoor/casdoor-cli/internal/auth"
	"github.com/casdoor/casdoor-cli/internal/config"
	"github.com/casdoor/casdoor-cli/internal/output"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func newLoginCommand(a *app) *cobra.Command {
	var p config.Profile
	var username string
	var passwordStdin, clientCredentials, noBrowser bool
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Log in to a Casdoor server",
		Long: `Log in to a Casdoor server and save it as a profile.

By default the browser opens Casdoor's login page (authorization code flow with PKCE).
Add the redirect URL, http://localhost:9000/callback by default, to the "Redirect URLs"
of the application in Casdoor.

--username logs in with a password instead (the "Password" grant type has to be enabled
in the application). --client-credentials does not log in a user at all, the CLI then
acts as the application with its client ID and client secret.

The flags default to the values saved in the profile, so "casdoor login" alone logs in
again to the same server.`,
		Example: `  casdoor login --endpoint https://door.casdoor.com --client-id <id> --organization casbin
  casdoor login --username alice
  casdoor login --profile ci --endpoint https://door.casdoor.com --client-id <id> --client-secret <secret> --organization casbin --client-credentials`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			name := a.cfg.CurrentName(a.profileName)
			profile := &config.Profile{}
			if saved, ok := a.cfg.Profiles[name]; ok {
				*profile = *saved
			}
			flags := cmd.Flags()
			if flags.Changed("endpoint") {
				profile.Endpoint = strings.TrimRight(p.Endpoint, "/")
			}
			if flags.Changed("client-id") {
				profile.ClientId = p.ClientId
			}
			if flags.Changed("client-secret") {
				profile.ClientSecret = p.ClientSecret
			}
			if flags.Changed("organization") {
				profile.Organization = p.Organization
			}
			if flags.Changed("application") {
				profile.Application = p.Application
			}
			if flags.Changed("port") {
				profile.RedirectPort = p.RedirectPort
			}
			if profile.Endpoint == "" || profile.ClientId == "" || profile.Organization == "" {
				return fmt.Errorf("--endpoint, --client-id and --organization are required for a new profile")
			}

			ctx := cmd.Context()
			switch {
			case clientCredentials:
				if profile.ClientSecret == "" {
					return fmt.Errorf("--client-credentials needs --client-secret")
				}
				if err := auth.ClientCredentials(ctx, profile); err != nil {
					return fmt.Errorf("the client ID or client secret is wrong: %w", err)
				}
				profile.AccessToken, profile.RefreshToken = "", ""
				profile.Expiry = time.Time{}
			case username != "":
				password, err := a.readPassword(passwordStdin)
				if err != nil {
					return err
				}
				token, err := auth.PasswordLogin(ctx, profile, username, password)
				if err != nil {
					return err
				}
				auth.SaveToken(profile, token)
			default:
				open := openBrowser
				if noBrowser {
					open = nil
				}
				token, err := auth.BrowserLogin(ctx, profile, open, a.stderr)
				if err != nil {
					return err
				}
				auth.SaveToken(profile, token)
			}

			a.cfg.Profiles[name] = profile
			a.cfg.Current = name
			if err := a.cfg.Save(); err != nil {
				return err
			}

			if clientCredentials {
				fmt.Fprintf(a.stdout, "Saved profile %q, the CLI acts as the application %s\n", name, profile.ClientId)
				return nil
			}
			c, _, err := a.client(ctx)
			if err != nil {
				return err
			}
			resp, err := apiGet(c, "get-account", nil)
			if err != nil {
				return err
			}
			user, _ := resp.Data.(map[string]any)
			fmt.Fprintf(a.stdout, "Logged in as %s/%s, saved as profile %q\n", output.Cell(user["owner"]), output.Cell(user["name"]), name)
			return nil
		},
	}
	cmd.Flags().StringVar(&p.Endpoint, "endpoint", "", "the URL of the Casdoor server, e.g. https://door.casdoor.com")
	cmd.Flags().StringVar(&p.ClientId, "client-id", "", "the client ID of the Casdoor application")
	cmd.Flags().StringVar(&p.ClientSecret, "client-secret", "", "the client secret of the application, not needed for the browser login")
	cmd.Flags().StringVar(&p.Organization, "organization", "", "the organization, the default owner of the objects")
	cmd.Flags().StringVar(&p.Application, "application", "", "the name of the application")
	cmd.Flags().IntVar(&p.RedirectPort, "port", auth.DefaultRedirectPort, "the port of the local server that receives the browser login")
	cmd.Flags().StringVar(&username, "username", "", "log in with this user's password instead of the browser")
	cmd.Flags().BoolVar(&passwordStdin, "password-stdin", false, "read the password from stdin")
	cmd.Flags().BoolVar(&clientCredentials, "client-credentials", false, "act as the application with the client ID and client secret, no user login")
	cmd.Flags().BoolVar(&noBrowser, "no-browser", false, "only print the login URL")
	return cmd
}

func (a *app) readPassword(fromStdin bool) (string, error) {
	if fromStdin {
		line, err := bufio.NewReader(a.stdin).ReadString('\n')
		if err != nil && err != io.EOF {
			return "", err
		}
		return strings.TrimRight(line, "\r\n"), nil
	}
	f, ok := a.stdin.(*os.File)
	if !ok || !term.IsTerminal(int(f.Fd())) {
		return "", fmt.Errorf("no terminal to ask for the password, use --password-stdin")
	}
	fmt.Fprint(a.stderr, "Password: ")
	password, err := term.ReadPassword(int(f.Fd()))
	fmt.Fprintln(a.stderr)
	return string(password), err
}

func openBrowser(url string) error {
	switch runtime.GOOS {
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		return exec.Command("open", url).Start()
	default:
		return exec.Command("xdg-open", url).Start()
	}
}

func newLogoutCommand(a *app) *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Forget the login of the profile, the server settings are kept",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			name := a.cfg.CurrentName(a.profileName)
			p, ok := a.cfg.Profiles[name]
			if !ok {
				return fmt.Errorf("no profile %q", name)
			}
			p.AccessToken, p.RefreshToken = "", ""
			p.Expiry = time.Time{}
			if err := a.cfg.Save(); err != nil {
				return err
			}
			fmt.Fprintf(a.stdout, "Logged out of profile %q\n", name)
			return nil
		},
	}
}

func newWhoamiCommand(a *app) *cobra.Command {
	return &cobra.Command{
		Use:   "whoami",
		Short: "Show the logged-in user",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, _, err := a.client(cmd.Context())
			if err != nil {
				return err
			}
			resp, err := apiGet(c, "get-account", nil)
			if err != nil {
				return err
			}
			// Casdoor returns the token of the session too, it does not belong in the output.
			if user, ok := resp.Data.(map[string]any); ok {
				delete(user, "accessToken")
			}
			if a.output == "table" {
				return a.print(resp.Data, []string{"owner", "name", "displayName", "email", "phone", "isAdmin", "type"})
			}
			return a.print(resp.Data, nil)
		},
	}
}
