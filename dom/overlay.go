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

func (m *overlayDocument) Layers() map[string]Container {
	c := make(map[string]Container, len(m.names))
	for n, v := range m.overlays {
		c[n] = v.Clone().(Container)
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
		for k, v := range value.(Container).Flatten() {
			m.Put(overlay, utils.ToPath(path, k), v)
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

func firstValidListItem(idx int, lists ...List) Node {
	for _, list := range lists {
		if list.Size() > idx {
			return list.Items()[idx]
		}
	}
	return nilLeaf
}

func leafMappingFn(n Leaf) interface{} {
	return n.Value()
}

func listMappingFn(n List) []interface{} {
	res := make([]interface{}, n.Size())
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

func walkNode(layer, path string, parent, node Node, fn OverlayVisitorFn) bool {
	if node.IsContainer() {
		if !walkContainer(layer, path, node.(ContainerBuilder), fn) {
			return false
		}
	} else if node.IsList() {
		if !walkList(layer, path, node.(ListBuilder), fn) {
			return false
		}
	} else {
		if !fn(layer, path, parent, node) {
			return false
		}
	}
	return true
}

func walkList(layer, path string, list ListBuilder, fn OverlayVisitorFn) bool {
	for idx, li := range list.Items() {
		p := utils.ToListPath(path, idx)
		if walkNode(layer, p, list, li, fn) == false {
			return false
		}
	}
	return true
}

func walkContainer(layer, path string, con ContainerBuilder, fn OverlayVisitorFn) bool {
	for k, v := range con.Children() {
		p := utils.ToPath(path, k)
		if walkNode(layer, p, con, v, fn) == false {
			return false
		}
	}
	return true
}

func (m *overlayDocument) Walk(fn OverlayVisitorFn) {
	for _, n := range m.names {
		c := m.overlays[n]
		if !walkContainer(n, "", c, fn) {
			return
		}
	}
}

func NewOverlayDocument() OverlayDocument {
	return &overlayDocument{
		names:    []string{},
		overlays: map[string]ContainerBuilder{},
	}
}

var _ OverlayDocument = &overlayDocument{}
