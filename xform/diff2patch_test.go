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
	"testing"

	"github.com/rkosegi/yaml-toolkit/diff"
	"github.com/rkosegi/yaml-toolkit/dom"
	"github.com/rkosegi/yaml-toolkit/patch"
	"github.com/stretchr/testify/assert"
)

func TestDiffMod2PatchOpAdd(t *testing.T) {
	var op *patch.OpObj
	op = DiffMod2PatchOp(diff.Modification{
		Type:  diff.ModAdd,
		Path:  "list1[1][2][3].item_obj20.sub.sublist[3].efgh[0]",
		Value: "abc",
	})
	assert.NotNil(t, op)
	assert.Equal(t, patch.OpAdd, op.Op)
	assert.Equal(t, "abc", op.Value.(dom.Leaf).Value())
	assert.Equal(t, 10, len(op.Path))
	assert.Equal(t, "0", string(op.Path[9]))
	assert.Equal(t, "list1", string(op.Path[0]))

	op = DiffMod2PatchOp(diff.Modification{
		Type: diff.ModDelete,
		Path: "root.sub1.sub2",
	})
	assert.NotNil(t, op)
	assert.Equal(t, patch.OpRemove, op.Op)
	assert.Equal(t, 3, len(op.Path))

	op = DiffMod2PatchOp(diff.Modification{
		Type:  diff.ModChange,
		Path:  "root.sub1.sub2",
		Value: 123,
	})
	assert.NotNil(t, op)
	assert.Equal(t, patch.OpReplace, op.Op)
	assert.Equal(t, 3, len(op.Path))

	op = DiffMod2PatchOp(diff.Modification{
		Type: "invalid",
	})
	assert.Nil(t, op)
}
