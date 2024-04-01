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
	"os"
	"slices"
	"strings"
	"testing"
)

var (
	b = Builder()
)

func TestBuilderFromYamlString(t *testing.T) {
	doc, err := b.FromReader(strings.NewReader(`
abc: 123
def: xyz
`), DefaultYamlDecoder)
	assert.Nil(t, err)
	assert.True(t, doc.IsContainer())
	assert.False(t, doc.IsLeaf())
	assert.False(t, doc.SameAs(nilLeaf))
	assert.Equal(t, "xyz", doc.Children()["def"].(Leaf).Value())
	assert.Equal(t, 123, doc.Children()["abc"].(Leaf).Value())
}

func TestBuilderFromInvalidYamlString(t *testing.T) {
	doc, err := b.FromReader(strings.NewReader(`This is not a yaml`), DefaultYamlDecoder)
	assert.NotNil(t, err)
	assert.Nil(t, doc)
}

func TestBuilderFromJsonString(t *testing.T) {
	doc, err := b.FromReader(strings.NewReader(`
{
	"def": "xyz",
	"abc": 123
}
`), DefaultJsonDecoder)
	assert.Nil(t, err)
	assert.True(t, doc.IsContainer())
	assert.Equal(t, "xyz", doc.Children()["def"].(Leaf).Value())
	assert.Equal(t, float64(123), doc.Children()["abc"].(Leaf).Value())
}

func TestBuildAndSerialize(t *testing.T) {
	builder := b.Container()
	builder.AddContainer("root").
		AddContainer("level1").
		AddContainer("level2").
		AddContainer("level3").
		AddValue("leaf1", LeafNode("Hello"))
	var buf bytes.Buffer
	err := builder.Serialize(&buf, DefaultNodeMappingFn, DefaultJsonEncoder)
	assert.Nil(t, err)
	assert.Equal(t, `{
  "root": {
    "level1": {
      "level2": {
        "level3": {
          "leaf1": "Hello"
        }
      }
    }
  }
}
`, buf.String())
}

func TestRemove(t *testing.T) {
	builder := b.Container()
	builder.AddContainer("root").
		AddContainer("level1").
		AddValue("leaf1", LeafNode("Hello"))
	builder.Remove("root")
	var buf bytes.Buffer
	err := builder.Serialize(&buf, DefaultNodeMappingFn, DefaultJsonEncoder)
	assert.Nil(t, err)
	assert.Equal(t, "{}\n", buf.String())
}

func TestBuilderFromFile(t *testing.T) {
	data, err := os.ReadFile("../testdata/doc1.yaml")
	assert.Nil(t, err)
	doc, err := b.FromReader(bytes.NewReader(data), DefaultYamlDecoder)
	assert.Nil(t, err)
	assert.True(t, doc.IsContainer())
	assert.Equal(t, "leaf1", doc.Child("level1").(Container).Child("level2a").(Container).Child("level3a").(Leaf).Value())
	assert.Equal(t, 3, doc.Child("level1").(Container).Child("level2b").(Leaf).Value())
}

func TestLookup(t *testing.T) {
	data, err := os.ReadFile("../testdata/doc1.yaml")
	assert.Nil(t, err)
	doc, err := b.FromReader(bytes.NewReader(data), DefaultYamlDecoder)
	assert.Nil(t, err)

	assert.NotNil(t, doc.Lookup("level1"))
	assert.Nil(t, doc.Lookup("level1a"))
	assert.Nil(t, doc.Lookup(""))
	assert.Nil(t, doc.Lookup("level1.level2b.level3"))
	assert.Equal(t, "leaf1", doc.Lookup("level1.level2a.level3a").(Leaf).Value())
}

func TestFlatten(t *testing.T) {
	data, err := os.ReadFile("../testdata/doc1.yaml")
	assert.Nil(t, err)
	doc, err := b.FromReader(bytes.NewReader(data), DefaultYamlDecoder)
	assert.Nil(t, err)
	fm := doc.Flatten()
	assert.Equal(t, 5, len(fm))
	assert.NotNil(t, fm["level1.level2b"])
}

func TestFlatten2(t *testing.T) {
	data, err := os.ReadFile("../testdata/doc2.yaml")
	assert.Nil(t, err)
	doc, err := b.FromReader(bytes.NewReader(data), DefaultYamlDecoder)
	assert.Nil(t, err)
	fm := doc.Flatten()
	assert.Equal(t, 5, len(fm))
	assert.Equal(t, LeafNode(1), fm["root[2][0]"])
}

func TestFromMap(t *testing.T) {
	c := b.FromMap(map[string]interface{}{
		"test1.test2":  "abc",
		"test1.test22": 123,
	})
	assert.Equal(t, "abc", c.Lookup("test1.test2").(Leaf).Value())
}

func TestRemoveAt(t *testing.T) {
	c := b.FromMap(map[string]interface{}{
		"test1.test2":       "abc",
		"test1.test22":      123,
		"testA.testB.testC": "Hi",
	})
	c.RemoveAt("testA.non-existing.another")
	assert.NotNil(t, c.Lookup("test1.test22"))
	c.RemoveAt("test1.test22")
	assert.Nil(t, c.Lookup("test1.test22"))
}

