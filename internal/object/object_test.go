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

package object

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestApplySets(t *testing.T) {
	obj := Object{"name": "alice", "properties": Object{"team": "dev"}}
	err := ApplySets(obj, []string{
		"displayName=Alice Smith",
		"phone=0123",
		"score:=42",
		"isAdmin:=true",
		"roles:=[\"r1\",\"r2\"]",
		"properties.level=senior",
		"address.city=Paris",
		"bio=a=b",
	})
	if err != nil {
		t.Fatal(err)
	}

	want := Object{
		"name":        "alice",
		"displayName": "Alice Smith",
		"phone":       "0123",
		"score":       json.Number("42"),
		"isAdmin":     true,
		"roles":       []any{"r1", "r2"},
		"properties":  Object{"team": "dev", "level": "senior"},
		"address":     Object{"city": "Paris"},
		"bio":         "a=b",
	}
	if !reflect.DeepEqual(obj, want) {
		t.Fatalf("got %#v\nwant %#v", obj, want)
	}
}

func TestApplySetsErrors(t *testing.T) {
	for _, set := range []string{"novalue", "=value", ":=1", "score:=notjson"} {
		if err := ApplySets(Object{}, []string{set}); err == nil {
			t.Errorf("--set %q should fail", set)
		}
	}
}

func TestDecodeAndMerge(t *testing.T) {
	fromJSON, err := Decode([]byte(`{"displayName":"A","properties":{"level":"senior"},"score":7}`))
	if err != nil {
		t.Fatal(err)
	}
	fromYAML, err := Decode([]byte("bio: from yaml\nproperties:\n  team: ops\n"))
	if err != nil {
		t.Fatal(err)
	}

	obj := Object{"name": "alice", "properties": Object{"team": "dev", "x": "y"}}
	Merge(obj, fromJSON)
	Merge(obj, fromYAML)

	want := Object{
		"name":        "alice",
		"displayName": "A",
		"bio":         "from yaml",
		"score":       json.Number("7"),
		"properties":  Object{"team": "ops", "x": "y", "level": "senior"},
	}
	if !reflect.DeepEqual(obj, want) {
		t.Fatalf("got %#v\nwant %#v", obj, want)
	}

	if _, err = Decode([]byte("- a list")); err == nil {
		t.Error("a list is not an object")
	}
}

func TestReadFileStdin(t *testing.T) {
	obj, err := ReadFile("-", strings.NewReader(`{"name":"bob"}`))
	if err != nil {
		t.Fatal(err)
	}
	if String(obj, "name") != "bob" || String(obj, "missing") != "" {
		t.Fatalf("got %v", obj)
	}
}

func TestSplitId(t *testing.T) {
	cases := []struct{ id, owner, name string }{
		{"alice", "org", "alice"},
		{"other/alice", "other", "alice"},
	}
	for _, c := range cases {
		owner, name := SplitId(c.id, "org")
		if owner != c.owner || name != c.name {
			t.Errorf("SplitId(%q) = %q, %q", c.id, owner, name)
		}
	}
}
