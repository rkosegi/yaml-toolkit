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
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/vmware-labs/yaml-jsonpath/pkg/yamlpath"
	"gopkg.in/yaml.v3"
	"os"
	"testing"
)

func TestPutAndLookup(t *testing.T) {
	d := NewOverlayDocument()
	assert.Nil(t, d.Lookup("main", "abc"))
	d.Put("main", "abc", LeafNode("123"))
	assert.Nil(t, d.Lookup("main", ""))
	assert.Equal(t, "123", d.Lookup("main", "abc").(Leaf).Value())
	d.Put("main", "xyz.efg", LeafNode(42))
	assert.Equal(t, 42, d.Lookup("main", "xyz.efg").(Leaf).Value())
	assert.True(t, d.LookupAny("xyz").IsContainer())
	assert.Nil(t, d.LookupAny("w"))
	assert.Nil(t, d.LookupAny("xyz.efg.abc"))
}

func TestSerialize(t *testing.T) {
	d := NewOverlayDocument()
	d.Put("", "key1", LeafNode("abc"))
	d.Put("layer-2", "key1.key2.key3", LeafNode("hello"))
	d.Put("layer-2", "key1.key11", LeafNode("ola!"))
	d.Put("layer-3", "key1.key2.key4", LeafNode(7))
	var buf bytes.Buffer
	assert.Nil(t, d.Serialize(&buf, DefaultNodeMappingFn, DefaultYamlEncoder))
	assert.True(t, buf.Len() > 0)
	buf.Reset()
	assert.Nil(t, d.Serialize(&buf, DefaultNodeMappingFn, DefaultYamlEncoder))
	assert.True(t, buf.Len() > 0)
}

func TestLoad(t *testing.T) {
	d := NewOverlayDocument()
	var doc map[string]interface{}
	data, err := os.ReadFile("../testdata/doc1.yaml")
	assert.Nil(t, err)
	err = yaml.NewDecoder(bytes.NewReader(data)).Decode(&doc)
	assert.Nil(t, err)

	d.Populate("layer-1", "key1.key11", &map[string]interface{}{
		"b": "xyz",
		"a": 12,
	})
	d.Populate("layer-1", "key1.key12", &doc)
	n := d.LookupAny("key1")
	assert.True(t, n.IsContainer())
	c1 := n.(Container).Child("key12")
	c1.IsContainer()
	assert.True(t, c1.IsContainer())
	var buf bytes.Buffer
	assert.Nil(t, d.Serialize(&buf, DefaultNodeMappingFn, DefaultYamlEncoder))

	var node yaml.Node
	err = yaml.NewDecoder(&buf).Decode(&node)
	assert.Nil(t, err)

	p, err := yamlpath.NewPath("$..key12")
	assert.Nil(t, err)
	nodes, err := p.Find(&node)
	assert.Nil(t, err)
	assert.NotNil(t, nodes)
	assert.Equal(t, "3", nodes[0].Content[1].Content[3].Value)
}
