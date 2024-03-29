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
	"github.com/rkosegi/yaml-toolkit/utils"
	"io"
	"slices"
	"strings"
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

type overlayDocument struct {
	names    []string
	overlays map[string]ContainerBuilder
}

func (m *overlayDocument) Search(fn SearchValueFunc) []Coordinate {
	var r []Coordinate
	for _, l := range m.names {
		if paths := m.overlays[l].Search(fn); paths != nil {
			for _, path := range paths {
				r = append(r, &coordinate{
					path:  path,
					layer: l,
				})
			}
		}
	}
	return r
}

func (m *overlayDocument) Layers() []string {
	c := make([]string, len(m.names))
	copy(c, m.names)
	return c
}

func (m *overlayDocument) Put(overlay, path string, value Node) {
	if value.IsContainer() {
		for k, v := range value.(Container).Flatten() {
			m.Put(overlay, utils.ToPath(path, k), v)
		}
	} else {
		current := m.ensureOverlay(overlay)
		components := m.pathComponents(path)
		current = ensurePath(current, components[:len(components)-1])
		current.AddValue(components[len(components)-1], value.(Leaf))
	}
}

func ensurePath(node ContainerBuilder, pc []string) ContainerBuilder {
	for _, component := range pc {
		if listPathRe.MatchString(component) {
			list, index, _ := ensureList(component, node)
			if list.Items()[int(index)] == nilLeaf {
				c := &containerBuilderImpl{}
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
		m.overlays[name] = &containerBuilderImpl{}
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
	appendMap(data, current)
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
	if !n.IsList() && !n.IsContainer() && n.(Leaf).Value() == nil {
		return false
	}
	return true
}

func coalesce(nodes ...Node) Node {
	for _, node := range nodes {
		if hasValue(node) {
			return node
		}
	}
	return nilLeaf
}

func firstValidListItem(idx int, lists ...List) Node {
	for _, list := range lists {
		if len(list.Items()) > idx {
			return list.Items()[idx]
		}
	}
	return nilLeaf
}

func leafMappingFn(n Leaf) interface{} {
	return n.Value()
}

func listMappingFn(n List) []interface{} {
	res := make([]interface{}, len(n.Items()))
	for i, item := range n.Items() {
		if item.IsContainer() {
			res[i] = containerMappingFn(item.(Container))
		} else if item.IsList() {
			res[i] = listMappingFn(item.(List))
		} else {
			res[i] = leafMappingFn(item.(Leaf))
		}
	}
	return res
}

func containerMappingFn(n Container) map[string]interface{} {
	res := map[string]interface{}{}
	for k, v := range n.(Container).Children() {
		if v.IsContainer() {
			res[k] = containerMappingFn(v.(Container))
		} else if v.IsList() {
			res[k] = listMappingFn(v.(List))
		} else {
			res[k] = leafMappingFn(v.(Leaf))
		}
	}
	return res
}

func (m *overlayDocument) Serialize(writer io.Writer, mappingFunc NodeMappingFunc, encFn EncoderFunc) error {
	return encFn(writer, mappingFunc(m.Merged()))
}

func NewOverlayDocument() OverlayDocument {
	return &overlayDocument{
		names:    []string{},
		overlays: map[string]ContainerBuilder{},
	}
}

var _ OverlayDocument = &overlayDocument{}
