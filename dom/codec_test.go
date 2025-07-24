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
		Child("list1").AsList().Items()[0].AsContainer().
		Child("prop2").AsLeaf().Value())

	assert.Nil(t, decodeYamlNode(&yaml.Node{
		Kind: yaml.AliasNode,
	}))
}
