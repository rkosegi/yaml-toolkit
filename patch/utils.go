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

import "github.com/rkosegi/yaml-toolkit/dom"

// removeListItem deletes item from given list. remaining Nodes are shifted to left by one
func removeListItem(list dom.ListBuilder, removeAt int) {
	items := list.Items()
	list.Clear()
	for i := 0; i < removeAt; i++ {
		list.Append(items[i])
	}
	for i := removeAt + 1; i < len(items); i++ {
		list.Append(items[i])
	}
}

// insertListItem inserts Node at given index.
func insertListItem(list dom.ListBuilder, insertAt int, toInsert dom.Node) {
	items := list.Items()
	list.Clear()
	for i := 0; i < insertAt; i++ {
		list.Append(items[i])
	}
	list.Append(toInsert)
	for i := insertAt; i < len(items); i++ {
		list.Append(items[i])
	}
}