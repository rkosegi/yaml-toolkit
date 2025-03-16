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
				TemplateFile: &TemplateFileOp{
					File:   "../testdata/simple.template",
					Output: "/tmp/abc.out",
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
	err := op.Do(newMockActBuilder().data(d).build())
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
	err := op.Do(newMockActBuilder().data(d).build())
	assert.Error(t, err)
}

func TestForeachQuery(t *testing.T) {
	var (
		err error
	)

	type testcase struct {
		qry        string
		tmpl       string
		variable   string
		path       string
		validateFn func(data dom.ContainerBuilder)
	}
	data := b.FromMap(map[string]interface{}{
		"leaf": "X",
		"sub": map[string]interface{}{
			"leaf1": "Y",
		},
		"items": []interface{}{"a", "b", "c"},
	})
	for _, qry := range []string{
		"leaf", "sub", "items",
	} {
		op := &ForEachOp{
			Query: &qry,
			Action: ActionSpec{
				Operations: OpSpec{
					Abort: &AbortOp{},
				},
			},
		}
		assert.Error(t, op.Do(newMockActBuilder().data(data).build()))
	}
	for _, tc := range []testcase{
		{
			qry:      "leaf",
			tmpl:     "{{ .forEach }}",
			variable: "forEach",
			path:     "Result",
			validateFn: func(d dom.ContainerBuilder) {
				assert.Equal(t, "X", d.Lookup("Result").(dom.Leaf).Value())
			},
		},
		{
			validateFn: func(d dom.ContainerBuilder) {
				assert.Equal(t, "Y", d.Lookup("Result").(dom.Leaf).Value())
			},
			qry:      "sub",
			tmpl:     "{{ get .sub .forEach }}",
			variable: "forEach",
			path:     "Result",
		},
		{
			validateFn: func(d dom.ContainerBuilder) {
				assert.Equal(t, 3, len(d.Lookup("Result").(dom.Container).Children()))
			},
			qry:      "items",
			tmpl:     "{{ .XYZ }}",
			variable: "XYZ",
			path:     "Result.{{ .XYZ }}",
		},
	} {
		op := &ForEachOp{
			Variable: ptr(tc.variable),
			Query:    ptr(tc.qry),
			Action: ActionSpec{
				Operations: OpSpec{
					Template: &TemplateOp{
						Template: tc.tmpl,
						Path:     tc.path,
					},
				},
			},
		}
		err = op.Do(newMockActBuilder().data(data).build())
		assert.NoError(t, err)
		tc.validateFn(data)
	}
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
	err := op.Do(newMockActBuilder().data(d).build())
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
	err = op.Do(newMockActBuilder().testLogger(t).build())
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
	err = op.Do(newMockActBuilder().data(d).build())
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
	err := op.Do(newMockActBuilder().testLogger(t).build())
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
	err := op.Do(newMockActBuilder().data(d).build())
	assert.Error(t, err)
}
