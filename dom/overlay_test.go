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
	c := Builder().Container()
	c.AddContainer("test1").AddValue("test2", LeafNode("Hi"))
	c.AddValue("test3", LeafNode("no"))
	d.Put("", "key2", c)
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
	assert.True(t, c1.IsContainer())
	assert.Equal(t, 1, len(d.Layers()))

	var buf bytes.Buffer
	assert.Nil(t, d.Serialize(&buf, DefaultNodeMappingFn, DefaultYamlEncoder))
	var node yaml.Node
	err = yaml.NewDecoder(&buf).Decode(&node)
	assert.NoError(t, err)
	assert.Equal(t, "key1", node.Content[0].Content[0].Value)
	assert.Equal(t, "level2b", node.Content[0].Content[1].Content[3].Content[1].Content[4].Value)
}

func TestLoad2(t *testing.T) {
	d := NewOverlayDocument()
	var doc map[string]interface{}
	data, err := os.ReadFile("../testdata/doc2.yaml")
	assert.Nil(t, err)
	err = yaml.NewDecoder(bytes.NewReader(data)).Decode(&doc)
	assert.Nil(t, err)
	d.Populate("layer-1", "", &doc)
	props := d.Merged().Flatten()
	assert.Equal(t, 5, len(props))
}

func TestLoadLookupList(t *testing.T) {
	d := NewOverlayDocument()
	d.Put("", "key1.key2[0].key3", LeafNode("hello"))
	n := d.LookupAny("key1.key2[0].key3")
	assert.Equal(t, "hello", n.(Leaf).Value())
}

func TestCoalesce(t *testing.T) {
	assert.Equal(t, nilLeaf, coalesce(nilLeaf))
	assert.Equal(t, 123, coalesce(nilLeaf,
		LeafNode(123), nilLeaf).(Leaf).Value())
}

func TestFirstValidListItem(t *testing.T) {
	assert.Equal(t, 456, firstValidListItem(1,
		ListNode(),
		ListNode(nilLeaf),
		ListNode(LeafNode(123), LeafNode(456))).(Leaf).Value())
	assert.Equal(t, nilLeaf, firstValidListItem(2, ListNode()))
}

func TestHasValue(t *testing.T) {
	assert.True(t, hasValue(LeafNode(1)))
	assert.True(t, hasValue(Builder().Container()))
	assert.True(t, hasValue(ListNode()))
	assert.False(t, hasValue(nil))
	assert.False(t, hasValue(nilLeaf))
	assert.False(t, hasValue(LeafNode(nil)))
}

func TestOverlaySearch(t *testing.T) {
	d := NewOverlayDocument()
	d.Put("first", "root.second", LeafNode(1))
	res := d.Search(SearchEqual(1))
	t.Log(res.String())
	assert.Equal(t, 1, len(res))
	assert.Equal(t, "first", res[0].Layer())
	assert.Equal(t, "root.second", res[0].Path())
	res = d.Search(SearchEqual(2))
	assert.Equal(t, 0, len(res))
}
