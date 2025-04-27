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

package dom

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCastAsSubtypes(t *testing.T) {
	asContainer := func(n Node) { n.AsContainer() }
	asLeaf := func(n Node) { n.AsLeaf() }
	asList := func(n Node) { n.AsList() }
	container := b.Container()
	list := ListNode()
	leafNode := nilLeaf
	assert.NotNil(t, container.AsContainer())

	checkFn := func(node Node, nfn func(n Node)) bool {
		r := true
		defer func() {
			recover()
			r = false
		}()
		nfn(node)
		return r
	}
	assert.True(t, checkFn(container, asContainer))
	assert.False(t, checkFn(container, asLeaf))
	assert.False(t, checkFn(container, asList))
	assert.False(t, checkFn(list, asContainer))
	assert.False(t, checkFn(list, asLeaf))
	assert.True(t, checkFn(list, asList))
	assert.False(t, checkFn(leafNode, asContainer))
	assert.True(t, checkFn(leafNode, asLeaf))
	assert.False(t, checkFn(leafNode, asList))

}
