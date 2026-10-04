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

// Package object edits Casdoor objects as generic JSON objects, so every resource works
// without a Go struct for it.
package object

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Object = map[string]any

// Decode parses a JSON or YAML object.
func Decode(data []byte) (Object, error) {
	obj := Object{}
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) > 0 && trimmed[0] == '{' {
		decoder := json.NewDecoder(bytes.NewReader(trimmed))
		decoder.UseNumber()
		if err := decoder.Decode(&obj); err != nil {
			return nil, err
		}
		return obj, nil
	}
	if err := yaml.Unmarshal(trimmed, &obj); err != nil {
		return nil, err
	}
	if obj == nil {
		return nil, fmt.Errorf("not a JSON or YAML object")
	}
	return obj, nil
}

// ReadFile reads an object from a JSON or YAML file, "-" is stdin.
func ReadFile(path string, stdin io.Reader) (Object, error) {
	var data []byte
	var err error
	if path == "-" {
		data, err = io.ReadAll(stdin)
	} else {
		data, err = os.ReadFile(path)
	}
	if err != nil {
		return nil, err
	}
	obj, err := Decode(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", path, err)
	}
	return obj, nil
}

// Merge copies the fields of src into dst, objects are merged recursively.
func Merge(dst, src Object) {
	for k, v := range src {
		if srcObj, ok := v.(Object); ok {
			if dstObj, ok := dst[k].(Object); ok {
				Merge(dstObj, srcObj)
				continue
			}
		}
		dst[k] = v
	}
}

// ApplySets applies "key=value" (a string) and "key:=json" (any JSON value) assignments.
// Nested fields are separated by dots, like "properties.team=dev".
func ApplySets(obj Object, sets []string) error {
	for _, set := range sets {
		key, value, err := parseSet(set)
		if err != nil {
			return err
		}

		parts := strings.Split(key, ".")
		cur := obj
		for _, part := range parts[:len(parts)-1] {
			next, ok := cur[part].(Object)
			if !ok {
				next = Object{}
				cur[part] = next
			}
			cur = next
		}
		cur[parts[len(parts)-1]] = value
	}
	return nil
}

func parseSet(set string) (string, any, error) {
	i := strings.Index(set, "=")
	if i <= 0 {
		return "", nil, fmt.Errorf("invalid --set %q, use key=value or key:=json", set)
	}
	if set[i-1] != ':' {
		return set[:i], set[i+1:], nil
	}

	key := set[:i-1]
	if key == "" {
		return "", nil, fmt.Errorf("invalid --set %q, use key=value or key:=json", set)
	}
	decoder := json.NewDecoder(strings.NewReader(set[i+1:]))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return "", nil, fmt.Errorf("invalid JSON in --set %q: %w", set, err)
	}
	if v, ok := value.(map[string]any); ok {
		value = Object(v)
	}
	return key, value, nil
}

// String returns a field as a string, empty when it is missing.
func String(obj Object, key string) string {
	switch v := obj[key].(type) {
	case nil:
		return ""
	case string:
		return v
	default:
		return fmt.Sprint(v)
	}
}

// SplitId splits "owner/name", the owner is defaultOwner when the id has no owner.
func SplitId(id, defaultOwner string) (string, string) {
	if i := strings.Index(id, "/"); i >= 0 {
		return id[:i], id[i+1:]
	}
	return defaultOwner, id
}
