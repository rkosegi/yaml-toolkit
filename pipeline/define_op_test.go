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

	"github.com/stretchr/testify/assert"
)

func TestForDefCall(t *testing.T) {
	root := ActionSpec{
		Children: ChildActions{
			"init": ActionSpec{
				ActionMeta: ActionMeta{
					Order: 1,
				},
				Operations: OpSpec{
					Define: &DefineOp{
						Name: "dummy",
						Action: ActionSpec{
							Operations: OpSpec{
								Log: &LogOp{
									Message: "{{ .myargs.msg }}",
								},
							},
						},
					},
				},
			},
			"run": ActionSpec{
				ActionMeta: ActionMeta{
					Order: 2,
				},
				Operations: OpSpec{
					ForEach: &ForEachOp{
						Item: &(ValOrRefSlice{&ValOrRef{Val: "a"}, &ValOrRef{Val: "b"}}),
						Action: ActionSpec{
							Operations: OpSpec{
								Call: &CallOp{
									Name: "dummy",
									Args: map[string]interface{}{
										"msg": "{{ .forEach }}",
									},
									ArgsPath: ptr("myargs"),
								},
							},
						},
					},
				},
			},
		},
	}

	ctx := newMockActBuilder().testLogger(t).build()
	err := ctx.Executor().Execute(root)
	assert.NoError(t, err)
}

func TestDefineTwice(t *testing.T) {
	ctx := mockEmptyActCtx()
	assert.NoError(t, ctx.Executor().Execute(&DefineOp{
		Name:   "op",
		Action: ActionSpec{},
	}))
	assert.Error(t, ctx.Executor().Execute(&DefineOp{
		Name:   "op",
		Action: ActionSpec{},
	}))
}
