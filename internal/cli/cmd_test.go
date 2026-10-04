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
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// fakeCasdoor keeps objects by "kind:owner/name" and answers Casdoor's generic API.
type fakeCasdoor struct {
	mu      sync.Mutex
	objects map[string]map[string]any
	auth    []string
	queries []string
}

func (f *fakeCasdoor) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.auth = append(f.auth, r.Header.Get("Authorization"))
	f.queries = append(f.queries, r.URL.RawQuery)

	reply := func(data any, data2 any) {
		_ = json.NewEncoder(w).Encode(map[string]any{"status": "ok", "data": data, "data2": data2})
	}
	fail := func(msg string) {
		_ = json.NewEncoder(w).Encode(map[string]any{"status": "error", "msg": msg})
	}
	action := strings.TrimPrefix(r.URL.Path, "/api/")
	q := r.URL.Query()
	body, _ := io.ReadAll(r.Body)

	switch {
	case action == "get-account":
		reply(map[string]any{"owner": "org", "name": "admin", "accessToken": "secret-token"}, nil)
	case action == "enforce":
		var request []string
		_ = json.Unmarshal(body, &request)
		reply([]bool{strings.Join(request, " ") == "alice data1 read"}, []string{q.Get("permissionId")})
	case action == "get-users":
		var list []any
		for key, obj := range f.objects {
			if strings.HasPrefix(key, "user:"+q.Get("owner")+"/") {
				list = append(list, obj)
			}
		}
		reply(list, nil)
	case action == "get-user", action == "get-organization":
		kind := strings.TrimPrefix(action, "get-")
		obj, ok := f.objects[kind+":"+q.Get("id")]
		if !ok {
			reply(nil, nil)
			return
		}
		reply(obj, nil)
	case strings.HasPrefix(action, "add-"), strings.HasPrefix(action, "update-"), strings.HasPrefix(action, "delete-"):
		verb, kind, _ := strings.Cut(action, "-")
		var obj map[string]any
		if err := json.Unmarshal(body, &obj); err != nil {
			fail(err.Error())
			return
		}
		key := kind + ":" + obj["owner"].(string) + "/" + obj["name"].(string)
		switch verb {
		case "add":
			f.objects[key] = obj
		case "update":
			if q.Get("id") != obj["owner"].(string)+"/"+obj["name"].(string) {
				fail("id mismatch " + q.Get("id"))
				return
			}
			f.objects[key] = obj
		case "delete":
			delete(f.objects, key)
		}
		reply("Affected", nil)
	default:
		fail("unknown action " + action)
	}
}

func setup(t *testing.T) (*fakeCasdoor, func(args ...string) (string, error)) {
	fake := &fakeCasdoor{objects: map[string]map[string]any{}}
	server := httptest.NewServer(fake)
	t.Cleanup(server.Close)

	t.Setenv("CASDOOR_CONFIG", filepath.Join(t.TempDir(), "config.json"))
	t.Setenv("CASDOOR_PROFILE", "")
	t.Setenv("CASDOOR_ENDPOINT", server.URL)
	t.Setenv("CASDOOR_CLIENT_ID", "client")
	t.Setenv("CASDOOR_CLIENT_SECRET", "secret")
	t.Setenv("CASDOOR_ORGANIZATION", "org")
	t.Setenv("CASDOOR_APPLICATION", "")
	t.Setenv("CASDOOR_ACCESS_TOKEN", "")

	run := func(args ...string) (string, error) {
		var out bytes.Buffer
		a := &app{stdin: strings.NewReader(""), stdout: &out, stderr: io.Discard}
		root := newRootCommand(a)
		root.SetArgs(args)
		err := root.Execute()
		return out.String(), err
	}
	return fake, run
}

