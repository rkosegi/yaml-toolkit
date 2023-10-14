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

package dom

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

var d = `
root:
  list:
    - item1: abc
      msg: Hi
    - 10.3
    - item3:
      - sub: 123
        msg: Hello`

func TestList(t *testing.T) {
	doc, err := b.FromReader(strings.NewReader(d), DefaultYamlDecoder)
	assert.NoError(t, err)
	assert.NotNil(t, doc)
	assert.False(t, doc.IsList())
	l := doc.Child("root").(Container).Child("list").(List)
	assert.False(t, l.IsContainer())
	assert.False(t, l.IsLeaf())
	assert.True(t, l.IsList())
	assert.Equal(t, 3, len(l.Items()))
	assert.Equal(t, 123, l.Items()[2].(Container).
		Child("item3").(List).Items()[0].(Container).
		Child("sub").(Leaf).Value())
	assert.False(t, l.SameAs(nilLeaf))
}

func TestMutateList(t *testing.T) {
	doc, err := b.FromReader(strings.NewReader(d), DefaultYamlDecoder)
	assert.NoError(t, err)
	assert.NotNil(t, doc)
	l := doc.Child("root").(Container).Child("list").(ListBuilder)
	assert.Equal(t, 3, len(l.Items()))

	l.Set(0, LeafNode(123))
	l.Set(1, LeafNode("abc"))

	assert.Equal(t, "abc", l.Items()[1].(Leaf).Value())
	assert.Equal(t, 123, l.Items()[0].(Leaf).Value())
	l.Clear()
	assert.Equal(t, 0, len(l.Items()))
}

func TestSetOutOfBounds(t *testing.T) {
	defer func() {
		recover()
	}()
	l := &listBuilderImpl{}
	l.Set(0, LeafNode(123))
	assert.Fail(t, "should not be here")
}
