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

func TestForeachCloneWith(t *testing.T) {
	op := ForEachOp{
		Item: &([]string{"a", "b", "c"}),
		Action: ActionSpec{
			Operations: OpSpec{},
		},
	}
	a := op.CloneWith(mockEmptyActCtx()).(*ForEachOp)
	assert.NotNil(t, a)
	assert.Equal(t, 3, len(*a.Item))
}

func TestForeachStringItem(t *testing.T) {
	op := ForEachOp{
		Item: &([]string{"a", "b", "c"}),
		Action: ActionSpec{
			Operations: OpSpec{
				Set: &SetOp{
					Path: "{{ .forEach }}",
					Data: map[string]interface{}{
						"X": "abc",
					},
				},
				Env: &EnvOp{},
				Export: &ExportOp{
					File:   "/tmp/a-{{ .forEach }}.yaml",
					Format: OutputFormatYaml,
				},
				Exec: &ExecOp{
					Program: "sh",
					Args:    &[]string{"-c", "rm /tmp/a-{{ .forEach }}.yaml"},
				},
				Log: &LogOp{
					Message: "Hi {{ .forEach }}",
				},
				Loop: &LoopOp{
					Test: "false",
					Action: ActionSpec{
						Operations: OpSpec{Log: &LogOp{
							Message: "Ola!",
						}},
					},
				},
			},
		},
	}
	d := b.Container()
	err := op.Do(mockActCtx(d))
	assert.NoError(t, err)
	assert.Equal(t, "abc", d.Lookup("a.X").(dom.Leaf).Value())
	assert.Equal(t, "abc", d.Lookup("b.X").(dom.Leaf).Value())
	assert.Equal(t, "abc", d.Lookup("c.X").(dom.Leaf).Value())
}

func TestForeachStringItemChildError(t *testing.T) {
	op := ForEachOp{
		Item: &([]string{"a", "b", "c"}),
		Action: ActionSpec{
			Operations: OpSpec{
				Set: &SetOp{
					Path: "{{ .forEach }}",
				},
			},
		},
	}
	d := b.Container()
	err := op.Do(mockActCtx(d))
	assert.Error(t, err)
}

func TestForeachQuery(t *testing.T) {
	var (
		err error
		op  *ForEachOp
	)
	d := b.FromMap(map[string]interface{}{
		"items": []interface{}{"a", "b", "c"},
	})
	op = &ForEachOp{
		Query: ptr("items"),
		Action: ActionSpec{
			Operations: OpSpec{
				Template: &TemplateOp{
					Template: "{{ .forEach }}",
					Path:     "Result.{{ .forEach }}",
				},
			},
		},
	}
	err = op.Do(mockActCtx(d))
	assert.Equal(t, 3, len(d.Lookup("Result").(dom.Container).Children()))
	assert.NoError(t, err)
	op = &ForEachOp{
		Query: ptr("items"),
		Action: ActionSpec{
			Operations: OpSpec{
				Abort: &AbortOp{},
			},
		},
	}
	err = op.Do(mockActCtx(d))
	assert.Error(t, err)
}

func TestForeachGlob(t *testing.T) {
	op := ForEachOp{
		Glob: strPointer("../testdata/doc?.yaml"),
		Action: ActionSpec{
			Operations: OpSpec{
				Import: &ImportOp{
					File: "{{ .forEach }}",
					Path: "import.files.{{ b64enc (osBase .forEach) }}",
					Mode: ParseFileModeYaml,
				},
			},
		},
	}
	d := b.Container()
	err := op.Do(mockActCtx(d))
	assert.NoError(t, err)
	assert.Equal(t, 2, len(d.Lookup("import.files").(dom.Container).Children()))
}

func TestForeachActionSpec(t *testing.T) {
	var (
		err error
		op  *ForEachOp
	)
	op = &ForEachOp{
		Item: &([]string{"a", "b", "c"}),
		Action: ActionSpec{
			Children: map[string]ActionSpec{
				"sub": {
					Operations: OpSpec{
						Log: &LogOp{
							Message: "Hi {{ .forEach }}",
						},
					},
				},
			},
		},
	}
	err = op.Do(mockActCtxLog(t))
	assert.NoError(t, err)

	op = &ForEachOp{
		Item: &([]string{"a", "b", "c"}),
		Action: ActionSpec{
			Children: map[string]ActionSpec{
				"sub": {
					Operations: OpSpec{
						Template: &TemplateOp{
							Path:     "X",
							Template: "{{ add .X 1 }}",
						},
					},
				},
			},
		},
	}
	d := b.Container()
	d.AddValue("X", dom.LeafNode(100))
	err = op.Do(mockActCtx(d))
	assert.NoError(t, err)
	assert.Equal(t, "103", d.Lookup("X").(dom.Leaf).Value())
}

func TestForeachGlobChildError(t *testing.T) {
	op := ForEachOp{
		Glob: strPointer("../testdata/doc?.yaml"),
		Action: ActionSpec{
			Operations: OpSpec{
				Set: &SetOp{
					Path: "{{ .forEach }}",
				},
			},
		},
	}
	err := op.Do(mockActCtxLog(t))
	assert.Error(t, err)
}

func TestForeachGlobInvalid(t *testing.T) {
	op := ForEachOp{
		Glob: strPointer("[]]"),
		Action: ActionSpec{
			Operations: OpSpec{
				Import: &ImportOp{
					File: "{{ .forEach }}",
					Path: "import.files.{{ b64enc (osBase .forEach) }}",
					Mode: ParseFileModeYaml,
				},
			},
		},
	}
	d := b.Container()
	err := op.Do(mockActCtx(d))
	assert.Error(t, err)
}
