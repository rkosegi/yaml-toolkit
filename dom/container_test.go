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
	"os"
	"slices"
	"strings"
	"testing"

	. "github.com/rkosegi/yaml-toolkit/path"
	"github.com/rkosegi/yaml-toolkit/query"
	"github.com/stretchr/testify/assert"
)

func TestBuilderFromYamlString(t *testing.T) {
	doc, err := DecodeReader(strings.NewReader(`
abc: 123
def: xyz
`), DefaultYamlDecoder)
	assert.Nil(t, err)
	assert.True(t, doc.IsContainer())
	assert.False(t, doc.IsLeaf())
	assert.False(t, doc.SameAs(nilLeaf))
	assert.Equal(t, "xyz", doc.AsContainer().Children()["def"].AsLeaf().Value())
	assert.Equal(t, 123, doc.AsContainer().Children()["abc"].AsLeaf().Value())
}

func TestBuilderFromInvalidJsonString(t *testing.T) {
	doc, err := DecodeReader(strings.NewReader(`This is not a json`), DefaultJsonDecoder)
	assert.NotNil(t, err)
	assert.Nil(t, doc)
}

func TestBuilderFromJsonString(t *testing.T) {
	doc, err := DecodeReader(strings.NewReader(`
{
	"def": "xyz",
	"abc": 123
}
`), DefaultJsonDecoder)
	assert.Nil(t, err)
	assert.True(t, doc.IsContainer())
	assert.Equal(t, "xyz", doc.AsContainer().Child("def").AsLeaf().Value())
	assert.Equal(t, float64(123), doc.AsContainer().Children()["abc"].AsLeaf().Value())
}

func TestBuildAndSerialize(t *testing.T) {
	builder := ContainerNode()
	builder.AddContainer("root").
		AddContainer("level1").
		AddContainer("level2").
		AddContainer("level3").
		AddValue("leaf1", LeafNode("Hello"))
	var buf bytes.Buffer
	assert.NoError(t, EncodeToWriter(builder, DefaultJsonEncoder, &buf))
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
	builder := ContainerNode()
	builder.AddContainer("root").
		AddContainer("level1").
		AddValue("leaf1", LeafNode("Hello"))
	builder.Remove("root")
	var buf bytes.Buffer
	assert.NoError(t, EncodeToWriter(builder, DefaultJsonEncoder, &buf))
	assert.Equal(t, "{}\n", buf.String())
}

func getTestDoc(t *testing.T, name string) ContainerBuilder {
	data, err := os.ReadFile("../testdata/" + name + ".yaml")
	assert.NoError(t, err)
	doc, err := DecodeReader(bytes.NewReader(data), DefaultYamlDecoder)
	assert.NoError(t, err)
	return doc.(ContainerBuilder)
}

func TestBuilderFromFile(t *testing.T) {
	doc := getTestDoc(t, "doc1")
	assert.True(t, doc.IsContainer())
	assert.Equal(t, "leaf1", doc.Child("level1").AsContainer().Child("level2a").AsContainer().Child("level3a").AsLeaf().Value())
	assert.Equal(t, 3, doc.Child("level1").AsContainer().Child("level2b").AsLeaf().Value())
}

func TestContainerAsAny(t *testing.T) {
	fm := getTestDoc(t, "doc1").AsAny().(map[string]interface{})
	assert.Equal(t, 1, len(fm))
	assert.NotNil(t, fm["level1"])
}

func TestFlatten(t *testing.T) {
	fm := getTestDoc(t, "doc1").Flatten()
	assert.Equal(t, 5, len(fm))
	assert.NotNil(t, fm["level1.level2b"])
}

func TestFlatten2(t *testing.T) {
	fm := getTestDoc(t, "doc2").Flatten()
	assert.Equal(t, 5, len(fm))
	assert.Equal(t, LeafNode(1), fm["root[2][0]"])
}

