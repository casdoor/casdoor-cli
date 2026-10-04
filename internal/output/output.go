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

package output

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"text/tabwriter"
	"unicode/utf8"

	"gopkg.in/yaml.v3"
)

const maxCellWidth = 60

var Formats = []string{"table", "json", "yaml", "name"}

// Print writes data (an object, a list of objects or any JSON value) in the format. The
// columns are used by "table", all fields are shown when there are none.
func Print(w io.Writer, format string, data any, columns []string) error {
	switch format {
	case "json":
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		encoder.SetEscapeHTML(false)
		return encoder.Encode(data)
	case "yaml":
		encoder := yaml.NewEncoder(w)
		encoder.SetIndent(2)
		if err := encoder.Encode(data); err != nil {
			return err
		}
		return encoder.Close()
	case "name":
		for _, row := range rows(data) {
			fmt.Fprintf(w, "%s/%s\n", Cell(row["owner"]), Cell(row["name"]))
		}
		return nil
	case "table", "":
		return printTable(w, data, columns)
	default:
		return fmt.Errorf("unknown output format %q, use one of: %s", format, strings.Join(Formats, ", "))
	}
}

func rows(data any) []map[string]any {
	switch v := data.(type) {
	case map[string]any:
		return []map[string]any{v}
	case []any:
		res := make([]map[string]any, 0, len(v))
		for _, item := range v {
			if obj, ok := item.(map[string]any); ok {
				res = append(res, obj)
			}
		}
		return res
	default:
		return nil
	}
}

func printTable(w io.Writer, data any, columns []string) error {
	switch v := data.(type) {
	case map[string]any:
		// One object: a field per line.
		return printFields(w, v, columns)
	case []any:
		if len(columns) == 0 {
			columns = allColumns(rows(v))
		}
		tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
		header := make([]string, len(columns))
		for i, column := range columns {
			header[i] = strings.ToUpper(column)
		}
		fmt.Fprintln(tw, strings.Join(header, "\t"))
		for _, item := range v {
			obj, ok := item.(map[string]any)
			if !ok {
				fmt.Fprintln(tw, Cell(item))
				continue
			}
			cells := make([]string, len(columns))
			for i, column := range columns {
				cells[i] = truncate(Cell(obj[column]))
			}
			fmt.Fprintln(tw, strings.Join(cells, "\t"))
		}
		return tw.Flush()
	default:
		_, err := fmt.Fprintln(w, Cell(v))
		return err
	}
}

func printFields(w io.Writer, obj map[string]any, columns []string) error {
	keys := columns
	if len(keys) == 0 {
		for k, v := range obj {
			if !isEmpty(v) {
				keys = append(keys, k)
			}
		}
		sort.Strings(keys)
	}
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	for _, k := range keys {
		fmt.Fprintf(tw, "%s:\t%s\n", k, Cell(obj[k]))
	}
	return tw.Flush()
}

func allColumns(objs []map[string]any) []string {
	seen := map[string]bool{}
	var columns []string
	for _, obj := range objs {
		for k := range obj {
			if !seen[k] {
				seen[k] = true
				columns = append(columns, k)
			}
		}
	}
	sort.Strings(columns)
	return columns
}

func isEmpty(v any) bool {
	switch x := v.(type) {
	case nil:
		return true
	case string:
		return x == ""
	case bool:
		return !x
	case float64:
		return x == 0
	case json.Number:
		return x.String() == "0"
	case []any:
		return len(x) == 0
	case map[string]any:
		return len(x) == 0
	}
	return false
}

// Cell formats a value for a table cell.
func Cell(v any) string {
	switch x := v.(type) {
	case nil:
		return ""
	case string:
		return x
	case []any:
		parts := make([]string, len(x))
		for i, item := range x {
			parts[i] = Cell(item)
		}
		return strings.Join(parts, ",")
	case map[string]any:
		data, _ := json.Marshal(x)
		return string(data)
	default:
		return fmt.Sprint(x)
	}
}

func truncate(s string) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if utf8.RuneCountInString(s) <= maxCellWidth {
		return s
	}
	runes := []rune(s)
	return string(runes[:maxCellWidth-3]) + "..."
}
