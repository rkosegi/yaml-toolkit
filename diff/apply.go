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
	"github.com/rkosegi/yaml-toolkit/props"
)

var pp = props.NewPathParser()

func applySingle(node dom.ContainerBuilder, mod Modification) {
	switch mod.Type {
	case ModAdd, ModChange:
		node.Set(pp.MustParse(mod.Path), dom.LeafNode(mod.Value))

	case ModDelete:
		node.Delete(pp.MustParse(mod.Path))
	}
}

func Apply(node dom.ContainerBuilder, mods []Modification) {
	for _, mod := range mods {
		applySingle(node, mod)
	}
}
