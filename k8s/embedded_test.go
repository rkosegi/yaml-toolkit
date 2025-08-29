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
	"errors"
	"io"
	"os"
	"testing"

	"github.com/rkosegi/yaml-toolkit/dom"
	"github.com/stretchr/testify/assert"
)

var (
	anyErr = errors.New("any error")
)

const tempFilePattern = "yt.*.yaml"

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
	f, err := os.CreateTemp("", tempFilePattern)
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
	f, err := os.CreateTemp("", tempFilePattern)
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

func TestBuildInvalid(t *testing.T) {
	d, err := NewBuilder().Open()
	assert.Error(t, err)
	assert.Nil(t, d)
	d, err = NewBuilder().Decoder(DecodeEmbeddedProps()).Open()
	assert.Error(t, err)
	assert.Nil(t, d)
}

func TestLoadEmbeddedProps(t *testing.T) {
	d, err := Properties("../testdata/secret3.yaml")
	assert.Nil(t, err)
	assert.NotNil(t, d)
	assert.Equal(t, "123", d.Document().Child("prop2").
		AsContainer().Child("level2").
		AsContainer().Child("level3").(dom.Leaf).Value())

	f, err := os.CreateTemp("", tempFilePattern)
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
	d, err = Properties(f.Name())
	assert.Nil(t, err)
	assert.NotNil(t, d)
	d.Document().AddContainer("abc").AddValue("def", dom.LeafNode("123"))
	assert.Nil(t, d.Save())

}

func TestBuildCreate(t *testing.T) {
	f, err := os.CreateTemp("", tempFilePattern)
	defer func() {
		_ = os.Remove(f.Name())
	}()
	assert.Nil(t, err)
	doc, err := NewBuilder().Manifest(f.Name()).
		Encoder(EncodeEmbeddedProps()).
		Decoder(DecodeEmbeddedProps()).
		Create("Secret", "secret1", WithNamespace("something"))
	assert.NoError(t, err)
	doc.Document().AddValue("prop1", dom.LeafNode(123))
	err = doc.Save()
	assert.NoError(t, err)
	assert.NotNil(t, doc)
}

func TestFailBeforeSave(t *testing.T) {
	f, err := os.CreateTemp("", tempFilePattern)
	defer func() {
		_ = os.Remove(f.Name())
	}()
	assert.NoError(t, err)
	doc, err := NewBuilder().Manifest(f.Name()).
		Encoder(EncodeEmbeddedDoc("embedded.txt", func(w io.Writer, v interface{}) error {
			return anyErr
		})).
		Decoder(DecodeEmbeddedProps()).
		Create("ConfigMap", "temp")
	assert.NoError(t, err)
	assert.NotNil(t, doc)
	assert.Error(t, doc.Save())
}

func TestDecodeEmbeddedDoc(t *testing.T) {
	var (
		mf  Manifest
		err error
		cb  dom.ContainerBuilder
	)
	mf, err = ManifestFromBytes([]byte(`kind: ConfigMap
data:
  item: not a json`))
	assert.NoError(t, err)
	cb, err = DecodeEmbeddedDoc("item", dom.DefaultJsonDecoder)(mf)
	assert.Error(t, err)
	assert.Nil(t, cb)
}

func TestFailOnSave(t *testing.T) {
	f, err := os.CreateTemp("", tempFilePattern)
	defer func() {
		_ = os.Remove(f.Name())
	}()
	assert.NoError(t, err)
	doc, err := NewBuilder().Manifest(f.Name()).
		Encoder(EncodeEmbeddedProps()).
		Decoder(DecodeEmbeddedProps()).
		Create("ConfigMap", "temp")
	assert.NoError(t, err)
	assert.NoError(t, os.Chmod(f.Name(), 0o400))
	assert.NotNil(t, doc)
	assert.Error(t, doc.Save())
}

func TestFailOnCreate(t *testing.T) {
	f, err := os.CreateTemp("", tempFilePattern)
	defer func() {
		_ = os.Remove(f.Name())
	}()
	assert.NoError(t, f.Close())
	// this will cause permission denied during document creation
	assert.NoError(t, os.Chmod(f.Name(), 0o400))
	assert.NoError(t, err)
	doc, err := NewBuilder().Manifest(f.Name()).
		Encoder(EncodeEmbeddedProps()).
		Decoder(DecodeEmbeddedProps()).
		Create("ConfigMap", "temp")
	assert.Nil(t, doc)
	assert.Error(t, err)
}
