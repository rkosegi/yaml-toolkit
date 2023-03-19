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
	"github.com/rkosegi/yaml-toolkit/dom"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestLoadEmbeddedYaml(t *testing.T) {
	d, err := YamlDoc("../testdata/cm1.yaml", "application.yaml")
	assert.Nil(t, err)
	assert.NotNil(t, d)
	c1 := d.Document().Child("abc").(dom.Container)
	assert.NotNil(t, c1)
	assert.True(t, c1.IsContainer())
	c2 := c1.Child("def").(dom.Container)
	assert.NotNil(t, c2)
	assert.True(t, c2.IsContainer())
	n := c2.Child("leaf2").(dom.Leaf)
	assert.NotNil(t, n)
	assert.Equal(t, "Hello", n.Value())
}

func TestLoadEmbeddedJson(t *testing.T) {
	d, err := JsonDoc("../testdata/cm2.yaml", "config.json")
	assert.Nil(t, err)
	assert.NotNil(t, d)
	c1 := d.Document().Child("key1").(dom.Container)
	assert.NotNil(t, c1)
	assert.True(t, c1.IsContainer())
	n := c1.Child("leaf2").(dom.Leaf)
	assert.NotNil(t, n)
	assert.Equal(t, "abc", n.Value())
}

func TestLoadNonExistingFile(t *testing.T) {
	d, err := JsonDoc("nonexisting.yaml", "xyz.yaml")
	assert.Error(t, err)
	assert.Nil(t, d)
}

func TestLoadMissingItem(t *testing.T) {
	f, err := os.CreateTemp("", "yt.*.yaml")
	defer func() {
		_ = os.Remove(f.Name())
	}()
	assert.Nil(t, err)
	_, err = f.Write([]byte(`
kind: ConfigMap
data: {}
`))
	assert.Nil(t, err)
	assert.Nil(t, f.Close())
	d, err := YamlDoc(f.Name(), "application.yaml")
	assert.Nil(t, err)
	assert.NotNil(t, d)
	d.Document().AddContainer("abc").AddValue("def", dom.LeafNode("123"))
	assert.Nil(t, d.Save())
}

func TestLoadInvalidDoc(t *testing.T) {
	f, err := os.CreateTemp("", "yt.*.yaml")
	defer func() {
		_ = os.Remove(f.Name())
	}()
	assert.Nil(t, err)
	_, err = f.Write([]byte(`
kind: ConfigMap
data:
  application.yaml: this is not a yaml
`))
	assert.Nil(t, err)
	assert.Nil(t, f.Close())
	d, err := YamlDoc(f.Name(), "application.yaml")
	assert.Error(t, err)
	assert.Nil(t, d)
}