func TestFromMap(t *testing.T) {
	c := DecodeAnyToNode(map[string]interface{}{
		"test1.test2":  "abc",
		"test1.test22": 123,
	}).AsContainer()
	assert.Equal(t, "abc", c.Child("test1.test2").AsLeaf().Value())
}

func TestAddValueAt(t *testing.T) {
	c := ContainerNode()
	c.AddValueAt("test1.test2.test31", LeafNode("abc"))
	c.AddValueAt("test1.test2.test32", LeafNode(123))
	c.AddValueAt("test1.test2.test33", LeafNode(nil))
	c.AddValueAt("test1.test2.test34", ListNode(
		LeafNode("Hello"),
		ContainerNode(),
		ListNode()),
	)
	assert.Equal(t, "abc", c.Child("test1").
		AsContainer().Child("test2").
		AsContainer().Child("test31").
		AsLeaf().Value())
	var buff bytes.Buffer
	assert.NoError(t, EncodeToWriter(c, DefaultYamlEncoder, &buff))
	assert.Equal(t, `test1:
  test2:
    test31: abc
    test32: 123
    test33: null
    test34:
      - Hello
      - {}
      - []
`, buff.String())
}

func TestFromReaderNullLeaf(t *testing.T) {
	c, err := DecodeReader(strings.NewReader(`
leaf0: null
level1: 123
`), DefaultYamlDecoder)
	assert.NotNil(t, c)
	assert.Nil(t, err)
	assert.NotNil(t, c.AsContainer().Child("leaf0"))
	assert.Nil(t, c.AsContainer().Child("leaf0").AsLeaf().Value())
}

func TestSearch(t *testing.T) {
	c, err := DecodeReader(strings.NewReader(`
leaf0: null
level1: 123
to.element1: Hi
to.element2: Hi
`), DefaultYamlDecoder)
	assert.NotNil(t, c)
	assert.Nil(t, err)
	assert.Nil(t, c.AsContainer().Search(SearchEqual(456)))
	assert.Equal(t, []string{"level1"}, c.AsContainer().Search(SearchEqual(123)))
	x := c.AsContainer().Search(SearchEqual("Hi"))
	assert.Equal(t, 2, len(x))
	assert.True(t, slices.Contains(x, "to.element1"))
	assert.True(t, slices.Contains(x, "to.element2"))
}

func TestAddListAt(t *testing.T) {
	p0 := NewBuilder(Simple("root")).Build()
	p1 := NewBuilder(Simple("root"), Simple("sub")).Build()
	root := ContainerNode().AddContainer("root")
	root.Set(ChildOf(p0, Simple("list"), Numeric(0)), LeafNode(123))
	root.Set(ChildOf(p1, Simple("sub2"), Numeric(5)), LeafNode("abc"))
	root.Set(ChildOf(p1, Simple("sub2"), Numeric(4), Simple("sub3")), LeafNode(456))

	assert.Equal(t, 123, root.Child("root").
		AsContainer().Child("list").
		AsList().Get(0).AsLeaf().Value())
	assert.Equal(t, "abc", root.Child("root").
		AsContainer().Child("sub").
		AsContainer().Child("sub2").
		AsList().Get(5).AsLeaf().Value())
}

func TestCompact(t *testing.T) {
	c, err := DecodeReader(strings.NewReader(`
root:
  level2:
    leaf1: 123
    orphan: {}
`), DefaultYamlDecoder)
	assert.NotNil(t, c)
	assert.NoError(t, err)
	p := NewBuilder(Simple("root"), Simple("level2"), Simple("orphan")).Build()
	assert.NotNil(t, c.AsContainer().Get(p))
	c.AsContainer().Walk(CompactFn)
	assert.Nil(t, c.AsContainer().Get(p))
}

