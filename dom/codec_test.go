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

	"github.com/rkosegi/yaml-toolkit/common"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestDecodeYamlNode(t *testing.T) {
	doc1 := `
root:
  list1:
    - item1: abc
      prop2: 123
      prop3: 3.3
      prop4: true
      prop5: false
`
	var (
		dn  Node
		n   yaml.Node
		err error
	)
	err = yaml.NewDecoder(strings.NewReader(doc1)).Decode(&n)
	assert.NoError(t, err)
	dn = YamlNodeDecoder()(&n)
	assert.NotNil(t, dn)
	l := dn.AsContainer().
		Child("root").AsContainer().
		Child("list1").AsList()
	assert.Equal(t, 1, l.Size())

	t.Run("parse int", func(t *testing.T) {
		assert.Equal(t, 123, l.Get(0).AsContainer().
			Child("prop2").AsLeaf().Value())
	})

	t.Run("parse float", func(t *testing.T) {
		assert.Equal(t, 3.3, l.Get(0).AsContainer().
			Child("prop3").AsLeaf().Value())
	})
	t.Run("parse bool", func(t *testing.T) {
		assert.Equal(t, true, l.Get(0).AsContainer().
			Child("prop4").AsLeaf().Value())
		assert.Equal(t, false, l.Get(0).AsContainer().
			Child("prop5").AsLeaf().Value())
	})

	assert.Nil(t, decodeYamlNode(&yaml.Node{
		Kind: yaml.AliasNode,
	}))
}

func getYamlNodeFromTextForTest(t *testing.T, text string) *yaml.Node {
	var out yaml.Node
	assert.NoError(t, yaml.NewDecoder(strings.NewReader(text)).Decode(&out))
	return &out
}

func TestDecodeYamlNodeScalar(t *testing.T) {
	dec := YamlNodeDecoder()
	t.Run("decode number should yield int", func(t *testing.T) {
		o := dec(getYamlNodeFromTextForTest(t, "1"))
		assert.NotNil(t, o)
		assert.Equal(t, 1, o.AsLeaf().Value())
	})
	t.Run("decode string should yield string", func(t *testing.T) {
		o := dec(getYamlNodeFromTextForTest(t, "abcd"))
		assert.NotNil(t, o)
		assert.Equal(t, "abcd", o.AsLeaf().Value())
	})
}

func TestDecodeAnyToNode(t *testing.T) {
	type testData1 struct {
		A int
		B struct {
			B1 string
			B2 float64
			b3 int
		}
		c chan struct{}
		D []int
		M map[string]int
	}
	x1 := &testData1{
		A: 13,
		B: struct {
			B1 string
			B2 float64
			b3 int
		}{
			B1: "abc",
			B2: 3.5,
		},
		D: []int{4, 9, 3},
		M: map[string]int{
			"A": 1,
		},
	}
	y1Src := `
A: 13
B:
  B1: abc
  B2: 3.5
D: [4,9,3]
M:
  A: 1`

	y1, err := DecodeReader(strings.NewReader(y1Src), DefaultYamlDecoder)
	assert.NoError(t, err)
	res := DecodeAnyToNode(x1).(ContainerBuilder)
	assert.NotNil(t, res)
	assert.Equal(t, 13, res.Child("A").AsLeaf().Value())
	assert.Equal(t, y1.(ContainerBuilder).Seal(), res.Seal())

	// chan is not considered during decode
	assert.Nil(t, DecodeAnyToNode(make(chan struct{})))

	assert.Equal(t, nilLeaf, DecodeAnyToNode(nil))
}

func TestDecodeFromReader(t *testing.T) {
	var (
		out Node
		err error
	)
	t.Run("Decode a valid YAML container", func(t *testing.T) {
		out, err = DecodeReader(strings.NewReader(`
root:
  sub:
    sub2: 123
`), DefaultYamlDecoder)
	})
	assert.NoError(t, err)
	assert.True(t, out.IsContainer())
	assert.Equal(t, 123, out.AsContainer().Child("root").
		AsContainer().Child("sub").
		AsContainer().Child("sub2").
		AsLeaf().Value())

	t.Run("Decode leaf value as YAML", func(t *testing.T) {
		out, err = DecodeReader(strings.NewReader(`just a leaf`), DefaultYamlDecoder)
		assert.NoError(t, err)
		assert.Equal(t, "just a leaf", out.AsLeaf().Value())
	})
	t.Run("Decode invalid JSON", func(t *testing.T) {
		out, err = DecodeReader(strings.NewReader(`something`), DefaultJsonDecoder)
		assert.Error(t, err)
		assert.Nil(t, out)
	})
	t.Run("failure from io.Reader should fail decoding", func(t *testing.T) {
		_, err = DecodeReader(common.FailingReader(), DefaultYamlDecoder)
		assert.Error(t, err)
	})
	t.Run("expect panic from MustDecodeReader", func(t *testing.T) {
		defer func() {
			recover()
		}()
		MustDecodeReader(common.FailingReader(), DefaultYamlDecoder)
		t.Fail()
	})
	t.Run("calling MustDecodeReader with valid json should pass", func(t *testing.T) {
		out = MustDecodeReader(strings.NewReader(`{"A":"xyz"}`), DefaultJsonDecoder)
		assert.True(t, out.IsContainer())
		assert.Equal(t, "xyz", out.AsContainer().Child("A").AsLeaf().Value())
	})
}
