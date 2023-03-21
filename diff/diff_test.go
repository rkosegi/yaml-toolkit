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

package diff

import (
	"bytes"
	"github.com/google/go-cmp/cmp"
	"github.com/rkosegi/yaml-toolkit/dom"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func leavesEqual(l1, l2 dom.Leaf) bool {
	if l1 == nil && l2 == nil {
		return true
	}
	if l1 != nil && l2 != nil {
		return cmp.Equal(l1.Value(), l2.Value())
	}
	return false
}

func assertHasChange(t *testing.T, mod Modification, mods *[]Modification) {
	for _, m := range *mods {
		if cmp.Equal(m, mod, cmp.Comparer(func(m1, m2 Modification) bool {
			return cmp.Equal(m.Type, mod.Type) && cmp.Equal(m.Path, mod.Path) &&
				leavesEqual(m.Value, mod.Value) && leavesEqual(m.OldValue, mod.OldValue)
		})) {
			return
		}
	}
	t.Fatalf("expected change not present %s, all changes: %v", mod.String(), mods)
}

func diffStrDocs(t *testing.T, doc1, doc2 string) *[]Modification {
	cleft, err := dom.Builder().FromReader(strings.NewReader(doc1), dom.DefaultYamlDecoder)
	if err != nil {
		t.Fatal(err)
	}
	cright, err := dom.Builder().FromReader(strings.NewReader(doc2), dom.DefaultYamlDecoder)
	if err != nil {
		t.Fatal(err)
	}
	return Diff(cleft, cright)
}

func TestDiffSimple1(t *testing.T) {
	c1 := dom.Builder().Container()
	c1.AddValue("leaf1", dom.LeafNode("abc"))
	c2 := dom.Builder().Container()
	c2.AddValue("leaf1", dom.LeafNode(456))
	res := Diff(c1, c2)
	assert.Equal(t, 1, len(*res))
	assertHasChange(t, Modification{
		Type:     ModChange,
		Path:     "leaf1",
		Value:    dom.LeafNode(456),
		OldValue: dom.LeafNode("abc"),
	}, res)
}

func TestDiffSimple2(t *testing.T) {
	res := diffStrDocs(t, `
leaf0: 123
level1:
  level2:
    leaf12: abcd`, `
leaf0: 1234
level1:
  level2:
    leaf12: 456`)
	assert.Equal(t, 2, len(*res))
	assertHasChange(t, Modification{
		Type:     ModChange,
		Path:     "leaf0",
		Value:    dom.LeafNode(1234),
		OldValue: dom.LeafNode(123),
	}, res)
	assertHasChange(t, Modification{
		Type:     ModChange,
		Path:     "level1.level2.leaf12",
		Value:    dom.LeafNode(456),
		OldValue: dom.LeafNode("abcd"),
	}, res)
}

func TestDiffReplace1(t *testing.T) {
	res := diffStrDocs(t, `
leaf0: 123
level1:
  level2:
    leaf12: abcd
leaf2: Hi
`, `
leaf0: 1234
level1:
  level2: 123
`)
	assert.Equal(t, 4, len(*res))
	assertHasChange(t, Modification{
		Type: ModDelete,
		Path: "level1.level2",
	}, res)
	assertHasChange(t, Modification{
		Type:  ModAdd,
		Path:  "leaf2",
		Value: dom.LeafNode("Hi"),
	}, res)
	assertHasChange(t, Modification{
		Type:  ModAdd,
		Path:  "level1.level2",
		Value: dom.LeafNode(123),
	}, res)
	assertHasChange(t, Modification{
		Type:     ModChange,
		Path:     "leaf0",
		Value:    dom.LeafNode(1234),
		OldValue: dom.LeafNode(123),
	}, res)
}

func TestDiffReplaceAndApply(t *testing.T) {
	cleft, err := dom.Builder().FromReader(strings.NewReader(`
leaf0: 1234
level1:
  level2: 123
another:
  container:
    leaf13: Hi
`), dom.DefaultYamlDecoder)
	if err != nil {
		t.Fatal(err)
	}
	cright, err := dom.Builder().FromReader(strings.NewReader(`leaf0: 123
leaf1: 456
level1:
  level2:
    leaf12: abcd`), dom.DefaultYamlDecoder)
	if err != nil {
		t.Fatal(err)
	}
	res := Diff(cleft, cright)

	assert.Equal(t, 5, len(*res))

	assertHasChange(t, Modification{
		Type:     ModChange,
		Path:     "leaf0",
		Value:    dom.LeafNode(123),
		OldValue: dom.LeafNode(1234),
	}, res)
	assertHasChange(t, Modification{
		Type: ModDelete,
		Path: "level1.level2",
	}, res)
	assertHasChange(t, Modification{
		Type:  ModAdd,
		Path:  "level1.level2.leaf12",
		Value: dom.LeafNode("abcd"),
	}, res)
	assertHasChange(t, Modification{
		Type:  ModAdd,
		Path:  "another.container.leaf13",
		Value: dom.LeafNode("Hi"),
	}, res)
	assertHasChange(t, Modification{
		Type: ModDelete,
		Path: "leaf1",
	}, res)
	Apply(cright, *res)
	var buf bytes.Buffer
	err = cright.Serialize(&buf, dom.DefaultNodeMappingFn, dom.DefaultYamlEncoder)
	assert.Nil(t, err)
	assert.Equal(t, `another:
  container:
    leaf13: Hi
leaf0: 123
level1:
  level2:
    leaf12: abcd
`, buf.String())
}

func TestModString(t *testing.T) {
	mod := Modification{}
	assert.True(t, len(mod.String()) > 0)
}
