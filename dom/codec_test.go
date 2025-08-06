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
	"gopkg.in/yaml.v3"
)

func TestDecodeYamlNode(t *testing.T) {
	doc1 := `
root:
  list1:
    - item1: abc
      prop2: 123
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
	assert.Equal(t, 1, dn.AsContainer().
		Child("root").AsContainer().
		Child("list1").AsList().Size())
	assert.Equal(t, "123", dn.AsContainer().
		Child("root").AsContainer().
		Child("list1").AsList().Get(0).AsContainer().
		Child("prop2").AsLeaf().Value())

	assert.Nil(t, decodeYamlNode(&yaml.Node{
		Kind: yaml.AliasNode,
	}))
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

	y1, err := b.FromReader(strings.NewReader(y1Src), DefaultYamlDecoder)
	assert.NoError(t, err)
	res := DecodeAnyToNode(x1).(ContainerBuilder)
	assert.NotNil(t, res)
	assert.Equal(t, 13, res.Child("A").AsLeaf().Value())
	assert.Equal(t, y1.Seal(), res.Seal())

	// chan is not considered during decode
	assert.Nil(t, DecodeAnyToNode(make(chan struct{})))

	assert.Equal(t, nilLeaf, DecodeAnyToNode(nil))
}
