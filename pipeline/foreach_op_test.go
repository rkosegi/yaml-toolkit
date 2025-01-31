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
		Item:   &([]string{"a", "b", "c"}),
		Action: OpSpec{},
	}
	a := op.CloneWith(mockEmptyActCtx()).(*ForEachOp)
	assert.NotNil(t, a)
	assert.Equal(t, 3, len(*a.Item))
}

func TestForeachStringItem(t *testing.T) {
	op := ForEachOp{
		Item: &([]string{"a", "b", "c"}),
		Action: OpSpec{
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
		Action: OpSpec{
			Set: &SetOp{
				Path: "{{ .forEach }}",
			},
		},
	}
	d := b.Container()
	err := op.Do(mockActCtx(d))
	assert.Error(t, err)
}

func TestForeachGlob(t *testing.T) {
	op := ForEachOp{
		Glob: strPointer("../testdata/doc?.yaml"),
		Action: OpSpec{
			Import: &ImportOp{
				File: "{{ .forEach }}",
				Path: "import.files.{{ b64enc (osBase .forEach) }}",
				Mode: ParseFileModeYaml,
			},
		},
	}
	d := b.Container()
	err := op.Do(mockActCtx(d))
	assert.NoError(t, err)
	assert.Equal(t, 2, len(d.Lookup("import.files").(dom.Container).Children()))
}

func TestForeachGlobChildError(t *testing.T) {
	op := ForEachOp{
		Glob: strPointer("../testdata/doc?.yaml"),
		Action: OpSpec{
			Set: &SetOp{
				Path: "{{ .forEach }}",
			},
		},
	}
	d := b.Container()
	err := op.Do(mockActCtx(d))
	assert.Error(t, err)
}

func TestForeachGlobInvalid(t *testing.T) {
	op := ForEachOp{
		Glob: strPointer("[]]"),
		Action: OpSpec{
			Import: &ImportOp{
				File: "{{ .forEach }}",
				Path: "import.files.{{ b64enc (osBase .forEach) }}",
				Mode: ParseFileModeYaml,
			},
		},
	}
	d := b.Container()
	err := op.Do(mockActCtx(d))
	assert.Error(t, err)
}
