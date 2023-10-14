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
	"testing"
)

func TestLeafValue(t *testing.T) {
	l := LeafNode(10)
	assert.False(t, l.IsContainer())
	assert.False(t, l.IsList())
	assert.True(t, l.IsLeaf())
	assert.Equal(t, 10, l.Value())

	l = LeafNode("abc")
	assert.False(t, l.IsContainer())
	assert.False(t, l.SameAs(nil))
	assert.True(t, l.SameAs(nilLeaf))
	assert.Equal(t, "abc", l.Value())
}
