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
	assert.Equal(t, "{\"root\":{\"level1\":{\"level2\":{\"level3\":{\"leaf1\":\"Hello\"}}}}}\n", buf.String())
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
