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
	"encoding/base64"
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"os"
)

const (
	keyData       = "data"
	keyStringData = "stringData"
	keyBinaryData = "binaryData"
)

var (
	errKindMissing = fmt.Errorf("'kind' elelemt is not present in the root of document")
)

type handler interface {
	afterLoad(k *manifest) error
	beforeSave(k *manifest)
}

func afterLoadBinary(k *manifest, srcKey string) error {
	if data, ok := k.doc[srcKey].(map[string]interface{}); ok {
		for dk, dv := range data {
			if ed, err := base64.StdEncoding.DecodeString(dv.(string)); err != nil {
				return err
			} else {
				k.binData[dk] = ed
			}
		}
	}
	return nil
}

func afterLoadString(k *manifest, srcKey string) {
	if data, ok := k.doc[srcKey].(map[string]interface{}); ok {
		for dk, dv := range data {
			k.strData[dk] = dv.(string)
		}
	}
}

func beforeSaveBinary(k *manifest, key string) {
	if len(k.binData) == 0 {
		delete(k.doc, key)
	} else {
		k.doc[key] = map[string]interface{}{}
		for dk, dv := range k.binData {
			k.doc[key].(map[string]interface{})[dk] = base64.StdEncoding.EncodeToString(dv)
		}
	}
}

func beforeSaveString(k *manifest, key string) {
	if len(k.strData) == 0 {
		delete(k.doc, key)
	} else {
		k.doc[key] = map[string]interface{}{}
		for dk, dv := range k.strData {
			k.doc[key].(map[string]interface{})[dk] = dv
		}
	}
}

type dataHandler struct {
	bk string
	tk string
}

func (d *dataHandler) afterLoad(k *manifest) error {
	if err := afterLoadBinary(k, d.bk); err != nil {
		return err
	}
	afterLoadString(k, d.tk)
	return nil
}

func (d *dataHandler) beforeSave(k *manifest) {
	beforeSaveBinary(k, d.bk)
	beforeSaveString(k, d.tk)
}

type binaryDataFacade struct {
	m *manifest
}

type manifest struct {
	doc       map[string]interface{}
	strData   map[string]string
	binData   map[string][]byte
	strDataIf *stringDataFacade
	binDataIf *binaryDataFacade
	h         handler
}

func (b *binaryDataFacade) Get(name string) []byte {
	return b.m.binData[name]
}

func (b *binaryDataFacade) List() []string {
	var ret []string
	for k := range b.m.binData {
		ret = append(ret, k)
	}
	return ret
}

func (b *binaryDataFacade) Remove(key string) {
	delete(b.m.binData, key)
}

func (b *binaryDataFacade) Update(key string, value []byte) {
	b.m.binData[key] = value
}

type stringDataFacade struct {
	m *manifest
}

func (s *stringDataFacade) Get(name string) *string {
	if d, ok := s.m.strData[name]; !ok {
		return nil
	} else {
		return &d
	}
}

func (s *stringDataFacade) List() []string {
	var ret []string
	for k := range s.m.strData {
		ret = append(ret, k)
	}
	return ret
}

func (s *stringDataFacade) Remove(name string) {
	delete(s.m.strData, name)
}

func (s *stringDataFacade) Update(name, value string) {
	s.m.strData[name] = value
}

func (k *manifest) StringData() StringData {
	return k.strDataIf
}

func (k *manifest) BinaryData() BinaryData {
	return k.binDataIf
}

// WriteTo writes contents of this manifest into io.Writer
func (k *manifest) WriteTo(w io.Writer) (read int64, err error) {
	k.h.beforeSave(k)
	if data, err := yaml.Marshal(k.doc); err != nil {
		return 0, err
	} else {
		n, err := w.Write(data)
		return int64(n), err
	}
}

func ManifestFromReader(r io.Reader) (Manifest, error) {
	var buff bytes.Buffer
	if _, err := buff.ReadFrom(r); err != nil {
		return nil, err
	} else {
		return ManifestFromBytes(buff.Bytes())
	}
}

func ManifestFromFile(file string) (Manifest, error) {
	if data, err := os.ReadFile(file); err != nil {
		return nil, err
	} else {
		return ManifestFromBytes(data)
	}
}

func ManifestFromBytes(data []byte) (Manifest, error) {
	var doc map[string]interface{}
	err := yaml.Unmarshal(data, &doc)
	if err != nil {
		return nil, err
	}
	var tk, bk string

	if kind, ok := doc["kind"]; ok {
		if kind.(string) == "Secret" {
			bk = keyData
			tk = keyStringData
		} else if kind.(string) == "ConfigMap" {
			bk = keyBinaryData
			tk = keyData
		} else {
			return nil, fmt.Errorf("unsupported manifest kind: %s", kind)
		}
	} else {
		return nil, errKindMissing
	}

	m := &manifest{
		doc: doc,
		h: &dataHandler{
			bk: bk,
			tk: tk,
		},
		binData: map[string][]byte{},
		strData: map[string]string{},
	}
	m.strDataIf = &stringDataFacade{m: m}
	m.binDataIf = &binaryDataFacade{m: m}
	if err = m.h.afterLoad(m); err != nil {
		return nil, err
	} else {
		return m, nil
	}
}
