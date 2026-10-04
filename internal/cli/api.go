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
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func newApiCommand(a *app) *cobra.Command {
	var data, file string
	var post bool
	cmd := &cobra.Command{
		Use:   "api ACTION [key=value...]",
		Short: "Call any Casdoor API and print the JSON response",
		Long: `Call any Casdoor API, for what the other commands do not cover. ACTION is the path
after /api/, the key=value arguments are the query parameters. The request is a GET, or
a POST when there is a body (--data or --file) or --post.

See the Swagger docs of Casdoor for the APIs: https://door.casdoor.com/swagger`,
		Example: `  casdoor api get-account
  casdoor api get-global-users pageSize=10 p=1
  casdoor api update-user id=casbin/alice columns=displayName --data '{"owner":"casbin","name":"alice","displayName":"Alice"}'`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, _, err := a.client(cmd.Context())
			if err != nil {
				return err
			}

			query := map[string]string{}
			for _, arg := range args[1:] {
				k, v, ok := strings.Cut(arg, "=")
				if !ok {
					return fmt.Errorf("invalid query parameter %q, use key=value", arg)
				}
				query[k] = v
			}

			body := []byte(data)
			if file != "" {
				if file == "-" {
					body, err = io.ReadAll(a.stdin)
				} else {
					body, err = os.ReadFile(file)
				}
				if err != nil {
					return err
				}
			}

			url := c.GetUrl(strings.TrimPrefix(args[0], "/api/"), query)
			var resp []byte
			if post || len(body) > 0 {
				resp, err = c.DoPostBytesRaw(url, "", bytes.NewReader(body))
			} else {
				resp, err = c.DoGetBytesRaw(url)
			}
			if err != nil {
				return err
			}

			var out bytes.Buffer
			if json.Indent(&out, resp, "", "  ") != nil {
				_, err = a.stdout.Write(resp)
				return err
			}
			out.WriteByte('\n')
			_, err = out.WriteTo(a.stdout)
			return err
		},
	}
	cmd.Flags().StringVarP(&data, "data", "d", "", "the request body")
	cmd.Flags().StringVarP(&file, "file", "f", "", "read the request body from a file, - for stdin")
	cmd.Flags().BoolVar(&post, "post", false, "send a POST even without a body")
	return cmd
}
