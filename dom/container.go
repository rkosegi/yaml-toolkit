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
	"github.com/rkosegi/yaml-toolkit/path"
	"github.com/rkosegi/yaml-toolkit/query"
)

var (
	nilLeaf = LeafNode(nil)
)

type containerImpl struct {
	base
	children map[string]Node
}

// SimplePathAsString just delegates to fmt.Stringer() implemented by path.Path
func SimplePathAsString(p path.Path) string {
	return p.String()
}

func (c *containerImpl) Flatten(keyFn PathToStringFunc) map[string]Leaf {
	ret := make(map[string]Leaf)
	c.Walk(func(p path.Path, parent Node, node Node) bool {
		if node.IsLeaf() {
			ret[keyFn(p)] = node.AsLeaf()
		}
		return true
	}, WalkOptDFS())
	return ret
}

func (c *containerImpl) AsAny() any {
	return encodeContainerFn(c)
}

func (c *containerImpl) Equals(node Node) bool {
	if node == nil || !node.IsContainer() {
		return false
	}
	for k, v := range c.children {
		other := node.AsContainer().Child(k)
		if other == nil || !v.Equals(other) {
			return false
		}
	}
	return true
}

func (c *containerImpl) ensureChildren() {
	if c.children == nil {
		c.children = map[string]Node{}
	}
}

func (c *containerImpl) Child(name string) Node {
	c.ensureChildren()
	return c.children[name]
}

func (c *containerImpl) Query(qry query.Query) NodeList {
	return decodeQueryResult(qry.Select(c.AsAny()))
}

func (c *containerImpl) IsContainer() bool {
	return true
}

func (c *containerImpl) SameAs(node Node) bool {
	return node != nil && node.IsContainer()
}

func (c *containerImpl) Children() map[string]Node {
	c.ensureChildren()
	return c.children
}

func (c *containerImpl) Get(p path.Path) Node {
	return getFromNode(c, p)
}

func (c *containerImpl) Clone() Node {
	c2 := initContainer()
	c2.ensureChildren()
	for k, v := range c.children {
		c2.children[k] = v.Clone()
	}
	return c2
}

func (c *containerImpl) Walk(fn NodeVisitorFn, opts ...WalkOpt) {
	walkStart(fn, c, opts...)
}

func (c *containerImpl) AsContainer() Container {
	return c
}

type containerBuilderImpl struct {
	containerImpl
}

func initContainer() *containerImpl {
	cb := &containerImpl{}
	cb.desc = "container"
	return cb
}

func initContainerBuilder() *containerBuilderImpl {
	cb := &containerBuilderImpl{}
	cb.desc = "writable container"
	return cb
}

func (c *containerBuilderImpl) AsContainer() Container {
	return c
}

func (c *containerBuilderImpl) Seal() Container {
	return &c.containerImpl
}

func (c *containerBuilderImpl) Merge(other Container, opts ...MergeOption) ContainerBuilder {
	m := &merger{}
	m.init(opts...)
	return m.mergeContainers(c, other)
}

func (c *containerBuilderImpl) AddList(name string) ListBuilder {
	lb := initListBuilder()
	c.add(name, lb)
	return lb
}

func (c *containerBuilderImpl) Remove(name string) ContainerBuilder {
	delete(c.children, name)
	return c
}

func (c *containerBuilderImpl) AddContainer(name string) ContainerBuilder {
	cb := initContainerBuilder()
	c.add(name, cb)
	return cb
}

func (c *containerBuilderImpl) Set(p path.Path, node Node) ContainerBuilder {
	applyToNode(c, p, node)
	return c
}

func (c *containerBuilderImpl) Delete(p path.Path) ContainerBuilder {
	removeFromNode(c, p)
	return c
}

func (c *containerBuilderImpl) add(name string, child Node) {
	if child == nil {
		child = nilLeaf
	}
	c.ensureChildren()
	c.children[name] = child
}

func (c *containerBuilderImpl) AddValue(name string, value Node) ContainerBuilder {
	c.add(name, value)
	return c
}

// ContainerNode creates new ContainerBuilder
func ContainerNode() ContainerBuilder {
	return initContainerBuilder()
}
