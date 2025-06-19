/*
Copyright 2025 Richard Kosegi

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

package fluent

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/rkosegi/yaml-toolkit/dom"
	"github.com/stretchr/testify/assert"
)

func TestMorpherAddWithFilter(t *testing.T) {
	d := dom.Builder().Container()
	d.AddValue("a", dom.LeafNode(1))
	d.AddValueAt("b", dom.LeafNode(2))

	data, err := os.ReadFile("../testdata/doc2.yaml")
	assert.NoError(t, err)
	doc, err := b.FromReader(bytes.NewReader(data), dom.DefaultYamlDecoder)
	assert.NoError(t, err)

	res := NewMorpher().AddWithFilter(doc, func(path string, leaf dom.Leaf) bool {
		t.Logf("path=%v, leaf=%v", path, leaf)
		return !strings.HasPrefix(path, "root[2]")
	}).Result()

	assert.Equal(t, "str leaf", res.Child("root").AsList().Items()[0].AsLeaf().Value())
	assert.Equal(t, 2, len(res.Child("root").AsList().Items()))
}

func TestMorpherMutate(t *testing.T) {
	res := NewMorpher().Set(dom.Builder().FromMap(map[string]interface{}{
		"A": 123,
		"B": "Hi",
	})).Mutate(func(d dom.ContainerBuilder) {
		d.AddValue("A", dom.LeafNode(1))
	}).Result()

	assert.Equal(t, 1, res.Child("A").AsLeaf().Value())
	assert.Equal(t, "Hi", res.Child("B").AsLeaf().Value())
}
