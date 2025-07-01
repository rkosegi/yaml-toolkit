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
	"github.com/rkosegi/yaml-toolkit/path"
)

var matchAllFn = func(p path.Path, parent dom.Node, node dom.Node) bool {
	return true
}

type copier struct {
	replace   bool
	filterFn  dom.NodeVisitorFn
	mergeOpts []dom.MergeOption
}

type CopyParam func(*copier)

func CopyParamFilterFunc(fn dom.NodeVisitorFn) CopyParam {
	return func(p *copier) {
		p.filterFn = fn
	}
}

func CopyParamMergeOptions(mOpts ...dom.MergeOption) CopyParam {
	return func(p *copier) {
		p.mergeOpts = mOpts
	}
}

type CopyMode func(*copier)

// CopyModeReplace Copies source container and removes anything that existed
func CopyModeReplace(param ...CopyParam) CopyMode {
	return func(c *copier) {
		CopyParamFilterFunc(matchAllFn)(c)
		for _, p := range param {
			p(c)
		}
		c.replace = true
	}
}

func CopyModeMerge(param ...CopyParam) CopyMode {
	return func(c *copier) {
		CopyParamFilterFunc(matchAllFn)(c)
		for _, p := range param {
			p(c)
		}
		c.replace = false
	}
}

// Morpher allows fluent modification of document(s)
type Morpher interface {
	// Mutate invokes provided callback to modify dom.Container
	Mutate(func(dom.ContainerBuilder)) Morpher

	// Copy copies content of src dom.Container into Morpher.
	Copy(src dom.Container, mode CopyMode) Morpher

	// Result gets a fresh copy of result. Subsequent invocation of this method will always produce a new copy.
	Result() dom.ContainerBuilder
}

type morpher struct {
	d dom.ContainerBuilder
}

func (c *copier) do(src dom.Container) dom.ContainerBuilder {
	cb := dom.ContainerNode()

	src.Walk(func(p path.Path, parent dom.Node, node dom.Node) bool {
		if c.filterFn(p, parent, node) {
			cb.Set(p, node)
		}
		return true
	})
	return cb
}

func (m *morpher) Copy(src dom.Container, mode CopyMode) Morpher {
	c := &copier{}
	mode(c)
	cp := c.do(src)
	if c.replace {
		m.d = cp
	} else {
		m.d = m.d.Merge(cp, c.mergeOpts...)
	}
	return m
}

func (m *morpher) Mutate(fn func(doc dom.ContainerBuilder)) Morpher {
	fn(m.d)
	return m
}

func (m *morpher) Result() dom.ContainerBuilder {
	c := &copier{filterFn: matchAllFn}
	return c.do(m.d)
}

func NewMorpher() Morpher {
	return &morpher{d: dom.ContainerNode()}
}
