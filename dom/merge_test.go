/*
Copyright 2024 Richard Kosegi

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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMergeSimple(t *testing.T) {
	b1 := ContainerNode()
	b1.AddValueAt("root.list", ListNode(LeafNode(123), LeafNode(456)))
	b2 := ContainerNode()
	b2.AddValueAt("root.list[2]", LeafNode(789))
	m := &merger{}
	m.init()
	c := m.mergeContainers(b1, b2)
	assert.Equal(t, 3, len(c.Flatten()))
}

func TestMergeLists(t *testing.T) {
	m := &merger{}
	m.init()
	l := m.mergeListsMeld(
		ListNode(ListNode(LeafNode(123), LeafNode(456))).(ListBuilder),
		ListNode(ListNode()).(ListBuilder),
	)
	assert.Equal(t, 1, l.Size())
	assert.Equal(t, 2, l.Items()[0].(List).Size())
}

func TestMergeListsAppend(t *testing.T) {
	m := &merger{}
	m.init(ListsMergeAppend())
	l := m.mergeLists(
		ListNode(
			LeafNode(123),
			LeafNode(456),
		),
		ListNode(
			LeafNode(789),
		),
	)
	assert.Equal(t, 3, l.Size())
	assert.Equal(t, 123, l.Items()[0].AsLeaf().Value())
	assert.Equal(t, 456, l.Items()[1].AsLeaf().Value())
	assert.Equal(t, 789, l.Items()[2].AsLeaf().Value())
}

func TestMergeContainerFromTwoLists(t *testing.T) {
	c1 := ContainerNode()
	c1.AddValue("prop1", LeafNode(123))
	c2 := ContainerNode()
	c2.AddValue("prop2", LeafNode("abc"))
	m := &merger{}
	m.init()
	l := m.mergeListsMeld(ListNode(c1), ListNode(c2))
	assert.Equal(t, 1, l.Size())
}

func TestCoalesce(t *testing.T) {
	assert.Equal(t, nilLeaf, coalesce(nilLeaf))
	assert.Equal(t, 123, coalesce(nilLeaf,
		LeafNode(123), nilLeaf).AsLeaf().Value())
}

func TestMergeOverrideLeafValue(t *testing.T) {
	d1 := `
---
root:
  sub1:
    sub2:
      leaf: 1
`
	d2 := `
---
root:
  sub1:
    sub2:
      leaf: 2
`
	orig, err := b.FromReader(strings.NewReader(d1), DefaultYamlDecoder)
	assert.NoError(t, err)
	override, err := b.FromReader(strings.NewReader(d2), DefaultYamlDecoder)
	assert.NoError(t, err)
	result := orig.Merge(override)
	assert.Equal(t, 2, result.Lookup("root.sub1.sub2.leaf").AsLeaf().Value())
}
