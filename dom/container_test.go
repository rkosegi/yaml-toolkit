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
	"golang.org/x/exp/slices"
	"os"
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
	assert.Equal(t, 3, len(fm))
	assert.NotNil(t, fm["level1.level2b"])
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

func TestFindValue(t *testing.T) {
	c, err := Builder().FromReader(strings.NewReader(`
leaf0: null
level1: 123
path.to.element1: Hi
path.to.element2: Hi
`), DefaultYamlDecoder)
	assert.NotNil(t, c)
	assert.Nil(t, err)
	assert.Nil(t, c.FindValue(456))
	assert.Equal(t, []string{"level1"}, c.FindValue(123))
	x := c.FindValue("Hi")
	assert.Equal(t, 2, len(x))
	assert.True(t, slices.Contains(x, "path.to.element1"))
	assert.True(t, slices.Contains(x, "path.to.element2"))
}
