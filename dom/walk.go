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

import "github.com/rkosegi/yaml-toolkit/path"

// NodeVisitorFn is function that is called for each child Node within the parent Node.
// Returning false from this function will terminate iteration.
type NodeVisitorFn func(p path.Path, parent Node, node Node) bool

func walkList(pb path.Builder, l List, fn NodeVisitorFn) {
	for idx, item := range l.Items() {
		if !fn(pb.Append(path.Numeric(idx)).Build(), l, item) {
			return
		}
	}
}

func walkContainer(pb path.Builder, c Container, fn NodeVisitorFn) {
	for k, v := range c.Children() {
		x := pb.Append(path.Simple(k))
		// BFS
		if !fn(x.Build(), c, v) {
			return
		}
		if v.IsContainer() {
			walkContainer(x, v.AsContainer(), fn)
		} else if v.IsList() {
			walkList(x, v.AsList(), fn)
		}
	}
}
