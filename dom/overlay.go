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

package dom

import (
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/rkosegi/yaml-toolkit/common"
	"github.com/rkosegi/yaml-toolkit/path"
)

type coordinate struct {
	path  string
	layer string
}

func (c *coordinate) Layer() string {
	return c.layer
}

func (c *coordinate) Path() string {
	return c.path
}

func (cs Coordinates) String() string {
	sb := strings.Builder{}
	sb.WriteString("[")
	for _, c := range cs {
		sb.WriteString("[")
		sb.WriteString(fmt.Sprintf("layer=%s,path=%s", c.Layer(), c.Path()))
		sb.WriteString("],")
	}
	sb.WriteString("]\n")
	return strings.ReplaceAll(sb.String(), "],]", "]]")
}

type overlayDocument struct {
	names    []string
	overlays map[string]ContainerBuilder
}

func (m *overlayDocument) Search(fn SearchValueFunc) Coordinates {
	var r Coordinates
	for _, l := range m.names {
		if paths := m.overlays[l].Search(fn); paths != nil {
			for _, p := range paths {
				r = append(r, &coordinate{
					path:  p,
					layer: l,
				})
			}
		}
	}
	return r
}

func (m *overlayDocument) Layers() map[string]Container {
	c := make(map[string]Container, len(m.names))
	for n, v := range m.overlays {
		c[n] = v.Clone().AsContainer()
	}
	return c
}

func (m *overlayDocument) LayerNames() []string {
	names := make([]string, len(m.names))
	copy(names, m.names)
	return names
}

func (m *overlayDocument) Add(overlay string, value Container) {
	cb := m.ensureOverlay(overlay)
	for k, v := range value.Children() {
		cb.AddValue(k, v)
	}
}

func (m *overlayDocument) Put(overlay, path string, value Node) {
	if value.IsContainer() {
		for k, v := range value.AsContainer().Flatten() {
			m.Put(overlay, common.ToPath(path, k), v)
		}
	} else {
		current := m.ensureOverlay(overlay)
		components := m.pathComponents(path)
		current = ensurePath(current, components[:len(components)-1])
		current.AddValue(components[len(components)-1], value)
	}
}

func ensurePath(node ContainerBuilder, pc []string) ContainerBuilder {
	for _, component := range pc {
		if listPathRe.MatchString(component) {
			list, index, _ := ensureList(component, node)
			if list.Get(int(index)) == nilLeaf {
				c := initContainerBuilder()
				list.Set(index, c)
				node = c
				continue
			}
		}
		if n := node.Child(component); n == nil {
			node = node.AddContainer(component)
		} else {
			node = n.(ContainerBuilder)
		}
	}
	return node
}

func (m *overlayDocument) Merged(opts ...MergeOption) Container {
	mg := &merger{}
	mg.init(opts...)
	return mg.mergeOverlay(m)
}

func (m *overlayDocument) ensureOverlay(name string) ContainerBuilder {
	if m.overlays[name] == nil {
		m.overlays[name] = initContainerBuilder()
		m.names = append(m.names, name)
	}
	return m.overlays[name]
}

func (m *overlayDocument) pathComponents(path string) []string {
	return strings.Split(path, ".")
}

func (m *overlayDocument) Populate(overlay, path string, data *map[string]interface{}) {
	current := m.ensureOverlay(overlay)
	if path != "" {
		current = ensurePath(current, m.pathComponents(path))
	}
	decodeContainerFn(data, current)
}

func (m *overlayDocument) Lookup(overlay, path string) Node {
	if !slices.Contains(m.names, overlay) {
		return nil
	}
	return m.overlays[overlay].Lookup(path)
}

func (m *overlayDocument) LookupAny(path string) Node {
	for _, name := range m.names {
		if n := m.Lookup(name, path); n != nil {
			return n
		}
	}
	return nil
}

func hasValue(n Node) bool {
	if n == nil || n == nilLeaf {
		return false
	}
	if !n.IsList() && !n.IsContainer() && n.AsLeaf().Value() == nil {
		return false
	}
	return true
}

func firstValidListItem(idx int, lists ...List) Node {
	for _, list := range lists {
		if list.Size() > idx {
			return list.Get(idx)
		}
	}
	return nilLeaf
}

func (m *overlayDocument) Serialize(writer io.Writer, mappingFunc NodeEncoderFunc, encFn EncoderFunc) error {
	return encFn(writer, mappingFunc(m.Merged()))
}

func (m *overlayDocument) Walk(fn OverlayVisitorFn) {
	for _, n := range m.names {
		walkContainer(path.NewBuilder(), m.overlays[n], func(p path.Path, parent Node, node Node) bool {
			return fn(n, p, parent, node)
		})
	}
}

func NewOverlayDocument() OverlayDocument {
	return &overlayDocument{
		names:    []string{},
		overlays: map[string]ContainerBuilder{},
	}
}

var _ OverlayDocument = &overlayDocument{}