func TestWalk(t *testing.T) {
	c, err := DecodeReader(strings.NewReader(`
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
	c.AsContainer().Walk(func(p Path, parent Node, node Node) bool {
		if node.IsLeaf() && node.AsLeaf().Value() == 123 {
			return false
		}
		return true
	})

	x, err := DecodeReader(strings.NewReader(`
root:
  level1c:
    - value:
        - 1
    -
      - 1
      - 2
`), DefaultYamlDecoder)
	assert.NotNil(t, x)
	assert.NoError(t, err)
	cnt1, cnt2 := 0, 0
	x.AsContainer().Walk(func(p Path, parent Node, node Node) bool {
		if node.IsList() {
			cnt1++
		}
		if parent.IsList() {
			cnt2++
		}
		return true
	})
	assert.Equal(t, 3, cnt1)
	assert.Equal(t, 5, cnt2)
}

func TestContainerEquals(t *testing.T) {
	p1 := NewBuilder(Simple("a"), Simple("b"), Numeric(1)).Build()
	c := ContainerNode()
	c.Set(p1, LeafNode("123"))
	c2 := ContainerNode()
	c2.Set(p1, LeafNode("123"))

	assert.False(t, c.Equals(nil))
	assert.False(t, c.Equals(LeafNode(2)))
	assert.False(t, c.Equals(ContainerNode()))
	assert.True(t, c.Equals(c2))
}

func TestContainerClone(t *testing.T) {
	c := ContainerNode()
	c.Set(NewBuilder(Simple("a"), Simple("b"), Numeric(1)).Build(), LeafNode("123"))
	c.Set(NewBuilder(Simple("a"), Simple("x"), Simple("y")).Build(), LeafNode(123))
	c2 := c.Clone().AsContainer()
	assert.Equal(t, 123, c2.Child("a").
		AsContainer().Child("x").
		AsContainer().Child("y").
		AsLeaf().Value())
	assert.Equal(t, "123", c2.Child("a").
		AsContainer().Child("b").
		AsList().Get(1).AsLeaf().Value())
}

func TestMergeContainers(t *testing.T) {
	p := NewBuilder(Simple("l1")).Build()
	a := ContainerNode()
	c := ContainerNode()
	a.Set(ChildOf(p, Simple("l2c")), LeafNode(7))
	a.Set(ChildOf(p, Simple("l2d")), LeafNode("0987"))
	c.Set(ChildOf(p, Simple("l2a")), LeafNode("123"))
	c.Set(ChildOf(p, Simple("l2b")), LeafNode("abc"))
	d := c.Merge(a)
	assert.Equal(t, 4, len(d.Child("l1").AsContainer().Children()))
	assert.Equal(t, "abc", d.Child("l1").AsContainer().Child("l2b").AsLeaf().Value())
}

func TestContainerBuilderSeal(t *testing.T) {
	a := ContainerNode()
	c := a.Seal()
	_, isType := c.(ContainerBuilder)
	assert.False(t, isType)
}

func TestContainerBuilderSet(t *testing.T) {
	a := ContainerNode()
	p := NewBuilder(Simple("a"), Simple("b")).Build()
	a.Set(p, LeafNode(1))
	assert.Equal(t, 1, a.Child("a").AsContainer().Child("b").AsLeaf().Value())
}

func TestContainerBuilderDelete(t *testing.T) {
	a := getTestDoc(t, "doc1")
	p := NewBuilder(Simple("level1"), Simple("level2b")).Build()
	assert.Equal(t, 3, a.Get(p).AsLeaf().Value())
	a.Delete(p)
	assert.Nil(t, a.Get(p))
}

type asIsQueryImpl struct{}

func (a asIsQueryImpl) Select(data any) query.Result {
	return query.Result{data}
}

type noneQueryImpl struct{}

func (a noneQueryImpl) Select(_ any) query.Result { return nil }

func TestContainerQuery(t *testing.T) {
	a := getTestDoc(t, "doc1")
	r := a.Query(&asIsQueryImpl{})

	assert.NotNil(t, r)

	r = a.Query(&noneQueryImpl{})
	assert.Nil(t, r)
}
