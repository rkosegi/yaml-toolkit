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
	"github.com/stretchr/testify/assert"
)

func TestLoopOpCloneWith(t *testing.T) {
	op := &LoopOp{
		Init: &ActionSpec{
			Operations: OpSpec{
				Log: &LogOp{
					Message: "Ola",
				},
			},
		},
		PostAction: &ActionSpec{
			Operations: OpSpec{
				Abort: &AbortOp{
					Message: "Hi",
				},
			},
		},
		Test: "{{ false }}",
	}
	op = op.CloneWith(mockEmptyActCtx()).(*LoopOp)
	assert.Equal(t, "{{ false }}", op.Test)
	assert.Equal(t, "Ola", op.Init.Operations.Log.Message)
	assert.Equal(t, "Hi", op.PostAction.Operations.Abort.Message)
}

func TestLoopOpSimple(t *testing.T) {
	op := LoopOp{
		Init: &ActionSpec{
			Operations: OpSpec{
				Set: &SetOp{
					Data: map[string]interface{}{
						"i": 0,
					},
				},
			},
		},
		Action: ActionSpec{
			Operations: OpSpec{
				Log: &LogOp{
					Message: "Iteration {{ .i }}",
				},
			},
		},
		Test: "{{ lt (.i | int) 10 }}",
		PostAction: &ActionSpec{
			Operations: OpSpec{
				Template: &TemplateOp{
					Template: "{{ add .i  1 }}",
					Path:     "i",
				},
			},
		},
	}

	d := b.Container()
	ac := newMockActBuilder().data(d).build()
	err := op.Do(ac)
	assert.NoError(t, err)
	assert.Equal(t, "10", d.Lookup("i").(dom.Leaf).Value())
}

func TestLoopOpNegative(t *testing.T) {
	var (
		err error
		op  *LoopOp
	)
	op = &LoopOp{
		Init: &ActionSpec{
			Operations: OpSpec{
				Abort: &AbortOp{},
			},
		},
	}
	err = op.Do(mockEmptyActCtx())
	assert.Error(t, err)

	op = &LoopOp{
		Test: "{{ true }}",
		Action: ActionSpec{
			Operations: OpSpec{
				Abort: &AbortOp{},
			},
		},
	}
	err = op.Do(mockEmptyActCtx())
	assert.Error(t, err)

	op = &LoopOp{
		PostAction: &ActionSpec{
			Operations: OpSpec{
				Abort: &AbortOp{},
			},
		},
		Test:   "{{ true }}",
		Action: ActionSpec{},
	}
	err = op.Do(mockEmptyActCtx())
	assert.Error(t, err)

	op = &LoopOp{
		Test:   "{{ NotAFunction }}",
		Action: ActionSpec{},
	}
	err = op.Do(mockEmptyActCtx())
	assert.Error(t, err)

}