func TestAddValueAt(t *testing.T) {
	c := b.Container()
	c.AddValueAt("test1.test2.test31", LeafNode("abc"))
	c.AddValueAt("test1.test2.test32", LeafNode(123))
	c.AddValueAt("test1.test2.test33", LeafNode(nil))
	c.AddValueAt("test1.test2.test34", ListNode(
		LeafNode("Hello"),
		b.Container(),
		ListNode()),
	)
	assert.Equal(t, "abc", c.Lookup("test1.test2.test31").(Leaf).Value())
	var buff bytes.Buffer
	err := c.Serialize(&buff, DefaultNodeMappingFn, DefaultYamlEncoder)
	assert.Nil(t, err)
}

func TestFromReaderNullLeaf(t *testing.T) {
	c, err := Builder().FromReader(strings.NewReader(`
leaf0: null
level1: 123
`), DefaultYamlDecoder)
	assert.NotNil(t, c)
	assert.Nil(t, err)
	assert.NotNil(t, c.Child("leaf0"))
	assert.Nil(t, c.Child("leaf0").(Leaf).Value())
}

func TestSearch(t *testing.T) {
	c, err := Builder().FromReader(strings.NewReader(`
leaf0: null
level1: 123
path.to.element1: Hi
path.to.element2: Hi
`), DefaultYamlDecoder)
	assert.NotNil(t, c)
	assert.Nil(t, err)
	assert.Nil(t, c.Search(SearchEqual(456)))
	assert.Equal(t, []string{"level1"}, c.Search(SearchEqual(123)))
	x := c.Search(SearchEqual("Hi"))
	assert.Equal(t, 2, len(x))
	assert.True(t, slices.Contains(x, "path.to.element1"))
	assert.True(t, slices.Contains(x, "path.to.element2"))
}

func TestLookupList(t *testing.T) {
	c, err := Builder().FromReader(strings.NewReader(`
root:
  list:
    - item1: abc
    - 123
  not-a-list:
    prop: 456
`), DefaultYamlDecoder)
	assert.NotNil(t, c)
	assert.Nil(t, err)
	assert.Equal(t, "abc", c.Lookup("root.list[0].item1").(Leaf).Value())
	assert.Equal(t, 123, c.Lookup("root.list[1]").(Leaf).Value())
	assert.Nil(t, c.Lookup("root.list[2]"))
	assert.Nil(t, c.Lookup("root.not-a-list[0]"))
	assert.Nil(t, c.Lookup("root.not-exists-at-all[0]"))
}

func TestAddListAt(t *testing.T) {
	root := b.Container().AddContainer("root")
	root.AddValueAt("root.list[0]", LeafNode(123))
	root.AddValueAt("root.sub.sub2[5]", LeafNode("abc"))
	root.AddValueAt("root.sub.sub2[4].sub3", LeafNode(456))

	assert.Equal(t, 123, root.Lookup("root.list[0]").(Leaf).Value())
	assert.Equal(t, "abc", root.Lookup("root.sub.sub2[5]").(Leaf).Value())
}

func TestFromAny(t *testing.T) {
	type x struct {
		A string
		B int
	}
	r := b.FromAny(&x{A: "abc", B: 456}).Flatten()
	assert.Equal(t, 2, len(r))
}

func TestCompact(t *testing.T) {
	c, err := Builder().FromReader(strings.NewReader(`
root:
  level2:
    leaf1: 123
    orphan: {}
`), DefaultYamlDecoder)
	assert.NotNil(t, c)
	assert.NoError(t, err)
	assert.NotNil(t, c.Children()["root"].(ContainerBuilder).Children()["level2"].(ContainerBuilder).Children()["orphan"])
	c.Walk(CompactFn)
	assert.Nil(t, c.Children()["root"].(ContainerBuilder).Children()["level2"].(ContainerBuilder).Children()["orphan"])

	c, err = Builder().FromReader(strings.NewReader(`
root:
  level2:
    orphan: {}
`), DefaultYamlDecoder)
	assert.NotNil(t, c)
	assert.NoError(t, err)
	assert.NotNil(t, c.Children()["root"].(ContainerBuilder).Children()["level2"].(ContainerBuilder).Children()["orphan"])
	c.Walk(CompactFn)
	assert.Nil(t, c.Children()["root"])
}

func TestWalk(t *testing.T) {
	c, err := Builder().FromReader(strings.NewReader(`
root:
  level1:
    level2a:
      leaf1: 123
    level2b:
      leaf1: 123
      leaf2: 123
      leaf3: 123
`), DefaultYamlDecoder)
	assert.NotNil(t, c)
	assert.NoError(t, err)
	c.Walk(func(path string, parent ContainerBuilder, node Node) bool {
		if node.IsLeaf() && node.(Leaf).Value() == 123 {
			return false
		}
		return true
	})
}

func TestContainerEquals(t *testing.T) {
	c := Builder().Container()
	c.AddValueAt("a.b[1]", LeafNode("123"))
	c2 := Builder().Container()
	c2.AddValueAt("a.b[1]", LeafNode("123"))

	assert.False(t, c.Equals(nil))
	assert.False(t, c.Equals(LeafNode(2)))
	assert.False(t, c.Equals(Builder().Container()))
	assert.True(t, c.Equals(c2))
}

func TestContainerClone(t *testing.T) {
	c := Builder().Container()
	c.AddValueAt("a.b[1]", LeafNode("123"))
	c.AddValueAt("a.x.y", LeafNode(123))
	c2 := c.Clone().(Container)
	assert.Equal(t, 123, c2.Lookup("a.x.y").(Leaf).Value())
	assert.Equal(t, "123", c2.Lookup("a.b[1]").(Leaf).Value())
}
