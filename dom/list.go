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

import "fmt"

type listImpl struct {
	base
	items []Node
}

func initList() *listImpl {
	l := &listImpl{}
	l.desc = "list"
	return l
}

func (l *listImpl) IsList() bool {
	return true
}

func (l *listImpl) Items() []Node {
	c := make([]Node, l.Size())
	copy(c, l.items)
	return c
}

func (l *listImpl) AsSlice() []interface{} {
	return encodeListFn(l)
}

func (l *listImpl) Size() int {
	return len(l.items)
}

func (l *listImpl) Clone() Node {
	l2 := initList()
	for _, item := range l.items {
		l2.items = append(l2.items, item.Clone())
	}
	return l2
}

func (l *listImpl) AsList() List {
	return l
}

func (l *listImpl) AsAny() any {
	return encodeListFn(l)
}

func (l *listImpl) Equals(node Node) bool {
	if node == nil || !node.IsList() {
		return false
	}
	otherItems := node.(List).Items()
	if len(otherItems) != len(l.items) {
		return false
	}
	for i := 0; i < len(l.items); i++ {
		if !l.items[i].Equals(otherItems[i]) {
			return false
		}
	}
	return true
}

func (l *listImpl) SameAs(node Node) bool {
	return node != nil && node.IsList()
}

type listBuilderImpl struct {
	listImpl
}

func initListBuilder() *listBuilderImpl {
	lb := &listBuilderImpl{}
	lb.desc = "writable list"
	return lb
}

func (l *listBuilderImpl) Clear() ListBuilder {
	l.items = []Node{}
	return l
}

func (l *listBuilderImpl) Set(index uint, item Node) ListBuilder {
	for i := 0; i <= int(index); i++ {
		if i > len(l.items)-1 {
			l.Append(nilLeaf)
		}
	}
	l.items[index] = item
	return l
}

func (l *listBuilderImpl) MustSet(index uint, item Node) ListBuilder {
	if int(index) > len(l.items)-1 {
		panic(fmt.Sprintf("index out of bounds: %d", index))
	}
	l.items[index] = item
	return l
}

func (l *listBuilderImpl) Append(item Node) ListBuilder {
	l.items = append(l.items, item)
	return l
}

func (l *listBuilderImpl) Seal() List {
	return &l.listImpl
}

func ListNode(items ...Node) ListBuilder {
	l := initListBuilder()
	for _, item := range items {
		l.Append(item)
	}
	return l
}
