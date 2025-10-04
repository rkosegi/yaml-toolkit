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
	"github.com/rkosegi/yaml-toolkit/dom"
	"github.com/rkosegi/yaml-toolkit/path"
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

type (
	DifferContext interface {
		// Append appends modification to the result slice
		Append(mt ModificationType, p path.Path, oldVal, newVal interface{})
		// Flatten flattens out Node into sequence of Modifications.
		Flatten(node dom.Node, pb path.Builder)
		// DefaultListDiff uses default list diffing function
		DefaultListDiff(left, right dom.List, pb path.Builder)
	}
	ListDiffFn func(ctx DifferContext, left, right dom.List, pb path.Builder)
)

func WithListDiffFn(fn ListDiffFn) Opt {
	return func(d *differ) {
		d.ldFn = fn
	}
}

type Opt func(*differ)

func (m *Modification) String() string {
	return fmt.Sprintf("%s[Path=%s,Value=%v]", m.Type, m.Path, m.Value)
}

type differ struct {
	// function to compute difference between 2 lists
	ldFn ListDiffFn
	out  []Modification
}

func (d *differ) DefaultListDiff(left, right dom.List, pb path.Builder) {
	defaultListDiffFn(d, left, right, pb)
}

func (d *differ) Append(mt ModificationType, p path.Path, oldVal, newVal interface{}) {
	d.out = append(d.out, Modification{
		Type:     mt,
		Path:     pc.Serializer().Serialize(p),
		OldValue: oldVal,
		Value:    newVal,
	})
}

func diffLeaves(ctx DifferContext, left, right dom.Leaf, pb path.Builder) {
	// if values of 2 leaves are not equal, then emit ModChange
	if !cmp.Equal(left.AsLeaf().Value(), right.AsLeaf().Value()) {
		ctx.Append(ModChange, pb.Build(), left.AsLeaf().Value(), right.AsLeaf().Value())
	}
}

func (d *differ) diffLists(left dom.List, right dom.List, pb path.Builder) {
	d.ldFn(d, left, right, pb)
}

func (d *differ) diffContainers(left, right dom.Container, pb path.Builder) {
	for k, n := range left.Children() {
		childPath := pb.Append(path.Simple(k))
		if n2 := right.Child(k); n2 != nil {
			// already exists in right
			d.diffNodes(n, n2, childPath)
		} else {
			// not found in right Container,so flatten out n
			d.Flatten(n, childPath)
		}
	}
	for k := range right.Children() {
		if n := left.Child(k); n == nil {
			// k is present in right, but missing in left
			d.Append(ModDelete, pb.Append(path.Simple(k)).Build(), nil, nil)
		}
	}
}

func (d *differ) diffNodes(left, right dom.Node, pb path.Builder) {
	if left.SameAs(right) {
		if left.IsContainer() {
			d.diffContainers(left.AsContainer(), right.AsContainer(), pb)
		} else if left.IsList() {
			d.diffLists(left.AsList(), right.AsList(), pb)
		} else {
			diffLeaves(d, left.AsLeaf(), right.AsLeaf(), pb)
		}
	} else {
		// nodes are of different types. This scenario must be handled by
		//   1, removing old node (left)
		//   2, flattening out new one (right)
		d.Append(ModDelete, pb.Build(), nil, nil)
		d.Flatten(right, pb)
	}
}

func (d *differ) Flatten(node dom.Node, pb path.Builder) {
	if node.IsContainer() {
		for k, n := range node.AsContainer().Children() {
			d.Flatten(n, pb.Append(path.Simple(k)))
		}
	} else if node.IsList() {
		for i, n := range node.AsList().Items() {
			d.Flatten(n, pb.Append(path.Numeric(i)))
		}
	} else {
		d.Append(ModAdd, pb.Build(), nil, node.AsLeaf().Value())
	}
}

func defaultListDiffFn(ctx DifferContext, left, right dom.List, pb path.Builder) {
	if !left.Equals(right) {
		ctx.Append(ModDelete, pb.Build(), nil, nil)
		ctx.Flatten(left, pb)
	}
}

func (d *differ) do(left, right dom.Container) []Modification {
	d.diffContainers(left, right, path.NewBuilder())
	// make order of modifications deterministic
	sort.SliceStable(d.out, func(i, j int) bool {
		return d.out[i].Path < d.out[j].Path
	})
	return d.out
}

// Diff computes difference between 2 Containers
func Diff(left, right dom.Container, opts ...Opt) *[]Modification {
	d := &differ{}
	for _, opt := range append([]Opt{
		WithListDiffFn(defaultListDiffFn),
	}, opts...) {
		opt(d)
	}
	out := d.do(left, right)
	return &out
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