func TestResourceLifecycle(t *testing.T) {
	fake, run := setup(t)

	out, err := run("users", "create", "alice", "--set", "displayName=Alice", "--set", "score:=3", "--set", "properties.team=dev")
	if err != nil || out != "user org/alice created\n" {
		t.Fatalf("create: %q %v", out, err)
	}
	alice := fake.objects["user:org/alice"]
	if alice["displayName"] != "Alice" || alice["score"] != float64(3) || alice["createdTime"] == "" {
		t.Fatalf("created %v", alice)
	}
	if fake.auth[0] != "Basic Y2xpZW50OnNlY3JldA==" {
		t.Fatalf("the application authenticates with basic auth, got %q", fake.auth[0])
	}

	if _, err = run("users", "update", "alice", "--set", "email=alice@example.com"); err != nil {
		t.Fatalf("update: %v", err)
	}
	alice = fake.objects["user:org/alice"]
	if alice["email"] != "alice@example.com" || alice["displayName"] != "Alice" {
		t.Fatalf("update must keep the other fields: %v", alice)
	}

	if out, err = run("users", "list", "--columns", "name,email"); err != nil || !strings.Contains(out, "alice  alice@example.com") {
		t.Fatalf("list: %q %v", out, err)
	}
	if out, err = run("users", "get", "org/alice", "-o", "json"); err != nil || !strings.Contains(out, `"email": "alice@example.com"`) {
		t.Fatalf("get: %q %v", out, err)
	}
	if out, err = run("users", "list", "-o", "name"); err != nil || out != "org/alice\n" {
		t.Fatalf("list names: %q %v", out, err)
	}

	if _, err = run("users", "delete", "alice"); err == nil {
		t.Fatal("delete without a terminal needs --yes")
	}
	if out, err = run("users", "delete", "alice", "--yes"); err != nil || fake.objects["user:org/alice"] != nil {
		t.Fatalf("delete: %q %v", out, err)
	}
	if _, err = run("users", "get", "alice"); err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("get a deleted user: %v", err)
	}
}

func TestAdminOwned(t *testing.T) {
	fake, run := setup(t)
	if _, err := run("orgs", "create", "acme"); err != nil {
		t.Fatal(err)
	}
	if fake.objects["organization:admin/acme"] == nil {
		t.Fatalf("organizations are owned by admin: %v", fake.objects)
	}
	if _, err := run("orgs", "update", "acme", "--set", "displayName=ACME"); err != nil {
		t.Fatal(err)
	}
}

func TestListSearchNeedsPagination(t *testing.T) {
	fake, run := setup(t)
	if _, err := run("users", "list", "--search", "email=example.com", "--sort", "name", "--desc"); err != nil {
		t.Fatal(err)
	}
	q := fake.queries[len(fake.queries)-1]
	for _, want := range []string{"field=email", "value=example.com", "sortField=name", "sortOrder=descend", "pageSize=", "p=1", "owner=org"} {
		if !strings.Contains(q, want) {
			t.Errorf("query %q has no %q", q, want)
		}
	}
}

func TestWhoamiAndToken(t *testing.T) {
	fake, run := setup(t)
	t.Setenv("CASDOOR_ACCESS_TOKEN", "user-token")
	out, err := run("whoami", "-o", "json")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out, "secret-token") || !strings.Contains(out, `"name": "admin"`) {
		t.Fatalf("whoami: %q", out)
	}
	if fake.auth[len(fake.auth)-1] != "Bearer user-token" {
		t.Fatalf("a logged-in user authenticates with the token, got %q", fake.auth[len(fake.auth)-1])
	}
}

func TestEnforceAndApi(t *testing.T) {
	_, run := setup(t)
	out, err := run("enforce", "--permission", "org/perm", "alice", "data1", "read")
	if err != nil || out != "true\n" {
		t.Fatalf("enforce: %q %v", out, err)
	}
	if out, err = run("enforce", "--permission", "org/perm", "bob", "data1", "read"); err != nil || out != "false\n" {
		t.Fatalf("enforce: %q %v", out, err)
	}
	if _, err = run("enforce", "alice", "data1", "read"); err == nil {
		t.Fatal("enforce needs what to check against")
	}

	if out, err = run("api", "get-account"); err != nil || !strings.Contains(out, `"status": "ok"`) {
		t.Fatalf("api: %q %v", out, err)
	}
	if _, err = run("api", "nothing"); err == nil {
		t.Fatal("api errors are returned")
	}
}

func TestResourceTable(t *testing.T) {
	seen := map[string]bool{}
	for _, r := range resources {
		for _, name := range append([]string{r.Plural}, r.Aliases...) {
			if seen[name] {
				t.Errorf("%q is used twice", name)
			}
			seen[name] = true
		}
	}

	users, sessions := resources[0], (*resource)(nil)
	for _, r := range resources {
		if r.Plural == "sessions" {
			sessions = r
		}
	}
	if users.id("org", "alice") != "org/alice" || users.id("org", "other/alice") != "other/alice" {
		t.Error("user ids")
	}
	if sessions.id("org", "alice/app") != "org/alice/app" || sessions.id("org", "x/alice/app") != "x/alice/app" {
		t.Error("session ids")
	}
}
