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
	"github.com/rkosegi/yaml-toolkit/dom"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func assertHasChange(t *testing.T, mod Modification, mods *[]Modification) {
	for _, m := range *mods {
		if m.Path == mod.Path && m.Type == mod.Type &&
			((m.Value != nil && mod.Value != nil && m.Value.Value() == mod.Value.Value()) ||
				(m.Value == nil && mod.Value == nil)) {
			// t.Logf("Found: %s", mod)
			return
		}
	}
	t.Fatalf("expected change not present %s", mod.String())
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
		Type:  ModChange,
		Path:  "leaf1",
		Value: dom.LeafNode(456),
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
		Type:  ModChange,
		Path:  "leaf0",
		Value: dom.LeafNode(1234),
	}, res)
	assertHasChange(t, Modification{
		Type:  ModChange,
		Path:  "level1.level2.leaf12",
		Value: dom.LeafNode(456),
	}, res)
}

func TestDiffReplace1(t *testing.T) {
	res := diffStrDocs(t, `
leaf0: 123
level1:
  level2:
    leaf12: abcd
`, `
leaf0: 1234
level1:
  level2: 123
`)
	assert.Equal(t, 3, len(*res))
	assertHasChange(t, Modification{
		Type: ModDelete,
		Path: "level1.level2",
	}, res)
	assertHasChange(t, Modification{
		Type:  ModAdd,
		Path:  "level1.level2",
		Value: dom.LeafNode(123),
	}, res)
	assertHasChange(t, Modification{
		Type:  ModChange,
		Path:  "leaf0",
		Value: dom.LeafNode(1234),
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
		Type:  ModChange,
		Path:  "leaf0",
		Value: dom.LeafNode(123),
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
