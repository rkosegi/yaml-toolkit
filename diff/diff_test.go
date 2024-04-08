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
	"gopkg.in/yaml.v3"
	"strings"
	"testing"
)

func leavesEqual(l1, l2 interface{}) bool {
	if l1 == nil && l2 == nil {
		return true
	}
	if l1 != nil && l2 != nil {
		return cmp.Equal(l1, l2)
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
	data, _ := yaml.Marshal(mods)
	t.Fatalf("expected change not present %s, all changes:\n%v", mod.String(), string(data))
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
		Value:    456,
		OldValue: "abc",
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
		Value:    1234,
		OldValue: 123,
	}, res)
	assertHasChange(t, Modification{
		Type:     ModChange,
		Path:     "level1.level2.leaf12",
		Value:    456,
		OldValue: "abcd",
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
		Value: "Hi",
	}, res)
	assertHasChange(t, Modification{
		Type:  ModAdd,
		Path:  "level1.level2",
		Value: 123,
	}, res)
	assertHasChange(t, Modification{
		Type:     ModChange,
		Path:     "leaf0",
		Value:    1234,
		OldValue: 123,
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
		Value:    123,
		OldValue: 1234,
	}, res)
	assertHasChange(t, Modification{
		Type: ModDelete,
		Path: "level1.level2",
	}, res)
	assertHasChange(t, Modification{
		Type:  ModAdd,
		Path:  "level1.level2.leaf12",
		Value: "abcd",
	}, res)
	assertHasChange(t, Modification{
		Type:  ModAdd,
		Path:  "another.container.leaf13",
		Value: "Hi",
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

func TestDiffNotSimilar(t *testing.T) {
	res := diffStrDocs(t, `
list1:
  - item1
  - item2`, `
leaf0: 1234
level1:
  level2:
    leaf12: 456`)
	assert.Equal(t, 4, len(*res))
	assertHasChange(t, Modification{
		Type:  ModAdd,
		Path:  "list1[0]",
		Value: "item1",
	}, res)
	assertHasChange(t, Modification{
		Type:  ModAdd,
		Path:  "list1[1]",
		Value: "item2",
	}, res)
}

func TestDiffList(t *testing.T) {
	res := diffStrDocs(t, `
root: 123
list1:
  - item1
  - item_obj1:
      sub:
        leaf1: 789
        sublist:
          - abcd
          - efgh:
              - 123
  - item2`, `
root:
  - [ 4 ]
  - 2
list1:
  - 123
  - 456`)
	assert.Equal(t, 9, len(*res))
	assertHasChange(t, Modification{
		Type:  ModDelete,
		Path:  "list1",
		Value: nil,
	}, res)
	assertHasChange(t, Modification{
		Type:  ModAdd,
		Path:  "list1[0]",
		Value: "item1",
	}, res)
	assertHasChange(t, Modification{
		Type:  ModAdd,
		Path:  "list1[1].item_obj1.sub.leaf1",
		Value: 789,
	}, res)
	assertHasChange(t, Modification{
		Type:  ModAdd,
		Path:  "list1[2]",
		Value: "item2",
	}, res)
}

func TestDiffInnerList(t *testing.T) {
	res := diffStrDocs(t, `
root:
  list1:
    - item1`, `
root:
  list1:
    - item1`)
	assert.Equal(t, 0, len(*res))
}
