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
	"golang.org/x/exp/slices"
	"io"
	"strings"
)

type overlayDocument struct {
	names    []string
	overlays map[string]ContainerBuilder
}

func (m *overlayDocument) Put(overlay, path string, value Leaf) {
	current := m.ensureOverlay(overlay)
	components := m.pathComponents(path)
	for _, component := range components[:len(components)-1] {
		if n := current.Child(component); n == nil {
			current = current.AddContainer(component)
		} else {
			current = n.(ContainerBuilder)
		}
	}
	current.AddValue(components[len(components)-1], value)
}

func (m *overlayDocument) Merged() Container {
	var merged Container
	merged = &containerBuilderImpl{}
	for _, name := range m.names {
		merged = mergeContainers(merged.(ContainerBuilder), m.overlays[name].(ContainerBuilder))
	}
	return merged
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
	for _, comp := range m.pathComponents(path) {
		child := current.Child(comp)
		if child == nil {
			current = current.AddContainer(comp)
		} else {
			current = child.(ContainerBuilder)
		}
	}
	appendChild(data, current, path)
}

func (m *overlayDocument) Lookup(overlay, path string) Node {
	if !slices.Contains(m.names, overlay) {
		return nil
	}
	current := m.overlays[overlay]
	components := m.pathComponents(path)
	if len(components) == 0 || len(path) == 0 {
		return current
	}
	for _, comp := range components[:len(components)-1] {
		if n := current.Child(comp); n == nil {
			return nil
		} else {
			if n, ok := n.(ContainerBuilder); !ok {
				return nil
			} else {
				current = n
			}
		}
	}
	return current.Child(components[len(components)-1])
}

func (m *overlayDocument) LookupAny(path string) Node {
	for _, name := range m.names {
		if n := m.Lookup(name, path); n != nil {
			return n
		}
	}
	return nil
}

func mergeContainers(c1, c2 ContainerBuilder) Container {
	merged := map[string]Node{}
	for k, v := range c1.Children() {
		merged[k] = v
	}
	for k, v := range c2.Children() {
		if n, exists := merged[k]; exists && n.IsContainer() && v.IsContainer() {
			merged[k] = mergeContainers(n.(ContainerBuilder), v.(ContainerBuilder))
		} else {
			merged[k] = v
		}
	}
	r := &containerBuilderImpl{}
	r.children = merged
	return r
}

func leafMappingFn(n Leaf) interface{} {
	return n.Value()
}

func containerMappingFn(n Container) map[string]interface{} {
	res := map[string]interface{}{}
	for k, v := range n.(Container).Children() {
		if v.IsContainer() {
			res[k] = containerMappingFn(v.(Container))
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
