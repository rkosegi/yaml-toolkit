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
	"regexp"
	"strconv"

	"github.com/rkosegi/yaml-toolkit/common"
	"github.com/rkosegi/yaml-toolkit/path"
	"github.com/rkosegi/yaml-toolkit/query"
)

var (
	listPathRe = regexp.MustCompile("\\[\\d+]$")
	nilLeaf    = LeafNode(nil)
)

type containerImpl struct {
	base
	children map[string]Node
}

func flattenLeaf(node Leaf, path string, ret *map[string]Leaf) {
	(*ret)[path] = node
}

func flattenList(node List, path string, ret *map[string]Leaf) {
	for i, item := range node.Items() {
		p := fmt.Sprintf("%s[%d]", path, i)
		if item.IsContainer() {
			flattenContainer(item.AsContainer(), p, ret)
		} else if item.IsList() {
			flattenList(item.AsList(), p, ret)
		} else {
			flattenLeaf(item.AsLeaf(), p, ret)
		}
	}
}

func flattenContainer(node Container, path string, ret *map[string]Leaf) {
	for k, n := range node.Children() {
		p := common.ToPath(path, k)
		if n.IsContainer() {
			flattenContainer(n.AsContainer(), p, ret)
		} else if n.IsList() {
			flattenList(n.AsList(), p, ret)
		} else {
			flattenLeaf(n.AsLeaf(), p, ret)
		}
	}
}

func (c *containerImpl) Flatten() map[string]Leaf {
	ret := make(map[string]Leaf)
	flattenContainer(c, "", &ret)
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

func (c *containerImpl) Search(fn SearchValueFunc) []string {
	var r []string
	for k, v := range c.Flatten() {
		if fn(v.Value()) {
			r = append(r, k)
		}
	}
	return r
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

func (c *containerImpl) Walk(fn NodeVisitorFn) {
	walkContainer(path.NewBuilder(), c, fn)
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

func ensureList(name string, parent ContainerBuilder) (ListBuilder, uint, string) {
	idx := listPathRe.FindStringIndex(name)
	index, _ := strconv.Atoi(name[idx[0]+1 : idx[1]-1])
	name2 := name[0:idx[0]]
	var list ListBuilder
	if l := parent.Child(name2); l == nil {
		list = parent.AddList(name2)
	}
	for i := 0; i <= index; i++ {
		if list.Size() <= i {
			list.Append(nilLeaf)
		}
	}
	return list, uint(index), name2
}

func (c *containerBuilderImpl) add(name string, child Node) {
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
