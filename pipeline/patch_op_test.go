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

package pipeline

import (
	"testing"

	"github.com/rkosegi/yaml-toolkit/dom"
	"github.com/rkosegi/yaml-toolkit/patch"
	"github.com/stretchr/testify/assert"
)

func TestExecutePatchOp(t *testing.T) {
	var (
		ps PatchOp
		gd dom.ContainerBuilder
	)

	ps = PatchOp{
		Op:   patch.OpAdd,
		Path: "@#$%^&",
	}
	assert.Error(t, New(WithData(gd)).Execute(&ps))

	gd = b.Container()
	gd.AddValueAt("root.sub1.leaf2", dom.LeafNode("abcd"))
	ps = PatchOp{
		Op:   patch.OpReplace,
		Path: "/root/sub1",
		Value: anyValFromMap(map[string]interface{}{
			"leaf2": "xyz",
		}),
	}
	assert.NoError(t, New(WithData(gd)).Execute(&ps))
	assert.Equal(t, "xyz", gd.Lookup("root.sub1.leaf2").(dom.Leaf).Value())
	assert.Contains(t, ps.String(), "Op=replace,Path=/root/sub1")

	gd = b.Container()
	gd.AddValueAt("root.sub1.leaf3", dom.LeafNode("abcd"))
	ps = PatchOp{
		Op:   patch.OpMove,
		From: "/root/sub1",
		Path: "/root/sub2",
	}
	assert.NoError(t, New(WithData(gd)).Execute(&ps))
	assert.Equal(t, "abcd", gd.Lookup("root.sub2.leaf3").(dom.Leaf).Value())

	gd = b.Container()
	gd.AddValueAt("root.sub1.leaf3", dom.LeafNode("abcd"))
	ps = PatchOp{
		Op:   patch.OpMove,
		From: "%#$&^^*&",
		Path: "/root/sub2",
	}
	assert.Error(t, New(WithData(gd)).Execute(&ps))
}

func TestPatchOpAddValue(t *testing.T) {
	var (
		ps PatchOp
		gd dom.ContainerBuilder
	)
	gd = b.Container()
	gd.AddValueAt("root.sub.leaf1", dom.LeafNode("123"))
	ps = PatchOp{
		Op:    patch.OpAdd,
		Path:  "/root/sub/leaf2",
		Value: &AnyVal{v: dom.LeafNode(456)},
	}
	assert.NoError(t, New(WithData(gd)).Execute(&ps))
	m := gd.AsMap()
	assert.Equal(t, "123", m["root"].(map[string]interface{})["sub"].(map[string]interface{})["leaf1"].(string))
	assert.Equal(t, 456, m["root"].(map[string]interface{})["sub"].(map[string]interface{})["leaf2"].(int))
}
