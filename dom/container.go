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
	"io"
	"reflect"
	"strings"
)

type containerImpl struct {
	children map[string]Node
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

func appendChild(current *map[string]interface{}, parent ContainerBuilder, path string) {
	for k, v := range *current {
		t := reflect.ValueOf(v)
		switch t.Kind() {
		case reflect.Map:
			ref := v.(map[string]interface{})
			appendChild(&ref, parent.AddContainer(k), path+"/"+k)
		case reflect.Int, reflect.Float64, reflect.String:
			parent.AddValue(k, LeafNode(v))
		}
	}
}

type containerFactory struct {
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
		appendChild(&root, &doc, "")
		return &doc, err
	}
}

func Builder() ContainerFactory {
	return &containerFactory{}
}
