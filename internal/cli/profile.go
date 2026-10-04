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
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

func newProfileCommand(a *app) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "profile",
		Aliases: []string{"profiles"},
		Short:   "Manage the saved Casdoor servers and logins",
	}

	cmd.AddCommand(&cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List the profiles",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			current := a.cfg.CurrentName("")
			items := []any{}
			for _, name := range a.cfg.Names() {
				p := a.cfg.Profiles[name]
				items = append(items, map[string]any{
					"current":      name == current,
					"name":         name,
					"endpoint":     p.Endpoint,
					"organization": p.Organization,
					"application":  p.Application,
					"login":        loginState(p.AccessToken, p.ClientSecret, p.Expiry),
				})
			}
			return a.print(items, []string{"current", "name", "endpoint", "organization", "application", "login"})
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "use NAME",
		Short: "Make NAME the current profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, ok := a.cfg.Profiles[args[0]]; !ok {
				return fmt.Errorf("no profile %q", args[0])
			}
			a.cfg.Current = args[0]
			if err := a.cfg.Save(); err != nil {
				return err
			}
			fmt.Fprintf(a.stdout, "Switched to profile %q\n", args[0])
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:     "delete NAME",
		Aliases: []string{"rm"},
		Short:   "Delete a profile and its login",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, ok := a.cfg.Profiles[args[0]]; !ok {
				return fmt.Errorf("no profile %q", args[0])
			}
			delete(a.cfg.Profiles, args[0])
			if a.cfg.Current == args[0] {
				a.cfg.Current = ""
			}
			if err := a.cfg.Save(); err != nil {
				return err
			}
			fmt.Fprintf(a.stdout, "Deleted profile %q\n", args[0])
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "path",
		Short: "Print the path of the config file",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintln(a.stdout, a.cfg.File())
		},
	})
	return cmd
}

func loginState(accessToken, clientSecret string, expiry time.Time) string {
	switch {
	case accessToken != "" && !expiry.IsZero() && time.Now().After(expiry):
		return "expired"
	case accessToken != "":
		return "user"
	case clientSecret != "":
		return "application"
	default:
		return "none"
	}
}
