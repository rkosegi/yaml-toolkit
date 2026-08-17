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

package xform

import (
	"github.com/rkosegi/yaml-toolkit/diff"
	"github.com/rkosegi/yaml-toolkit/dom"
	"github.com/rkosegi/yaml-toolkit/patch"
)

// DiffMod2PatchOp converts given diff.Modification into *OpObj
func DiffMod2PatchOp(mod diff.Modification) *patch.OpObj {
	switch mod.Type {
	case diff.ModAdd:
		return &patch.OpObj{
			Op:    patch.OpAdd,
			Path:  PropPath2Pointer(mod.Path),
			Value: dom.LeafNode(mod.Value),
		}
	case diff.ModDelete:
		return &patch.OpObj{
			Path: PropPath2Pointer(mod.Path),
			Op:   patch.OpRemove,
		}
	case diff.ModChange:
		return &patch.OpObj{
			Op:    patch.OpReplace,
			Path:  PropPath2Pointer(mod.Path),
			Value: dom.LeafNode(mod.Value),
		}
	default:
		return nil
	}
}
