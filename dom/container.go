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
	"reflect"
	"strings"
)

type containerImpl struct {
	children map[string]Node
}

func flattenInto(node Container, path string, ret *map[string]Leaf) {
	for k, n := range node.Children() {
		p := utils.ToPath(path, k)
		if !n.IsContainer() {
			m := *ret
			m[p] = n.(Leaf)
		} else {
			flattenInto(n.(Container), p, ret)
		}
	}
}

func (c *containerImpl) Flatten() map[string]Leaf {
	ret := make(map[string]Leaf, 0)
	path := ""
	flattenInto(c, path, &ret)
	return ret
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

func (c *containerImpl) IsContainer() bool {
	return true
}

func (c *containerImpl) Children() map[string]Node {
	c.ensureChildren()
	return c.children
}

func (c *containerImpl) Lookup(path string) Node {
	if path == "" {
		return nil
	}
	c.ensureChildren()
	pc := strings.Split(path, ".")
	var current Container
	current = c
	for _, p := range pc[0 : len(pc)-1] {
		x := current.Child(p)
		if x == nil || !x.IsContainer() {
			return nil
		} else {
			current = x.(Container)
		}
	}
	return current.Child(pc[len(pc)-1])
}

type containerBuilderImpl struct {
	containerImpl
}

func (c *containerBuilderImpl) Remove(name string) {
	delete(c.children, name)
}

func (c *containerBuilderImpl) Serialize(writer io.Writer, mappingFunc NodeMappingFunc, encFn EncoderFunc) error {
	return encFn(writer, mappingFunc(c))
}

func (c *containerBuilderImpl) AddContainer(name string) ContainerBuilder {
	c.ensureChildren()
	cb := &containerBuilderImpl{}
	c.children[name] = cb
	return cb
}

func (c *containerBuilderImpl) AddValue(name string, value Leaf) {
	c.ensureChildren()
	c.children[name] = value
}

func (c *containerBuilderImpl) ancestorOf(path string, create bool) (ContainerBuilder, string) {
	var node ContainerBuilder
	node = c
	cp := strings.Split(path, ".")
	for _, p := range cp[0 : len(cp)-1] {
		x := node.Child(p)
		if x == nil || !x.IsContainer() {
			if create {
				node = node.AddContainer(p)
			} else {
				return nil, ""
			}
		} else {
			node = x.(ContainerBuilder)
		}
	}
	return node, cp[len(cp)-1]
}

func (c *containerBuilderImpl) AddValueAt(path string, value Leaf) {
	node, p := c.ancestorOf(path, true)
	node.AddValue(p, value)
}

func (c *containerBuilderImpl) RemoveAt(path string) {
	if node, p := c.ancestorOf(path, false); node != nil {
		node.Remove(p)
	}
}

func appendChild(current *map[string]interface{}, parent ContainerBuilder) {
	for k, v := range *current {
		if v == nil {
			parent.AddValue(k, LeafNode(v))
		} else {
			t := reflect.ValueOf(v)
			switch t.Kind() {
			case reflect.Map:
				ref := v.(map[string]interface{})
				appendChild(&ref, parent.AddContainer(k))
			case reflect.Int, reflect.Float64, reflect.String, reflect.Bool:
				parent.AddValue(k, LeafNode(v))
			}
		}
	}
}

type containerFactory struct {
}

func (f *containerFactory) FromMap(in map[string]interface{}) ContainerBuilder {
	b := f.Container()
	for k, v := range in {
		b.AddValueAt(k, LeafNode(v))
	}
	return b
}

func (f *containerFactory) Container() ContainerBuilder {
	return &containerBuilderImpl{}
}

func (f *containerFactory) FromReader(r io.Reader, fn DecoderFunc) (ContainerBuilder, error) {
	var root map[string]interface{}
	if err := fn(r, &root); err != nil {
		return nil, err
	} else {
		doc := containerBuilderImpl{}
		appendChild(&root, &doc)
		return &doc, err
	}
}

func Builder() ContainerFactory {
	return &containerFactory{}
}
