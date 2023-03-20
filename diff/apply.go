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

package diff

import (
	"github.com/rkosegi/yaml-toolkit/dom"
	"strings"
)

func applySingle(node dom.ContainerBuilder, mod Modification) {
	pc := strings.Split(mod.Path, ".")
	current := node
	switch mod.Type {
	case ModAdd, ModChange:
		for _, c := range pc[0 : len(pc)-1] {
			x := current.Child(c)
			if x == nil || !x.IsContainer() {
				current = current.AddContainer(c)
			} else {
				current = x.(dom.ContainerBuilder)
			}
		}
		current.AddValue(pc[len(pc)-1], mod.Value)

	case ModDelete:
		for _, c := range pc[0 : len(pc)-1] {
			x := current.Child(c)
			if x == nil || !x.IsContainer() {
				return
			} else {
				current = x.(dom.ContainerBuilder)
			}
		}
		current.Remove(pc[len(pc)-1])
	}
}

func Apply(node dom.ContainerBuilder, mods []Modification) {
	for _, mod := range mods {
		applySingle(node, mod)
	}
}
