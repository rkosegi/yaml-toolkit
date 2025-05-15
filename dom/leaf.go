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

import "github.com/google/go-cmp/cmp"

type leaf struct {
	base
	value interface{}
}

func (l *leaf) Clone() Node {
	return LeafNode(l.value)
}

func (l *leaf) Equals(node Node) bool {
	return node != nil && node.IsLeaf() && cmp.Equal(l.value, node.(Leaf).Value())
}

func (l *leaf) SameAs(node Node) bool {
	return node != nil && node.IsLeaf()
}

func (l *leaf) IsLeaf() bool {
	return true
}

func (l *leaf) Value() interface{} {
	return l.value
}

func (l *leaf) AsLeaf() Leaf {
	return l
}

func LeafNode(val interface{}) Leaf {
	ln := &leaf{value: val}
	ln.desc = "leaf"
	return ln
}
