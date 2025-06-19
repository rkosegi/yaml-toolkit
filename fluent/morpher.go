/*
Copyright 2025 Richard Kosegi

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

package fluent

import (
	"github.com/rkosegi/yaml-toolkit/dom"
)

type FilterPredicate func(path string, leaf dom.Leaf) bool

// Morpher allows fluent modification of document(s)
type Morpher interface {
	// AddWithFilter conditionally adds nodes from source dom.Container if they match predicate function
	AddWithFilter(srcDoc dom.Container, fn FilterPredicate) Morpher

	// Set just replaces current builder with content from provided dom.Container
	Set(srcDoc dom.Container) Morpher

	// Mutate invokes provided callback to modify document
	Mutate(func(dom.ContainerBuilder)) Morpher

	// Result gets a fresh copy of result. Subsequent invocation of this method will always produce a new copy.
	Result() dom.ContainerBuilder
}

type morpher struct {
	d dom.ContainerBuilder
}

func (m *morpher) Mutate(fn func(doc dom.ContainerBuilder)) Morpher {
	fn(m.d)
	return m
}

func (m *morpher) Result() dom.ContainerBuilder {
	return dom.Builder().From(m.d.Clone().AsContainer())
}

func (m *morpher) Set(srcDoc dom.Container) Morpher {
	m.d = dom.Builder().From(srcDoc)
	return m
}

func (m *morpher) AddWithFilter(src dom.Container, fn FilterPredicate) Morpher {
	for k, v := range src.Flatten() {
		if fn(k, v) {
			m.d.AddValueAt(k, v)
		}
	}
	return m
}

func NewMorpher() Morpher {
	return (&morpher{}).Set(dom.Builder().Container())
}
