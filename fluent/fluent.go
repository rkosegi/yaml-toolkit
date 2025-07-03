/*
Copyright 2024 Richard Kosegi

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
	"bytes"
	"io"
	"os"

	"github.com/rkosegi/yaml-toolkit/common"
	"github.com/rkosegi/yaml-toolkit/dom"
	"github.com/rkosegi/yaml-toolkit/utils"
	"gopkg.in/yaml.v3"
)

var b = dom.Builder()

// ConfigHelper allows to load, mutate and save configuration object
type ConfigHelper[T any] interface {
	// Add adds any object.
	Add(doc any /*merge strategy*/) ConfigHelper[T]
	// Load loads given file and merge it into current object
	Load(file string) ConfigHelper[T]
	// Save saves current object into file
	Save(file string) ConfigHelper[T]
	// Mutate performs inline operations on current object
	Mutate(fn func(builder dom.ContainerBuilder)) ConfigHelper[T]
	// Result Get current object
	Result() *T
}

type configHelper[T any] struct {
	c dom.ContainerBuilder
}

func panicIfError(err error) {
	if err != nil {
		panic(err)
	}
}

func any2dom(in any) dom.Container {
	if m, ok := in.(map[string]interface{}); ok {
		return dom.DefaultNodeDecoderFn(m)
	}
	var (
		buf bytes.Buffer
		m   map[string]interface{}
	)
	panicIfError(yaml.NewEncoder(&buf).Encode(in))
	panicIfError(yaml.NewDecoder(&buf).Decode(&m))
	return any2dom(m)
}

func dom2gen[T any](c dom.Container) *T {
	var (
		buf bytes.Buffer
		t   T
	)
	panicIfError(yaml.NewEncoder(&buf).Encode(dom.DefaultNodeEncoderFn(c)))
	panicIfError(yaml.NewDecoder(&buf).Decode(&t))
	return &t
}

func (c *configHelper[T]) Add(doc any) ConfigHelper[T] {
	if dc, ok := doc.(dom.Container); ok {
		c.c = c.c.Merge(dc)
	} else {
		c.c = c.c.Merge(any2dom(doc))
	}
	return c
}

func (c *configHelper[T]) Load(file string) ConfigHelper[T] {
	var (
		f   io.ReadCloser
		err error
		cb  dom.ContainerBuilder
	)
	fdp := common.DefaultFileDecoderProvider(file)
	f, err = utils.FileOpener(file)
	panicIfError(err)
	defer func() {
		_ = f.Close()
	}()
	cb, err = b.FromReader(f, fdp)
	panicIfError(err)
	return c.Add(cb)
}

func (c *configHelper[T]) Save(file string) ConfigHelper[T] {
	fep := common.DefaultFileEncoderProvider(file)
	f, err := os.OpenFile(file, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0o664)
	panicIfError(err)
	defer func() {
		_ = f.Close()
	}()
	panicIfError(fep(f, dom.DefaultNodeEncoderFn(c.c)))
	return c
}

func (c *configHelper[T]) Mutate(fn func(builder dom.ContainerBuilder)) ConfigHelper[T] {
	fn(c.c)
	return c
}

func (c *configHelper[T]) Result() *T {
	return dom2gen[T](c.c)
}

func NewConfigHelper[T any]( /* options */ ) ConfigHelper[T] {
	return &configHelper[T]{
		c: dom.ContainerNode(),
	}
}
