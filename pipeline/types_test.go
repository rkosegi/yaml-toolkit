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

package pipeline

import (
	"testing"

	"github.com/rkosegi/yaml-toolkit/dom"
	te "github.com/rkosegi/yaml-toolkit/pipeline/template_engine"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestAnyVal(t *testing.T) {
	type testStruct struct {
		V *AnyVal `yaml:"v,omitempty"`
	}
	var (
		err error
		ts  testStruct
		// n  *yaml.Node
	)
	d1 := `v: 123`
	err = yaml.Unmarshal([]byte(d1), &ts)
	assert.NoError(t, err)
	assert.Equal(t, "123", ts.V.Value().(dom.Leaf).Value())
	d2 := `
v:
  a: 123
  b: XYZ
`
	err = yaml.Unmarshal([]byte(d2), &ts)
	assert.NoError(t, err)
	assert.Equal(t, "123", ts.V.Value().(dom.Container).Child("a").(dom.Leaf).Value())
}

func TestValOrRef(t *testing.T) {
	type testStruct struct {
		V *ValOrRef `yaml:"v,omitempty"`
	}
	var (
		ts  testStruct
		err error
	)
	d1 := `
v: aaa
`
	err = yaml.Unmarshal([]byte(d1), &ts)
	assert.NoError(t, err)
	assert.Equal(t, "aaa", ts.V.Val)

	d2 := `
v:
  ref: a.b.c
`
	d := b.Container()
	d.AddValueAt("a.b.c", dom.LeafNode("X"))
	ts.V.Val = ""
	err = yaml.Unmarshal([]byte(d2), &ts)
	assert.NoError(t, err)
	assert.Equal(t, "X", ts.V.Resolve(newMockActBuilder().data(d).build()))

	d3 := `
v:
  not-a-ref: 8798
`
	ts.V.Val = ""
	err = yaml.Unmarshal([]byte(d3), &ts)
	assert.Error(t, err)

	d4 := `
v: []`
	ts.V.Val = ""
	err = yaml.Unmarshal([]byte(d4), &ts)
	assert.Error(t, err)

	d5 := `
v:
  ref: a.b.c
`
	d = b.Container()
	d.AddValueAt("a.b.c", dom.ListNode(dom.LeafNode("X")))
	ts.V.Val = ""
	err = yaml.Unmarshal([]byte(d5), &ts)
	assert.NoError(t, err)
	assert.Equal(t, "", ts.V.Resolve(newMockActBuilder().data(d).build()))
}

func TestValOrRefString(t *testing.T) {
	assert.Equal(t, "[Val=A]", (&ValOrRef{Val: "A"}).String())
	assert.Equal(t, "[Ref=a.b]", (&ValOrRef{Ref: "a.b"}).String())
	assert.Equal(t, "[Ref=x.y,Val=X]", (&ValOrRef{Val: "X", Ref: "x.y"}).String())
}

func TestStrKeyValuesAsAnyValuesMap(t *testing.T) {
	in := StrKeysStrValues{
		"A": "abc",
	}
	out := in.AsAnyValuesMap()
	assert.Equal(t, "abc", out["A"])
}

func TestStrKeyValuesRenderValues(t *testing.T) {
	in := StrKeysStrValues{
		"A": `{{ printf "%s %s" .X .Y }}`,
	}
	teng := te.DefaultTemplateEngine()
	out := in.RenderValues(teng, StrKeysAnyValues{
		"X": "Hello",
		"Y": "World",
	})
	assert.Equal(t, "Hello World", out["A"])
}
