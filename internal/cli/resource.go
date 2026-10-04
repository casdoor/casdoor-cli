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
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/casdoor/casdoor-cli/internal/object"
	"github.com/casdoor/casdoor-cli/internal/output"
	"github.com/casdoor/casdoor-go-sdk/casdoorsdk"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func newResourceCommand(a *app, r *resource) *cobra.Command {
	cmd := &cobra.Command{
		Use:     r.Plural,
		Aliases: r.Aliases,
		Short:   r.Short,
	}
	cmd.AddCommand(newListCommand(a, r))
	if !r.ListOnly {
		cmd.AddCommand(newGetCommand(a, r), newCreateCommand(a, r), newUpdateCommand(a, r), newDeleteCommand(a, r))
	}
	return cmd
}

func ownerFlag(cmd *cobra.Command, owner *string, r *resource) {
	def := "the organization of the profile"
	if r.AdminOwned {
		def = "admin"
	}
	cmd.Flags().StringVar(owner, "owner", "", "the owner of the "+r.Plural+", "+def+" by default")
}

func (a *app) owner(r *resource, flag, organization string) string {
	if flag != "" {
		return flag
	}
	return r.defaultOwner(organization)
}

func newListCommand(a *app, r *resource) *cobra.Command {
	var owner, search, sortField, columns string
	var page, pageSize int
	var desc bool
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List the " + r.Plural,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, p, err := a.client(cmd.Context())
			if err != nil {
				return err
			}

			query := map[string]string{"owner": a.owner(r, owner, p.Organization)}
			// Casdoor only searches and sorts the paginated lists.
			if pageSize <= 0 && (search != "" || sortField != "") {
				page, pageSize = 1, 1000000
			}
			if pageSize > 0 {
				query["p"] = strconv.Itoa(page)
				query["pageSize"] = strconv.Itoa(pageSize)
			}
			if search != "" {
				field, value, ok := strings.Cut(search, "=")
				if !ok {
					return fmt.Errorf("invalid --search %q, use field=value", search)
				}
				query["field"], query["value"] = field, value
			}
			if sortField != "" {
				query["sortField"] = sortField
				query["sortOrder"] = "ascend"
				if desc {
					query["sortOrder"] = "descend"
				}
			}

			resp, err := apiGet(c, "get-"+r.Plural, query)
			if err != nil {
				return err
			}
			items, _ := resp.Data.([]any)
			if items == nil {
				items = []any{}
			}
			return a.printList(r, items, columns)
		},
	}
	ownerFlag(cmd, &owner, r)
	cmd.Flags().IntVar(&page, "page", 1, "the page to get, with --page-size")
	cmd.Flags().IntVar(&pageSize, "page-size", 0, "the number of items per page, all items by default")
	cmd.Flags().StringVar(&search, "search", "", "only the items whose field contains the value, e.g. --search email=example.com")
	cmd.Flags().StringVar(&sortField, "sort", "", "the field to sort by")
	cmd.Flags().BoolVar(&desc, "desc", false, "sort in descending order")
	cmd.Flags().StringVar(&columns, "columns", "", "the comma-separated columns of the table, e.g. name,email")
	return cmd
}

func (a *app) printList(r *resource, items []any, columns string) error {
	if a.output == "name" {
		for _, item := range items {
			if obj, ok := item.(map[string]any); ok {
				fmt.Fprintf(a.stdout, "%s/%s\n", output.Cell(obj["owner"]), output.Cell(obj[r.nameField()]))
			}
		}
		return nil
	}
	cols := r.columns()
	if columns != "" {
		cols = strings.Split(columns, ",")
	}
	return a.print(items, cols)
}

// fetch gets one object, the name may be a full id.
func (a *app) fetch(cmd *cobra.Command, r *resource, ownerFlagValue, name string) (*objectRef, error) {
	c, p, err := a.client(cmd.Context())
	if err != nil {
		return nil, err
	}
	id := r.id(a.owner(r, ownerFlagValue, p.Organization), name)
	resp, err := apiGet(c, "get-"+r.Singular, map[string]string{r.idParam(): id})
	if err != nil {
		return nil, err
	}
	obj, ok := resp.Data.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%s %s not found", r.Singular, id)
	}
	return &objectRef{id: id, obj: obj, client: c}, nil
}

func newGetCommand(a *app, r *resource) *cobra.Command {
	var owner, columns string
	cmd := &cobra.Command{
		Use:   "get NAME",
		Short: "Show a " + r.Singular + ", NAME may also be the full id like owner/name",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ref, err := a.fetch(cmd, r, owner, args[0])
			if err != nil {
				return err
			}
			if a.output == "name" {
				fmt.Fprintln(a.stdout, ref.id)
				return nil
			}
			var cols []string
			if columns != "" {
				cols = strings.Split(columns, ",")
			}
			return a.print(ref.obj, cols)
		},
	}
	ownerFlag(cmd, &owner, r)
	cmd.Flags().StringVar(&columns, "columns", "", "only show these comma-separated fields")
	return cmd
}

type editFlags struct {
	owner string
	file  string
	sets  []string
}

func (f *editFlags) register(cmd *cobra.Command, r *resource) {
	ownerFlag(cmd, &f.owner, r)
	cmd.Flags().StringVarP(&f.file, "file", "f", "", "a JSON or YAML file with the fields, - for stdin")
	cmd.Flags().StringArrayVar(&f.sets, "set", nil, "set a field: key=value for a string, key:=json for any JSON value (repeatable), nested fields like properties.team=dev")
}

