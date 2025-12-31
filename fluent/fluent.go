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
	"path/filepath"

	"github.com/rkosegi/yaml-toolkit/common"
	"github.com/rkosegi/yaml-toolkit/dom"
	"github.com/rkosegi/yaml-toolkit/props"
)

// ConfigHelper allows to load, mutate and save configuration object
type ConfigHelper[T any] interface {
	// Add adds any object.
	Add(doc any /*merge strategy*/) ConfigHelper[T]
	// Read reads given io.Reader and process it into T using configured codec
	Read(reader io.Reader) ConfigHelper[T]
	// Load loads given file and merge it into current object
	Load(file string) ConfigHelper[T]
	// Save saves current object into file
	Save(file string) ConfigHelper[T]
	// Mutate performs inline operations on current object
	Mutate(fn func(builder dom.ContainerBuilder)) ConfigHelper[T]
	// Result Get current object
	Result() *T
}

type Opt[T any] func(helper *configHelper[T])

func WithCodec[T any](codec dom.FormatBiCodec) Opt[T] {
	return func(ch *configHelper[T]) {
		ch.codec = codec
	}
}

func WithInitialData[T any](data dom.ContainerBuilder) Opt[T] {
	return func(ch *configHelper[T]) {
		ch.c = data
	}
}

type configHelper[T any] struct {
	codec dom.FormatBiCodec
	c     dom.ContainerBuilder
}

func panicIfError(err error) {
	if err != nil {
		panic(err)
	}
}

func any2dom(in any, codec dom.FormatBiCodec) dom.Container {
	if _, ok := in.(map[string]interface{}); ok {
		return dom.DecodeAnyToNode(in).AsContainer()
	}
	var (
		buf bytes.Buffer
		m   map[string]interface{}
	)
	panicIfError(codec.Encoder()(&buf, in))
	panicIfError(codec.Decoder()(&buf, &m))
	return any2dom(m, codec)
}

func dom2gen[T any](c dom.Container, codec dom.FormatBiCodec) *T {
	var (
		buf bytes.Buffer
		t   T
	)
	panicIfError(codec.Encoder()(&buf, c.AsAny()))
	panicIfError(codec.Decoder()(&buf, &t))
	return &t
}

func (c *configHelper[T]) Add(doc any) ConfigHelper[T] {
	if doc == nil {
		return c
	}
	if dc, ok := doc.(dom.Container); ok {
		c.c = c.c.Merge(dc)
	} else {
		c.c = c.c.Merge(any2dom(doc, c.codec))
	}
	return c
}

func (c *configHelper[T]) readWithDecoder(r io.Reader, dec dom.DecoderFunc) ConfigHelper[T] {
	var t T
	panicIfError(dec(r, &t))
	return c.Add(t)
}

func (c *configHelper[T]) Read(r io.Reader) ConfigHelper[T] {
	return c.readWithDecoder(r, c.codec.Decoder())
}

func (c *configHelper[T]) Load(file string) ConfigHelper[T] {
	var (
		f   io.ReadCloser
		err error
	)
	fdp := DefaultFileDecoderProvider(file)
	f, err = common.FileOpener(file)
	panicIfError(err)
	defer func() {
		_ = f.Close()
	}()
	return c.readWithDecoder(f, fdp)
}

func (c *configHelper[T]) Save(file string) ConfigHelper[T] {
	fep := DefaultFileEncoderProvider(file)
	f, err := os.OpenFile(file, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0o664)
	panicIfError(err)
	defer func() {
		_ = f.Close()
	}()
	panicIfError(fep(f, c.c.AsAny()))
	return c
}

func (c *configHelper[T]) Mutate(fn func(builder dom.ContainerBuilder)) ConfigHelper[T] {
	fn(c.c)
	return c
}

func (c *configHelper[T]) Result() *T {
	return dom2gen[T](c.c, c.codec)
}

func NewConfigHelper[T any](opts ...Opt[T]) ConfigHelper[T] {
	ch := &configHelper[T]{}
	for _, opt := range append([]Opt[T]{
		WithCodec[T](dom.DefaultYamlCodec()),
		WithInitialData[T](dom.ContainerNode()),
	}, opts...) {
		opt(ch)
	}
	return ch
}

// DefaultFileEncoderProvider is FileEncoderProvider that uses file suffix to choose dom.EncoderFunc
func DefaultFileEncoderProvider(file string) dom.EncoderFunc {
	switch filepath.Ext(file) {
	case ".yaml", ".yml":
		return dom.DefaultYamlEncoder
	case ".json":
		return dom.DefaultJsonEncoder
	case ".properties":
		return props.EncoderFn
	default:
		return nil
	}
}

// FileDecoderProvider resolves dom.DecoderFunc for given file.
// If file is not recognized, nil is returned.
type FileDecoderProvider func(file string) dom.DecoderFunc

// DefaultFileDecoderProvider is FileDecoderProvider that uses file suffix to choose dom.DecoderFunc
func DefaultFileDecoderProvider(file string) dom.DecoderFunc {
	switch filepath.Ext(file) {
	case ".yaml", ".yml":
		return dom.DefaultYamlDecoder
	case ".json":
		return dom.DefaultJsonDecoder
	case ".properties":
		return props.DecoderFn
	default:
		return nil
	}
}
