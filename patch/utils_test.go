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

package patch

import (
	"github.com/rkosegi/yaml-toolkit/dom"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRemoveListItem(t *testing.T) {
	l := dom.ListNode(dom.LeafNode(1), dom.LeafNode(2), dom.LeafNode(3))
	assert.Equal(t, 3, len(l.Items()))
	removeListItem(l, 1)
	assert.Equal(t, 2, len(l.Items()))
	assert.Equal(t, 1, l.Items()[0].(dom.Leaf).Value())
	assert.Equal(t, 3, l.Items()[1].(dom.Leaf).Value())
}

func TestInsertListItem(t *testing.T) {
	l := dom.ListNode(dom.LeafNode("a"), dom.LeafNode("b"), dom.LeafNode("c"))
	assert.Equal(t, 3, len(l.Items()))
	insertListItem(l, 3, dom.LeafNode("d"))
	assert.Equal(t, 4, len(l.Items()))
	assert.Equal(t, "d", l.Items()[3].(dom.Leaf).Value())
}
