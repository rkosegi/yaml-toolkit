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

type leaf struct {
	value interface{}
}

func (b *leaf) SameAs(node Node) bool {
	return node != nil && node.IsLeaf()
}

func (b *leaf) IsContainer() bool {
	return false
}

func (b *leaf) IsLeaf() bool {
	return true
}

func (b *leaf) IsList() bool {
	return false
}

func (b *leaf) Value() interface{} {
	return b.value
}

func LeafNode(val interface{}) Leaf {
	return &leaf{value: val}
}

var _ Leaf = &leaf{}
