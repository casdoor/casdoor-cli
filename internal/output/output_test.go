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
	"bytes"
	"strings"
	"testing"
)

var users = []any{
	map[string]any{"owner": "org", "name": "alice", "email": "a@example.com", "roles": []any{"r1", "r2"}, "isAdmin": true},
	map[string]any{"owner": "org", "name": "bob", "email": strings.Repeat("b", 100)},
}

func TestTable(t *testing.T) {
	var buf bytes.Buffer
	if err := Print(&buf, "table", users, []string{"name", "email", "roles", "isAdmin"}); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 3 {
		t.Fatalf("got %q", buf.String())
	}
	if strings.Join(strings.Fields(lines[0]), " ") != "NAME EMAIL ROLES ISADMIN" {
		t.Errorf("header %q", lines[0])
	}
	if strings.Join(strings.Fields(lines[1]), " ") != "alice a@example.com r1,r2 true" {
		t.Errorf("row %q", lines[1])
	}
	if !strings.Contains(lines[2], "...") || strings.Contains(lines[2], strings.Repeat("b", 100)) {
		t.Errorf("long cells should be cut: %q", lines[2])
	}
}

func TestObjectFields(t *testing.T) {
	var buf bytes.Buffer
	if err := Print(&buf, "table", users[0], nil); err != nil {
		t.Fatal(err)
	}
	want := "email:    a@example.com\nisAdmin:  true\nname:     alice\nowner:    org\nroles:    r1,r2\n"
	if buf.String() != want {
		t.Fatalf("got %q", buf.String())
	}
}

func TestFormats(t *testing.T) {
	var buf bytes.Buffer
	if err := Print(&buf, "name", users, nil); err != nil {
		t.Fatal(err)
	}
	if buf.String() != "org/alice\norg/bob\n" {
		t.Errorf("name: %q", buf.String())
	}

	buf.Reset()
	if err := Print(&buf, "json", users[0], nil); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), `"email": "a@example.com"`) {
		t.Errorf("json: %q", buf.String())
	}

	buf.Reset()
	if err := Print(&buf, "yaml", users[0], nil); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "email: a@example.com") {
		t.Errorf("yaml: %q", buf.String())
	}

	if err := Print(&buf, "xml", users, nil); err == nil {
		t.Error("unknown format should fail")
	}
}
