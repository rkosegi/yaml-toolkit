/*
Copyright 2023 Richard Kosegi

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package diff

import (
	"fmt"
	"github.com/google/go-cmp/cmp"
	"github.com/rkosegi/yaml-toolkit/dom"
)

type ModificationType string

const (
	// ModChange leaf's value is being updated
	ModChange = ModificationType("Change")
	// ModDelete node is being removed
	ModDelete = ModificationType("Delete")
	// ModAdd leaf value is being added
	ModAdd = ModificationType("Add")
)

type Modification struct {
	Type  ModificationType
	Path  string
	Value dom.Leaf
}

func (m *Modification) String() string {
	return fmt.Sprintf("Mod[Type=%s,Path=%s,Value=%v]", m.Type, m.Path, m.Value)
}

func toPath(path, key string) string {
	if len(path) == 0 {
		return key
	} else {
		return fmt.Sprintf("%s.%s", path, key)
	}
}

// flatten converts dom.Container into series of ModAdd modifications, recursively
func flatten(node dom.Node, path string, res *[]Modification) {
	if node.IsContainer() {
		for k, n := range node.(dom.Container).Children() {
			subpath := toPath(path, k)
			if n.IsContainer() {
				flatten(n.(dom.Container), subpath, res)
			} else {
				appendMod(ModAdd, subpath, n.(dom.Leaf), res)
			}
		}
	} else {
		appendMod(ModAdd, path, node.(dom.Leaf), res)
	}
}

func handleExisting(left, right dom.Node, path string, res *[]Modification) {
	if left.IsContainer() && right.IsContainer() {
		diff(left.(dom.Container), right.(dom.Container), path, res)
	} else if !left.IsContainer() && !right.IsContainer() {
		if !cmp.Equal(left.(dom.Leaf).Value(), right.(dom.Leaf).Value()) {
			// update
			appendMod(ModChange, path, right.(dom.Leaf), res)
		}
	} else {
		// replace (del+add)
		appendMod(ModDelete, path, nil, res)
		flatten(right, path, res)
	}
}

func appendMod(t ModificationType, path string, val dom.Leaf, res *[]Modification) {
	*res = append(*res, Modification{
		Type:  t,
		Path:  path,
		Value: val,
	})
}

func diff(left, right dom.Container, path string, res *[]Modification) {
	for k, n := range left.Children() {
		if n2 := right.Child(k); n2 != nil {
			// already exists in right
			handleExisting(n, n2, toPath(path, k), res)
		} else {
			// not found in right Container,so flatten out Node into 1 or more ModAdds Modifications
			if n.IsContainer() {
				flatten(n.(dom.Container), toPath(path, k), res)
			} else {
				appendMod(ModAdd, toPath(path, k), n.(dom.Leaf), res)
			}
		}
	}
	for k := range right.Children() {
		if n2 := left.Child(k); n2 == nil {
			// k is present in right, but missing in left
			appendMod(ModDelete, toPath(path, k), nil, res)
		}
	}
}

// Diff computes semantic difference between 2 Containers
func Diff(left, right dom.Container) *[]Modification {
	var mods []Modification
	path := ""
	diff(left, right, path, &mods)
	return &mods
}
