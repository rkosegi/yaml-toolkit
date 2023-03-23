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

package k8s

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/rkosegi/yaml-toolkit/dom"
	"os"
	"strings"
)

type doc struct {
	file string
	item string
	cb   dom.ContainerBuilder
	m    Manifest
	enc  EncodeInternalFn
}

func (e *doc) Document() dom.ContainerBuilder {
	return e.cb
}

func (e *doc) Save() (err error) {
	if err = e.enc(e.m, e.cb); err != nil {
		return err
	}
	if f, err := os.OpenFile(e.file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644); err != nil {
		return err
	} else {
		_, err = e.m.WriteTo(f)
		if err != nil {
			return err
		}
		return f.Close()
	}
}

// YamlDoc loads given k8s manifest file and open embedded YAML document for processing
// if embedded document does not exist, it will be created
func YamlDoc(file, item string) (Document, error) {
	return NewBuilder().Manifest(file).
		Decoder(DecodeEmbeddedDoc(item, dom.DefaultYamlDecoder)).
		Encoder(EncodeEmbeddedDoc(item, dom.DefaultYamlEncoder)).
		Open()
}

// JsonDoc loads given k8s manifest file and open embedded JSON document for processing
// if embedded document does not exist, it will be created
func JsonDoc(file, item string) (Document, error) {
	return NewBuilder().Manifest(file).
		Decoder(DecodeEmbeddedDoc(item, dom.DefaultJsonDecoder)).
		Encoder(EncodeEmbeddedDoc(item, dom.DefaultJsonEncoder)).
		Open()
}

// Properties loads given k8s manifest and then loads embedded java properties
func Properties(file string) (Document, error) {
	return NewBuilder().Manifest(file).
		Decoder(DecodeEmbeddedProps()).
		Encoder(EncodeEmbeddedProps()).
		Open()
}

func DecodeEmbeddedDoc(item string, decFn dom.DecoderFunc) DecodeInternalFn {
	return func(m Manifest) (cb dom.ContainerBuilder, err error) {
		e := m.StringData().Get(item)
		if e != nil {
			if cb, err = dom.Builder().FromReader(strings.NewReader(*e), decFn); err != nil {
				return nil, err
			}
		} else {
			cb = dom.Builder().Container()
		}
		return cb, nil
	}
}

func EncodeEmbeddedDoc(item string, encFn dom.EncoderFunc) EncodeInternalFn {
	return func(m Manifest, node dom.ContainerBuilder) error {
		var buff bytes.Buffer
		if err := node.Serialize(&buff, dom.DefaultNodeMappingFn, encFn); err != nil {
			return err
		}
		m.StringData().Update(item, buff.String())
		return nil
	}
}

func DecodeEmbeddedProps() DecodeInternalFn {
	return func(m Manifest) (dom.ContainerBuilder, error) {
		c := dom.Builder().Container()
		for _, k := range m.StringData().List() {
			c.AddValueAt(k, dom.LeafNode(*m.StringData().Get(k)))
		}
		return c, nil
	}
}

func EncodeEmbeddedProps() EncodeInternalFn {
	return func(m Manifest, node dom.ContainerBuilder) error {
		for k, v := range node.Flatten() {
			m.StringData().Update(k, fmt.Sprintf("%v", v.Value()))
		}
		return nil
	}
}

type Builder interface {
	Manifest(file string) Builder
	Decoder(fn DecodeInternalFn) Builder
	Encoder(fn EncodeInternalFn) Builder
	Open() (Document, error)
}

func NewBuilder() Builder {
	return &builderImpl{}
}

type builderImpl struct {
	file string
	idec DecodeInternalFn
	ienc EncodeInternalFn
}

func (b *builderImpl) Manifest(file string) Builder {
	b.file = file
	return b
}

func (b *builderImpl) Decoder(fn DecodeInternalFn) Builder {
	b.idec = fn
	return b
}

func (b *builderImpl) Encoder(fn EncodeInternalFn) Builder {
	b.ienc = fn
	return b
}

func (b *builderImpl) Open() (Document, error) {
	if b.idec == nil {
		return nil, errors.New("decoder not set")
	}
	if b.ienc == nil {
		return nil, errors.New("enc not set")
	}
	m, err := ManifestFromFile(b.file)
	if err != nil {
		return nil, err
	}
	c, err := b.idec(m)
	if err != nil {
		return nil, err
	}

	return &doc{
		file: b.file,
		cb:   c,
		m:    m,
		enc:  b.ienc,
	}, nil
}
