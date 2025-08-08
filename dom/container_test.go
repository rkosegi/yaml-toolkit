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

	"github.com/rkosegi/yaml-toolkit/path"
	"github.com/rkosegi/yaml-toolkit/query"
	"github.com/stretchr/testify/assert"
)

func TestContainerEncodeDecode(t *testing.T) {
	x := getTestDoc(t, "doc1")
	assert.Equal(t, x.AsAny(), DefaultNodeEncoderFn(x))
}

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

func TestLookup(t *testing.T) {
	doc := getTestDoc(t, "doc1")
	assert.NotNil(t, doc.Lookup("level1"))
	assert.Nil(t, doc.Lookup("level1a"))
	assert.Nil(t, doc.Lookup(""))
	assert.Nil(t, doc.Lookup("level1.level2b.level3"))
	assert.Equal(t, "leaf1", doc.Lookup("level1.level2a.level3a").AsLeaf().Value())
}

func TestContainerAsMap(t *testing.T) {
	fm := getTestDoc(t, "doc1").AsMap()
	assert.Equal(t, 1, len(fm))
	assert.NotNil(t, fm["level1"])
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

func TestRemoveAt(t *testing.T) {
	c := DecodeAnyToNode(map[string]interface{}{
		"test2":  "abc",
		"test22": 123,
		"testC":  "Hi",
	}).(ContainerBuilder)
	c.RemoveAt("non-existing.another")
	assert.NotNil(t, c.Child("test22"))
	c.RemoveAt("test22")
	assert.Nil(t, c.Child("test22"))
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
	assert.Equal(t, "abc", c.Lookup("test1.test2.test31").AsLeaf().Value())
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
path.to.element1: Hi
path.to.element2: Hi
`), DefaultYamlDecoder)
	assert.NotNil(t, c)
	assert.Nil(t, err)
	assert.Nil(t, c.AsContainer().Search(SearchEqual(456)))
	assert.Equal(t, []string{"level1"}, c.AsContainer().Search(SearchEqual(123)))
	x := c.AsContainer().Search(SearchEqual("Hi"))
	assert.Equal(t, 2, len(x))
	assert.True(t, slices.Contains(x, "path.to.element1"))
	assert.True(t, slices.Contains(x, "path.to.element2"))
}

func TestLookupList(t *testing.T) {
	c, err := DecodeReader(strings.NewReader(`
root:
  list:
    - item1: abc
    - 123
  not-a-list:
    prop: 456
`), DefaultYamlDecoder)
	assert.NotNil(t, c)
	assert.Nil(t, err)
	assert.Equal(t, "abc", c.AsContainer().Lookup("root.list[0].item1").AsLeaf().Value())
	assert.Equal(t, 123, c.AsContainer().Lookup("root.list[1]").AsLeaf().Value())
	assert.Nil(t, c.AsContainer().Lookup("root.list[2]"))
	assert.Nil(t, c.AsContainer().Lookup("root.not-a-list[0]"))
	assert.Nil(t, c.AsContainer().Lookup("root.not-exists-at-all[0]"))
}

func TestAddListAt(t *testing.T) {
	root := ContainerNode().AddContainer("root")
	root.AddValueAt("root.list[0]", LeafNode(123))
	root.AddValueAt("root.sub.sub2[5]", LeafNode("abc"))
	root.AddValueAt("root.sub.sub2[4].sub3", LeafNode(456))

	assert.Equal(t, 123, root.Lookup("root.list[0]").AsLeaf().Value())
	assert.Equal(t, "abc", root.Lookup("root.sub.sub2[5]").AsLeaf().Value())
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
	p := path.NewBuilder().
		Append(path.Simple("root")).
		Append(path.Simple("level2")).
		Append(path.Simple("orphan")).
		Build()
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
	c.AsContainer().Walk(func(p path.Path, parent Node, node Node) bool {
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
	x.AsContainer().Walk(func(p path.Path, parent Node, node Node) bool {
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
	c := ContainerNode()
	c.AddValueAt("a.b[1]", LeafNode("123"))
	c2 := ContainerNode()
	c2.AddValueAt("a.b[1]", LeafNode("123"))

	assert.False(t, c.Equals(nil))
	assert.False(t, c.Equals(LeafNode(2)))
	assert.False(t, c.Equals(ContainerNode()))
	assert.True(t, c.Equals(c2))
}

func TestContainerClone(t *testing.T) {
	c := ContainerNode()
	c.AddValueAt("a.b[1]", LeafNode("123"))
	c.AddValueAt("a.x.y", LeafNode(123))
	c2 := c.Clone().AsContainer()
	assert.Equal(t, 123, c2.Lookup("a.x.y").AsLeaf().Value())
	assert.Equal(t, "123", c2.Lookup("a.b[1]").AsLeaf().Value())
}

func TestMergeContainers(t *testing.T) {
	a := ContainerNode()
	c := ContainerNode()
	a.AddValueAt("l1.l2c", LeafNode(7))
	a.AddValueAt("l1.l2d", LeafNode("0987"))
	c.AddValueAt("l1.l2a", LeafNode("123"))
	c.AddValueAt("l1.l2b", LeafNode("abc"))
	d := c.Merge(a)
	assert.Equal(t, 4, len(d.Lookup("l1").AsContainer().Children()))
	assert.Equal(t, "abc", d.Lookup("l1.l2b").AsLeaf().Value())
}

func TestContainerBuilderSeal(t *testing.T) {
	a := ContainerNode()
	c := a.Seal()
	_, isType := c.(ContainerBuilder)
	assert.False(t, isType)
}

func TestContainerBuilderSet(t *testing.T) {
	a := ContainerNode()
	p := path.NewBuilder().Append(path.Simple("a")).Append(path.Simple("b")).Build()
	a.Set(p, LeafNode(1))
	assert.Equal(t, 1, a.Child("a").AsContainer().Child("b").AsLeaf().Value())
}

func TestContainerBuilderDelete(t *testing.T) {
	a := getTestDoc(t, "doc1")
	p := path.NewBuilder().
		Append(path.Simple("level1")).
		Append(path.Simple("level2b")).
		Build()
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