// apply puts the fields of the file and the --set flags into obj.
func (f *editFlags) apply(a *app, obj object.Object) error {
	if f.file != "" {
		fields, err := object.ReadFile(f.file, a.stdin)
		if err != nil {
			return err
		}
		object.Merge(obj, fields)
	}
	return object.ApplySets(obj, f.sets)
}

func newCreateCommand(a *app, r *resource) *cobra.Command {
	var flags editFlags
	cmd := &cobra.Command{
		Use:   "create [NAME]",
		Short: "Create a " + r.Singular + " from --file and --set",
		Example: fmt.Sprintf("  casdoor %s create my-%s --set displayName=\"My %s\"\n  casdoor %s create -f %s.json",
			r.Plural, r.Singular, r.Singular, r.Plural, r.Singular),
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, p, err := a.client(cmd.Context())
			if err != nil {
				return err
			}

			obj := object.Object{}
			if err = flags.apply(a, obj); err != nil {
				return err
			}
			if len(args) == 1 {
				owner, name := object.SplitId(args[0], "")
				if owner != "" {
					obj["owner"] = owner
				}
				obj[r.nameField()] = name
			}
			if flags.owner != "" {
				obj["owner"] = flags.owner
			}
			if object.String(obj, "owner") == "" {
				obj["owner"] = r.defaultOwner(p.Organization)
			}
			if object.String(obj, r.nameField()) == "" {
				return fmt.Errorf("the %s has no %s, give NAME or --set %s=<value>", r.Singular, r.nameField(), r.nameField())
			}
			if object.String(obj, "createdTime") == "" {
				obj["createdTime"] = time.Now().Format(time.RFC3339)
			}

			resp, err := apiPost(c, "add-"+r.Singular, nil, obj)
			if err != nil {
				return err
			}
			if err = checkAffected(resp.Data); err != nil {
				return err
			}
			fmt.Fprintf(a.stdout, "%s %s/%s created\n", r.Singular, object.String(obj, "owner"), object.String(obj, r.nameField()))
			return nil
		},
	}
	flags.register(cmd, r)
	return cmd
}

func newUpdateCommand(a *app, r *resource) *cobra.Command {
	var flags editFlags
	cmd := &cobra.Command{
		Use:     "update NAME",
		Short:   "Update fields of a " + r.Singular + " from --file and --set, the other fields are kept",
		Example: fmt.Sprintf("  casdoor %s update my-%s --set displayName=\"New name\"", r.Plural, r.Singular),
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if flags.file == "" && len(flags.sets) == 0 {
				return fmt.Errorf("nothing to update, use --file or --set")
			}
			ref, err := a.fetch(cmd, r, flags.owner, args[0])
			if err != nil {
				return err
			}
			if err = flags.apply(a, ref.obj); err != nil {
				return err
			}

			resp, err := apiPost(ref.client, "update-"+r.Singular, map[string]string{r.idParam(): ref.id}, ref.obj)
			if err != nil {
				return err
			}
			if err = checkAffected(resp.Data); err != nil {
				return err
			}
			fmt.Fprintf(a.stdout, "%s %s updated\n", r.Singular, ref.id)
			return nil
		},
	}
	flags.register(cmd, r)
	return cmd
}

func newDeleteCommand(a *app, r *resource) *cobra.Command {
	var owner string
	var yes bool
	cmd := &cobra.Command{
		Use:     "delete NAME",
		Aliases: []string{"rm"},
		Short:   "Delete a " + r.Singular,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ref, err := a.fetch(cmd, r, owner, args[0])
			if err != nil {
				return err
			}
			if !yes {
				ok, err := a.confirm(fmt.Sprintf("Delete %s %s?", r.Singular, ref.id))
				if err != nil {
					return err
				}
				if !ok {
					return fmt.Errorf("canceled")
				}
			}

			resp, err := apiPost(ref.client, "delete-"+r.Singular, nil, ref.obj)
			if err != nil {
				return err
			}
			if err = checkAffected(resp.Data); err != nil {
				return err
			}
			fmt.Fprintf(a.stdout, "%s %s deleted\n", r.Singular, ref.id)
			return nil
		},
	}
	ownerFlag(cmd, &owner, r)
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "do not ask for confirmation")
	return cmd
}

// confirm asks a yes/no question, without a terminal it needs --yes.
func (a *app) confirm(question string) (bool, error) {
	if f, ok := a.stdin.(*os.File); !ok || !term.IsTerminal(int(f.Fd())) {
		return false, fmt.Errorf("%s use --yes to confirm when not running in a terminal", question)
	}
	fmt.Fprintf(a.stderr, "%s [y/N] ", question)
	answer, _ := bufio.NewReader(a.stdin).ReadString('\n')
	answer = strings.ToLower(strings.TrimSpace(answer))
	return answer == "y" || answer == "yes", nil
}

// checkAffected turns the "Unaffected" result of add/update/delete into an error.
func checkAffected(data any) error {
	if s, ok := data.(string); ok && s == "Unaffected" {
		return fmt.Errorf("nothing was changed, the object may not exist or have the same fields already")
	}
	return nil
}

type objectRef struct {
	id     string
	obj    object.Object
	client *casdoorsdk.Client
}
