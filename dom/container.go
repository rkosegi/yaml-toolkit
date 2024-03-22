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
	"gopkg.in/yaml.v3"
	"io"
	"reflect"
	"regexp"
	"strconv"
	"strings"
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
	m := *ret
	m[path] = node
}

func flattenList(node List, path string, ret *map[string]Leaf) {
	for i, item := range node.Items() {
		p := fmt.Sprintf("%s[%d]", path, i)
		if item.IsContainer() {
			flattenContainer(item.(Container), p, ret)
		} else if item.IsList() {
			flattenList(item.(List), p, ret)
		} else {
			flattenLeaf(item.(Leaf), p, ret)
		}
	}
}

func flattenContainer(node Container, path string, ret *map[string]Leaf) {
	for k, n := range node.Children() {
		p := utils.ToPath(path, k)
		if n.IsContainer() {
			flattenContainer(n.(Container), p, ret)
		} else if n.IsList() {
			flattenList(n.(List), p, ret)
		} else {
			flattenLeaf(n.(Leaf), p, ret)
		}
	}
}

func (c *containerImpl) Flatten() map[string]Leaf {
	ret := make(map[string]Leaf, 0)
	path := ""
	flattenContainer(c, path, &ret)
	return ret
}

func (c *containerImpl) ensureChildren() {
	if c.children == nil {
		c.children = map[string]Node{}
	}
}

func (c *containerImpl) Child(name string) Node {
	c.ensureChildren()
	if listPathRe.MatchString(name) {
		idx := listPathRe.FindStringIndex(name)
		index, _ := strconv.Atoi(name[idx[0]+1 : idx[1]-1])
		name2 := name[0:idx[0]]
		if n, ok := c.children[name2]; ok {
			if l, ok := n.(List); ok {
				if index > len(l.Items())-1 {
					// index out of bounds
					return nil
				}
				return l.Items()[index]
			} else {
				// not a list
				return nil
			}
		} else {
			// child not exists
			return nil
		}
	}
	return c.children[name]
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

func (c *containerBuilderImpl) Walk(fn WalkFn) {
	c.ensureChildren()
	for k, v := range c.children {
		if v.IsContainer() {
			v.(ContainerBuilder).Walk(fn)
		}
		if !fn(k, c, v) {
			return
		}
	}
}

func (c *containerBuilderImpl) AddList(name string) ListBuilder {
	lb := &listBuilderImpl{}
	c.add(name, lb)
	return lb
}

func (c *containerBuilderImpl) Remove(name string) {
	delete(c.children, name)
}

func (c *containerBuilderImpl) Serialize(writer io.Writer, mappingFunc NodeMappingFunc, encFn EncoderFunc) error {
	return encFn(writer, mappingFunc(c))
}

func (c *containerBuilderImpl) AddContainer(name string) ContainerBuilder {
	cb := &containerBuilderImpl{}
	c.add(name, cb)
	return cb
}

func ensureList(name string, parent ContainerBuilder) (ListBuilder, uint, string) {
	idx := listPathRe.FindStringIndex(name)
	index, _ := strconv.Atoi(name[idx[0]+1 : idx[1]-1])
	name2 := name[0:idx[0]]
	var list ListBuilder
	if l := parent.Child(name2); l == nil {
		list = parent.AddList(name2)
	} else {
		list = l.(ListBuilder)
	}
	for i := 0; i <= index; i++ {
		if len(list.Items()) <= i {
			list.Append(nilLeaf)
		}
	}
	return list, uint(index), name2
}

func (c *containerBuilderImpl) add(name string, child Node) {
	c.ensureChildren()
	if listPathRe.MatchString(name) {
		list, index, _ := ensureList(name, c)
		list.Set(index, child)
	} else {
		c.children[name] = child
	}
}

func (c *containerBuilderImpl) AddValue(name string, value Node) {
	c.add(name, value)
}

func (c *containerBuilderImpl) addChild(parent ContainerBuilder, name string) ContainerBuilder {
	if listPathRe.MatchString(name) {
		list, index, _ := ensureList(name, parent)
		c := &containerBuilderImpl{}
		list.Set(index, c)
		return c
	} else {
		return parent.AddContainer(name)
	}
}

func (c *containerBuilderImpl) ancestorOf(path string, create bool) (ContainerBuilder, string) {
	var node ContainerBuilder
	node = c
	cp := strings.Split(path, ".")
	for _, p := range cp[0 : len(cp)-1] {
		x := node.Child(p)
		if x == nil || !x.IsContainer() {
			if create {
				node = c.addChild(node, p)
			} else {
				return nil, ""
			}
		} else {
			node = x.(ContainerBuilder)
		}
	}
	return node, cp[len(cp)-1]
}

func (c *containerBuilderImpl) AddValueAt(path string, value Node) {
	node, p := c.ancestorOf(path, true)
	node.AddValue(p, value)
}

func (c *containerBuilderImpl) RemoveAt(path string) {
	if node, p := c.ancestorOf(path, false); node != nil {
		node.Remove(p)
	}
}

func appendMap(current *map[string]interface{}, parent ContainerBuilder) {
	for k, v := range *current {
		if v == nil {
			parent.AddValue(k, LeafNode(v))
		} else {
			t := reflect.ValueOf(v)
			switch t.Kind() {
			case reflect.Map:
				ref := v.(map[string]interface{})
				appendMap(&ref, parent.AddContainer(k))
			case reflect.Slice, reflect.Array:
				appendSlice(v.([]interface{}), parent.AddList(k))
			case reflect.Int, reflect.Float64, reflect.String, reflect.Bool:
				parent.AddValue(k, LeafNode(v))
			}
		}
	}
}

func appendSlice(items []interface{}, l ListBuilder) {
	for _, item := range items {
		t := reflect.ValueOf(item)
		switch t.Kind() {
		case reflect.Map:
			ref := item.(map[string]interface{})
			c := &containerBuilderImpl{}
			appendMap(&ref, c)
			l.Append(c)
		case reflect.Slice, reflect.Array:
			list := &listBuilderImpl{}
			appendSlice(item.([]interface{}), list)
			l.Append(list)
		case reflect.Int, reflect.Float64, reflect.String, reflect.Bool:
			l.Append(LeafNode(item))
		}
	}
}

type containerFactory struct {
}

func (f *containerFactory) FromAny(v interface{}) ContainerBuilder {
	var buff strings.Builder
	if err := utils.NewYamlEncoder(&buff).Encode(v); err != nil {
		panic(err)
	}
	var m map[string]interface{}
	if err := yaml.Unmarshal([]byte(buff.String()), &m); err != nil {
		panic(err)
	}
	return f.FromMap(m)
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
		appendMap(&root, &doc)
		return &doc, err
	}
}

func Builder() ContainerFactory {
	return &containerFactory{}
}

var _ ContainerBuilder = &containerBuilderImpl{}
