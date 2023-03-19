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
	"github.com/rkosegi/yaml-toolkit/dom"
	"os"
	"strings"
)

type doc struct {
	file string
	item string
	cb   dom.ContainerBuilder
	m    Manifest
	enc  dom.EncoderFunc
}

func (e *doc) Document() dom.ContainerBuilder {
	return e.cb
}

func (e *doc) Save() error {
	var buff bytes.Buffer
	if err := e.cb.Serialize(&buff, dom.DefaultNodeMappingFn, e.enc); err != nil {
		return err
	}
	e.m.StringData().Update(e.item, buff.String())
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
	return embeddedDoc(file, item, dom.DefaultYamlEncoder, dom.DefaultYamlDecoder)
}

// JsonDoc loads given k8s manifest file and open embedded JSON document for processing
// if embedded document does not exist, it will be created
func JsonDoc(file, item string) (Document, error) {
	return embeddedDoc(file, item, dom.DefaultJsonEncoder, dom.DefaultJsonDecoder)
}

func embeddedDoc(file, item string, encFn dom.EncoderFunc, decFn dom.DecoderFunc) (Document, error) {
	m, err := ManifestFromFile(file)
	if err != nil {
		return nil, err
	}
	var cb dom.ContainerBuilder
	e := m.StringData().Get(item)
	if e != nil {
		if cb, err = dom.Builder().FromReader(strings.NewReader(*e), decFn); err != nil {
			return nil, err
		}
	} else {
		cb = dom.Builder().Container()
	}
	return &doc{
		file: file,
		item: item,
		cb:   cb,
		m:    m,
		enc:  encFn,
	}, nil
}
