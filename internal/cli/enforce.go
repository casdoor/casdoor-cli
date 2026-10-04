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

	"github.com/casdoor/casdoor-cli/internal/output"
	"github.com/spf13/cobra"
)

func newEnforceCommand(a *app) *cobra.Command {
	var permission, model, resource, enforcer, owner string
	cmd := &cobra.Command{
		Use:   "enforce REQUEST...",
		Short: "Check a permission request (e.g. sub obj act) with Casbin",
		Long: `Check a request like "alice data1 read" against a permission, a model, a resource
or an enforcer. It prints true when the request is allowed.`,
		Example: `  casdoor enforce --permission casbin/permission-1 alice data1 read
  casdoor enforce --enforcer casbin/enforcer-1 alice data1 read`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := map[string]string{}
			for key, value := range map[string]string{"permissionId": permission, "modelId": model, "resourceId": resource, "enforcerId": enforcer, "owner": owner} {
				if value != "" {
					query[key] = value
				}
			}
			if len(query) == 0 {
				return fmt.Errorf("give one of --permission, --model, --resource, --enforcer or --owner")
			}

			c, _, err := a.client(cmd.Context())
			if err != nil {
				return err
			}
			resp, err := apiPost(c, "enforce", query, args)
			if err != nil {
				return err
			}
			if a.output != "table" {
				return a.print(map[string]any{"results": resp.Data, "ids": resp.Data2}, nil)
			}

			// One result per permission or enforcer that was checked.
			results, _ := resp.Data.([]any)
			ids, _ := resp.Data2.([]any)
			for i, result := range results {
				if i < len(ids) && len(results) > 1 {
					fmt.Fprintf(a.stdout, "%s\t%s\n", output.Cell(ids[i]), output.Cell(result))
				} else {
					fmt.Fprintln(a.stdout, output.Cell(result))
				}
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&permission, "permission", "", "the id of the permission, owner/name")
	cmd.Flags().StringVar(&model, "model", "", "the id of the model, owner/name")
	cmd.Flags().StringVar(&resource, "resource", "", "the id of the resource, owner/name")
	cmd.Flags().StringVar(&enforcer, "enforcer", "", "the id of the enforcer, owner/name")
	cmd.Flags().StringVar(&owner, "owner", "", "check all the permissions of this organization")
	return cmd
}
