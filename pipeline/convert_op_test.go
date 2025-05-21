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

package pipeline

import (
	"testing"

	"github.com/rkosegi/yaml-toolkit/dom"
	"github.com/stretchr/testify/assert"
)

func TestConvertOp(t *testing.T) {
	var err error
	gd := b.Container()
	gd.AddValue("x", dom.LeafNode("123"))
	ex := New(WithData(gd))

	err = ex.Execute(&ConvertOp{
		Path:   "x",
		Format: LeafValueFormatFloat64,
	})
	assert.NoError(t, err)
	assert.Equal(t, 123.0, gd.Lookup("x").AsLeaf().Value())

	gd.AddValue("x", dom.LeafNode("456"))
	err = ex.Execute(&ConvertOp{
		Path:   "x",
		Format: LeafValueFormatInt64,
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(456), gd.Lookup("x").AsLeaf().Value())

	gd.AddValue("x", dom.LeafNode("xyz"))
	for _, op := range []ConvertOp{
		{
			Path:   "x",
			Format: "unknown",
		},
		{
			Path: "x",
		},
		{
			Format: "unknown",
		},
		{
			Path:   "x",
			Format: LeafValueFormatFloat64,
		},
		{
			Path:   "x",
			Format: LeafValueFormatInt64,
		},
	} {
		t.Log("invalid op:", op.String())
		assert.Error(t, ex.Execute(&op))
	}
}

func TestConvertOpCloneWith(t *testing.T) {
	op := ConvertOp{
		Path:   "a.b.c",
		Format: LeafValueFormatFloat64,
	}
	co := op.CloneWith(newMockActBuilder().build()).(*ConvertOp)
	assert.Equal(t, "a.b.c", co.Path)
}
