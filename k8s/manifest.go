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
	"encoding/base64"
	"fmt"
	"os"
	"sigs.k8s.io/yaml"
)

const (
	keyData       = "data"
	keyStringData = "stringData"
)

// ManifestContext is in-memory representation of k8s Secret/ConfigMap manifest
// it allow to perform operation over `data` or `stringData` maps
type ManifestContext interface {
	// ListBinary gets all binary data
	ListBinary() map[string][]byte
	// List gets all data
	List() map[string]string
	// Remove removes data
	Remove(name string)
	// Update adds or updates data
	Update(name, value string)
	// UpdateBinary adds or updates binary data
	UpdateBinary(name string, value []byte)
	// RemoveBinary remove binary data
	RemoveBinary(name string)
	// ToBinary encodes given data item into binary form and removes it from data map
	ToBinary(name string) error
	// Save saves content of manifest to file
	Save() error
}

type k8sHelperImpl struct {
	doc  map[string]interface{}
	file string
}

func (k *k8sHelperImpl) ensureMap(key string) {
	if _, ok := k.doc[key].(map[string]interface{}); !ok {
		k.doc[key] = map[string]interface{}{}
	}
}

func (k *k8sHelperImpl) tidy() {
	// remove map if it has no items inside
	for _, key := range []string{keyData, keyStringData} {
		if data, ok := k.doc[key].(map[string]string); ok {
			if len(data) == 0 {
				delete(k.doc, key)
			}
		}
	}
}

func (k *k8sHelperImpl) ToBinary(name string) error {
	k.ensureMap(keyData)
	k.ensureMap(keyStringData)
	defer k.tidy()
	if v, ok := k.doc[keyStringData].(map[string]string)[name]; ok {
		k.UpdateBinary(name, []byte(v))
		k.Remove(name)
		return nil
	} else {
		return fmt.Errorf("data item not found: %s", name)
	}
}

func (k *k8sHelperImpl) ListBinary() map[string][]byte {
	k.ensureMap(keyData)
	ret := map[string][]byte{}
	if m, ok := k.doc[keyData].(map[string]string); ok {
		for key, val := range m {
			// error should not occur here, value is guaranteed to be fine
			v, _ := base64.StdEncoding.DecodeString(val)
			ret[key] = v
		}
	}
	return ret
}

func convertMap(in map[string]interface{}) map[string]string {
	ret := map[string]string{}
	for key, val := range in {
		ret[key] = val.(string)
	}
	return ret
}

func (k *k8sHelperImpl) List() map[string]string {
	k.ensureMap(keyStringData)
	return convertMap(k.doc[keyStringData].(map[string]interface{}))
}

func (k *k8sHelperImpl) remove(name, key string) {
	defer k.tidy()
	if data, ok := k.doc[key].(map[string]string); ok {
		delete(data, name)
	}
}

func (k *k8sHelperImpl) Remove(name string) {
	k.remove(name, keyStringData)
}

func (k *k8sHelperImpl) RemoveBinary(name string) {
	k.remove(name, keyData)
}

func (k *k8sHelperImpl) update(key, name, value string) {
	k.ensureMap(key)
	k.doc[key].(map[string]interface{})[name] = value
}

func (k *k8sHelperImpl) Update(name, value string) {
	k.update(keyStringData, name, value)
}

func (k *k8sHelperImpl) UpdateBinary(name string, value []byte) {
	k.update(keyData, name, base64.StdEncoding.EncodeToString(value))
}

func (k *k8sHelperImpl) Save() error {
	k.tidy()
	if bytes, err := yaml.Marshal(k.doc); err != nil {
		return err
	} else {
		return os.WriteFile(k.file, bytes, 0o644)
	}
}

func NewHelper(file string) (ManifestContext, error) {
	var doc map[string]interface{}
	bytes, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(bytes, &doc)
	if err != nil {
		return nil, err
	}
	return &k8sHelperImpl{
		file: file,
		doc:  doc,
	}, nil
}
