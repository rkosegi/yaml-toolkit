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
	"testing"

	"github.com/rkosegi/yaml-toolkit/dom"
	"github.com/rkosegi/yaml-toolkit/path"
	"github.com/stretchr/testify/assert"
)

func TestMorpherCopyMerge(t *testing.T) {
	d := dom.Builder().Container()
	d.AddValue("a", dom.LeafNode(1))
	d.AddValueAt("b", dom.LeafNode(2))

	data, err := os.ReadFile("../testdata/doc2.yaml")
	assert.NoError(t, err)
	doc, err := b.FromReader(bytes.NewReader(data), dom.DefaultYamlDecoder)
	assert.NoError(t, err)

	res := NewMorpher().Copy(doc, CopyModeMerge(
		CopyParamFilterFunc(func(p path.Path, parent dom.Node, node dom.Node) bool {
			pc := p.Components()
			return len(pc) > 1 && pc[1].NumericValue() != 2
		}))).Result()

	assert.Equal(t, "str leaf", res.Child("root").AsList().Items()[0].AsLeaf().Value())
	assert.Equal(t, 2, len(res.Child("root").AsList().Items()))
}

func TestMorpherCopyMergeWithList(t *testing.T) {
	d1 := dom.ContainerNode()
	d1.AddValue("a", dom.LeafNode("AAA"))
	d1.AddValue("c", dom.ListNode(dom.LeafNode(2), dom.LeafNode(3)))

	d2 := dom.ContainerNode()
	d2.AddValue("b", dom.LeafNode("BBB"))
	d2.AddValue("c", dom.ListNode(dom.LeafNode(8), dom.LeafNode(9)))

	res := NewMorpher().
		Copy(d1, CopyModeMerge()).
		Copy(d2, CopyModeMerge(CopyParamMergeOptions(dom.ListsMergeAppend()))).
		Result()

	assert.Equal(t, 4, len(res.Child("c").AsList().Items()))
}

func TestMorpherCopyReplace(t *testing.T) {
	d1 := dom.ContainerNode()
	d1.AddValue("a", dom.LeafNode(1))
	d2 := dom.ContainerNode()
	d2.AddValue("b", dom.LeafNode(2))
	res := NewMorpher().
		Copy(d1, CopyModeReplace()).
		Copy(d2, CopyModeReplace(CopyParamFilterFunc(func(p path.Path, parent dom.Node, node dom.Node) bool {
			return p.Last().Value() != "b"
		}))).
		Result()
	assert.Equal(t, 0, len(res.Children()))
}

func TestMorpherMutate(t *testing.T) {
	res := NewMorpher().Copy(dom.Builder().FromMap(map[string]interface{}{
		"A": 123,
		"B": "Hi",
	}), CopyModeReplace()).Mutate(func(d dom.ContainerBuilder) {
		d.AddValue("A", dom.LeafNode(1))
	}).Result()

	assert.Equal(t, 1, res.Child("A").AsLeaf().Value())
	assert.Equal(t, "Hi", res.Child("B").AsLeaf().Value())
}
