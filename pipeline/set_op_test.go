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
	"github.com/rkosegi/yaml-toolkit/dom"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestExecuteSetOp(t *testing.T) {
	var (
		ss  SetOp
		gd  dom.ContainerBuilder
		err error
	)
	ss = SetOp{
		Data: map[string]interface{}{
			"sub1": 123,
		},
	}
	gd = b.Container()
	assert.NoError(t, New(WithData(gd)).Execute(&ss))
	assert.Equal(t, 123, gd.Lookup("sub1").(dom.Leaf).Value())

	ss = SetOp{
		Data: map[string]interface{}{
			"sub1": 123,
		},
		Path: "sub0",
	}
	gd = b.Container()
	assert.NoError(t, New(WithData(gd)).Execute(&ss))
	assert.Equal(t, 123, gd.Lookup("sub0.sub1").(dom.Leaf).Value())
	assert.Contains(t, ss.String(), "sub0")

	ss = SetOp{}
	err = New(WithData(gd)).Execute(&ss)
	assert.Error(t, err)
	assert.Equal(t, ErrNoDataToSet, err)
}

func TestSetOpInvalidSetStrategy(t *testing.T) {
	assert.Error(t, New().Execute(&SetOp{
		Data:     map[string]interface{}{},
		Strategy: setStrategyPointer("unknown"),
	}))
}

func TestSetOpMergeRoot(t *testing.T) {
	var (
		gd dom.ContainerBuilder
		ss SetOp
	)

	ss = SetOp{
		Data: map[string]interface{}{
			"sub1": 123,
		},
		Strategy: setStrategyPointer(SetStrategyMerge),
	}
	gd = b.Container()
	gd.AddValue("sub2", dom.LeafNode(1))
	assert.NoError(t, New(WithData(gd)).Execute(&ss))
	assert.Equal(t, 123, gd.Lookup("sub1").(dom.Leaf).Value())
	assert.Equal(t, 2, len(gd.Children()))

	gd = b.Container()
	gd.AddValueAt("sub2.sub3a", dom.LeafNode(2))
	ss = SetOp{
		Data: map[string]interface{}{
			"sub2": map[string]interface{}{
				"sub3b": 123,
			},
		},
		Strategy: setStrategyPointer(SetStrategyMerge),
	}
	assert.NoError(t, New(WithData(gd)).Execute(&ss))
	assert.Equal(t, 2, len(gd.Lookup("sub2").(dom.Container).Children()))
	assert.Equal(t, 2, gd.Lookup("sub2.sub3a").(dom.Leaf).Value())
	assert.Equal(t, 123, gd.Lookup("sub2.sub3b").(dom.Leaf).Value())
}

func TestSetOpMergeSubPath(t *testing.T) {
	var (
		gd dom.ContainerBuilder
		ss SetOp
	)

	ss = SetOp{
		Data: map[string]interface{}{
			"sub20": 123,
		},
		Strategy: setStrategyPointer(SetStrategyMerge),
		Path:     "sub10",
	}
	gd = b.Container()

	assert.NoError(t, New(WithData(gd)).Execute(&ss))
	assert.Equal(t, 123, gd.Lookup("sub10.sub20").(dom.Leaf).Value())

	gd = b.Container()
	gd.AddValueAt("sub10.sub20.sub30", dom.LeafNode(2))
	ss = SetOp{
		Data: map[string]interface{}{
			"sub20": map[string]interface{}{
				"sub3b": 123,
			},
		},
		Path:     "sub10",
		Strategy: setStrategyPointer(SetStrategyMerge),
	}
	assert.NoError(t, New(WithData(gd)).Execute(&ss))
	assert.Equal(t, 2, len(gd.Lookup("sub10.sub20").(dom.Container).Children()))
	assert.Equal(t, 2, gd.Lookup("sub10.sub20.sub30").(dom.Leaf).Value())
	assert.Equal(t, 123, gd.Lookup("sub10.sub20.sub3b").(dom.Leaf).Value())
}
