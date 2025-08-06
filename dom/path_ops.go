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
	"github.com/rkosegi/yaml-toolkit/path"
)

// Node operations that uses path.Path

func applyToContainer(tgt ContainerBuilder, c path.Component, addList bool) Node {
	var x Node
	if x = tgt.Child(c.Value()); x == nil {
		if addList {
			x = ListNode()
		} else {
			x = ContainerNode()
		}
		tgt.AddValue(c.Value(), x)
	}
	return x
}

func applyToList(tgt ListBuilder, c path.Component, addList bool) Node {
	var x Node
	if tgt.Size() > c.NumericValue() {
		x = tgt.Get(c.NumericValue())
	} else {
		if addList {
			x = ListNode()
		} else {
			x = ContainerNode()
		}
		tgt.Set(uint(c.NumericValue()), x)
	}
	return x
}

func unsealListIfNeeded(l List) ListBuilder {
	switch l := l.(type) {
	case ListBuilder:
		return l
	default:
		lb := initListBuilder()
		lb.items = l.Items()
		return lb
	}
}

func unsealContainerIfNeeded(c Container) ContainerBuilder {
	switch c := c.(type) {
	case ContainerBuilder:
		return c
	default:
		cb := initContainerBuilder()
		cb.children = c.Children()
		return cb
	}
}

// applyToNode applies value val to targetNode at given path.
// Empty path is not allowed and will result in no-op.
// This function returns modified target node.
func applyToNode(targetNode Node, at path.Path, val Node) Node {
	var tgt Node
	if at.IsEmpty() {
		return targetNode
	}
	cp := at.Components()
	tgt = targetNode
	lastIdx := len(cp) - 1

	for idx, c := range cp[0 : len(cp)-1] {
		nextIsNum := false
		if idx+1 < len(cp) && cp[idx+1].IsNumeric() {
			nextIsNum = true
		}
		if c.IsNumeric() {
			tgt = applyToList(unsealListIfNeeded(tgt.(List)), c, nextIsNum)
		} else {
			tgt = applyToContainer(unsealContainerIfNeeded(tgt.AsContainer()), c, nextIsNum)
		}
	}
	c := cp[lastIdx]
	if c.IsNumeric() {
		unsealListIfNeeded(tgt.(List)).Set(uint(c.NumericValue()), val)
	} else {
		unsealContainerIfNeeded(tgt.AsContainer()).AddValue(c.Value(), val)
	}
	return targetNode
}

func removeChild(n Node, c path.Component) {
	if c.IsNumeric() {
		n.(ListBuilder).Set(uint(c.NumericValue()), nil)
	} else {
		n.(ContainerBuilder).Remove(c.Value())
	}
}

// removeFromNode removes Node from targetNode at given path.
// First non-existent node along the path will terminate iteration.
// This function returns (eventually modified) targetNode.
func removeFromNode(targetNode Node, at path.Path) Node {
	var p path.Path
	p = path.ParentOf(at)
	if p != nil {
		if x := getFromNode(targetNode, p); x == nil {
			return targetNode
		} else {
			removeChild(x, at.Last())
		}
	}
	return targetNode
}

// getFromNode gets value from source node at given path.
// First non-existent node along the path will terminate iteration and nil will be returned immediately.
// This function does not modify source node in any way.
func getFromNode(sourceNode Node, at path.Path) Node {
	var src Node
	src = sourceNode
	for _, pc := range at.Components() {
		if pc.IsNumeric() {
			src = src.AsList().Get(pc.NumericValue())
		} else {
			src = src.AsContainer().Child(pc.Value())
		}
		// non-existent node along the path
		if src == nil {
			return nil
		}
	}
	return src
}
