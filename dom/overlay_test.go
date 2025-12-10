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
	"fmt"
	"os"
	"testing"

	"github.com/rkosegi/yaml-toolkit/path"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestPutAndGet(t *testing.T) {
	od := NewOverlayDocument()
	assert.Nil(t, od.Get("main", path.NewBuilder().Append(path.Simple("abc")).Build()))
	od.Set("main", path.NewBuilder(path.Simple("abc")).Build(), LeafNode("123"))
	assert.Nil(t, od.Get("main", path.NewBuilder().Append(path.Simple("")).Build()))
	assert.Equal(t, "123", od.Get("main", path.NewBuilder().Append(path.Simple("abc")).Build()).AsLeaf().Value())
	od.Set("main", path.NewBuilder(path.Simple("xyz"), path.Simple("efg")).Build(), LeafNode(42))

	p1 := path.NewBuilder().Append(path.Simple("xyz"), path.Simple("efg")).Build()
	assert.Equal(t, 42, od.Get("main", p1).AsLeaf().Value())
	assert.True(t, od.GetAny(path.NewBuilder().Append(path.Simple("xyz")).Build()).IsContainer())
	assert.Nil(t, od.GetAny(path.NewBuilder().Append(path.Simple("w")).Build()))
	assert.Nil(t, od.GetAny(path.ChildOf(p1, path.Simple("abc"))))
}

func TestLoad(t *testing.T) {
	od := NewOverlayDocument()
	var doc map[string]interface{}
	data, err := os.ReadFile("../testdata/doc1.yaml")
	assert.Nil(t, err)
	err = yaml.NewDecoder(bytes.NewReader(data)).Decode(&doc)
	assert.Nil(t, err)

	od.Add("layer-1", DecodeAnyToNode(doc).AsContainer())

	n := od.GetAny(path.NewBuilder().Append(path.Simple("level1")).Build())
	assert.True(t, n.IsContainer())
	c1 := n.AsContainer().Child("level2a")
	assert.True(t, c1.IsContainer())
	assert.Equal(t, 1, len(od.Layers()))

	var buf bytes.Buffer
	assert.NoError(t, EncodeToWriter(od.Merged(), defYamlCodec.Encoder(), &buf))
	x, err := decodeFromReader(&buf, DefaultYamlDecoder)
	assert.NoError(t, err)
	assert.True(t, x.IsContainer())
	assert.Equal(t, 3, x.AsContainer().
		Child("level1").AsContainer().
		Child("level2b").AsLeaf().Value())
}

func TestLoad2(t *testing.T) {
	od := NewOverlayDocument()
	var doc map[string]interface{}
	data, err := os.ReadFile("../testdata/doc2.yaml")
	assert.Nil(t, err)
	err = yaml.NewDecoder(bytes.NewReader(data)).Decode(&doc)
	assert.Nil(t, err)
	// d.Populate("layer-1", "", &doc)
	od.Add("layer-1", DecodeAnyToNode(doc).AsContainer())
	props := od.Merged().Flatten(SimplePathAsString)
	assert.Equal(t, 5, len(props))
}

func TestFirstValidListItem(t *testing.T) {
	assert.Equal(t, 456, firstValidListItem(1,
		ListNode(),
		ListNode(nilLeaf),
		ListNode(LeafNode(123), LeafNode(456))).AsLeaf().Value())
	assert.Equal(t, nilLeaf, firstValidListItem(2, ListNode()))
}

func TestHasValue(t *testing.T) {
	assert.True(t, hasValue(LeafNode(1)))
	assert.True(t, hasValue(ContainerNode()))
	assert.True(t, hasValue(ListNode()))
	assert.False(t, hasValue(nil))
	assert.False(t, hasValue(nilLeaf))
	assert.False(t, hasValue(LeafNode(nil)))
}

func TestOverlaySearch(t *testing.T) {
	d := NewOverlayDocument()
	d.Set("first", path.NewBuilder(path.Simple("root"), path.Simple("second")).Build(), LeafNode(1))
	res := d.Search(SearchEqual(1), SimplePathAsString)
	t.Log(res.String())
	assert.Equal(t, 1, len(res))
	assert.Equal(t, "first", res[0].Layer())
	assert.Equal(t, `["root","second"]`, res[0].Path())
	res = d.Search(SearchEqual(2), SimplePathAsString)
	assert.Equal(t, 0, len(res))
}

func TestOverlayLayers(t *testing.T) {
	od := NewOverlayDocument()
	od.Set("layer2", path.NewBuilder(path.Simple("root"), path.Simple("other")).Build(), LeafNode(5))
	m := od.Layers()
	assert.Equal(t, 1, len(m))
	assert.Equal(t, 5, m["layer2"].Children()["root"].AsContainer().Children()["other"].AsLeaf().Value())
}

func TestOverlayLayerNames(t *testing.T) {
	layers := 10
	od := NewOverlayDocument()
	for i := 0; i < layers; i++ {
		od.Set(fmt.Sprintf("layer-%d", i), path.NewBuilder(path.Simple("root")).Build(), LeafNode(i))
	}
	m := od.LayerNames()
	assert.Equal(t, layers, len(m))
	for i := 0; i < layers; i++ {
		assert.Equal(t, fmt.Sprintf("layer-%d", i), m[i])
	}
}

func TestOverlayAdd(t *testing.T) {
	d := NewOverlayDocument()
	d.Add("layerX", ContainerNode().AddContainer("sub"))
	c := ContainerNode()
	c.AddValue("sub1.sub2", LeafNode(123))
	d.Add("layer1", c)
	assert.Equal(t, 123, d.Layers()["layer1"].Children()["sub1.sub2"].AsLeaf().Value())
}

func TestOverlayWalk(t *testing.T) {
	d := NewOverlayDocument()
	d.Set("layer1", path.NewBuilder(
		path.Simple("root"),
		path.Simple("sub10"),
		path.Simple("sub21")).Build(), LeafNode("leaf1"))
	d.Set("layer2", path.NewBuilder(
		path.Simple("root"),
		path.Simple("sub11"),
		path.Simple("sub22")).Build(), ListNode(
		LeafNode(1),
		LeafNode(2),
		LeafNode(3),
	))
	var cnt int
	d.Walk(func(layer string, p path.Path, parent Node, node Node) bool {
		t.Logf("layer=%s, path=%v, parent=%v, node=%v", layer, p, parent, node)
		cnt++
		if node.IsLeaf() && node.AsLeaf().Value() == 1 {
			t.Logf("Hit false condition, terminating walk")
			return false
		}
		return true
	})
	assert.Equal(t, 7, cnt)
	cnt = 0
	d.Walk(func(layer string, p path.Path, parent Node, node Node) bool {
		t.Logf("layer=%s, path=%v, parent=%v, node=%v", layer, p, parent, node)
		cnt++
		return true
	})
	assert.Equal(t, 9, cnt)
}

func TestOverlaySet(t *testing.T) {
	od := NewOverlayDocument()
	od.Set("test", path.NewBuilder(
		path.Simple("root"),
		path.Simple("sub"),
		path.Simple("list"),
		path.Numeric(3)).Build(), LeafNode(1))

	assert.Equal(t, 1, len(od.Layers()))
	c := od.Layers()["test"]
	assert.Len(t, c.Children(), 1)
	assert.Equal(t, 1, c.Child("root").
		AsContainer().Child("sub").
		AsContainer().Child("list").
		AsList().Get(3).
		AsLeaf().Value())
}
