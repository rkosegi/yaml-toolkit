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
	"sort"

	"github.com/google/go-cmp/cmp"
	"github.com/rkosegi/yaml-toolkit/common"
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
	Type     ModificationType `yaml:"type"`
	Path     string           `yaml:"path"`
	Value    interface{}      `yaml:"value,omitempty"`
	OldValue interface{}      `yaml:"oldValue,omitempty"`
}

func (m *Modification) String() string {
	return fmt.Sprintf("Mod[Type=%s,Path=%s,Value=%v]", m.Type, m.Path, m.Value)
}

func flattenContainer(c dom.Container, path string, res *[]Modification) {
	for k, n := range c.Children() {
		sub := common.ToPath(path, k)
		if n.IsContainer() {
			flattenContainer(n.(dom.Container), sub, res)
		} else if n.IsList() {
			flattenList(n.(dom.List), sub, res)
		} else {
			flattenLeaf(n.(dom.Leaf), sub, res)
		}
	}
}

func flattenList(l dom.List, path string, res *[]Modification) {
	for i, n := range l.Items() {
		sub := fmt.Sprintf("%s[%d]", path, i)
		if n.IsContainer() {
			flattenContainer(n.(dom.Container), sub, res)
		} else if n.IsList() {
			flattenList(n.(dom.List), ToListPath(path, i), res)
		} else {
			flattenLeaf(n.(dom.Leaf), sub, res)
		}
	}
}

func flattenLeaf(l dom.Leaf, path string, res *[]Modification) {
	appendMod(ModAdd, path, l.Value(), nil, res)
}

func flattenNode(node dom.Node, path string, res *[]Modification) {
	if node.IsContainer() {
		flattenContainer(node.(dom.Container), path, res)
	} else if node.IsList() {
		flattenList(node.(dom.List), path, res)
	} else {
		flattenLeaf(node.(dom.Leaf), path, res)
	}
}

func handleExisting(left, right dom.Node, path string, res *[]Modification) {
	if left.IsContainer() && right.IsContainer() {
		diff(left.(dom.Container), right.(dom.Container), path, res)
	} else if left.IsList() && right.IsList() {
		// lists don't merge
		diffList(left.(dom.List), right.(dom.List), path, res)
	} else if left.IsLeaf() && right.IsLeaf() {
		if !cmp.Equal(left.(dom.Leaf).Value(), right.(dom.Leaf).Value()) {
			// update
			appendMod(ModChange, path, left.(dom.Leaf).Value(), right.(dom.Leaf).Value(), res)
		}
	} else {
		// replace (del+add)
		appendMod(ModDelete, path, nil, nil, res)
		flattenNode(right, path, res)
	}
}

func diffList(left, right dom.List, path string, res *[]Modification) {
	if !left.Equals(right) {
		appendMod(ModDelete, path, nil, nil, res)
		flattenList(left, path, res)
	}
}

func appendMod(t ModificationType, path string, val interface{}, oldVal interface{}, res *[]Modification) {
	*res = append(*res, Modification{
		Type:     t,
		Path:     path,
		Value:    val,
		OldValue: oldVal,
	})
}

func diff(left, right dom.Container, path string, res *[]Modification) {
	for k, n := range left.Children() {
		if n2 := right.Child(k); n2 != nil {
			// already exists in right
			handleExisting(n, n2, common.ToPath(path, k), res)
		} else {
			// not found in right Container,so flatten out Node into 1 or more ModAdds Modifications
			flattenNode(n, common.ToPath(path, k), res)
		}
	}
	for k := range right.Children() {
		if n2 := left.Child(k); n2 == nil {
			// k is present in right, but missing in left
			appendMod(ModDelete, common.ToPath(path, k), nil, nil, res)
		}
	}
}

func sortMods(mods []Modification) {
	sort.SliceStable(mods, func(i, j int) bool {
		return mods[i].Path < mods[j].Path
	})
}

// Diff computes semantic difference between 2 Containers
func Diff(left, right dom.Container) *[]Modification {
	var mods []Modification
	path := ""
	diff(left, right, path, &mods)
	// make order of modifications deterministic
	sortMods(mods)
	return &mods
}

// OverlayDocs computes semantic difference between 2 Overlay documents
func OverlayDocs(left, right dom.OverlayDocument) map[string]*[]Modification {
	res := make(map[string]*[]Modification)
	lmap := left.Layers()
	rmap := right.Layers()
	for ln, ll := range lmap {
		if rl, ok := rmap[ln]; ok {
			res[ln] = Diff(ll, rl)
		} else {
			res[ln] = Diff(ll, dom.ContainerNode())
		}
	}
	for rn, rl := range rmap {
		if ll, ok := lmap[rn]; ok {
			res[rn] = Diff(ll, rl)
		} else {
			res[rn] = Diff(dom.ContainerNode(), rl)
		}
	}
	return res
}

// ToListPath like ToPath, but for lists
func ToListPath(path string, index int) string {
	sub := fmt.Sprintf("[%d]", index)
	if len(path) == 0 {
		return sub
	} else {
		return path + sub
	}
}
