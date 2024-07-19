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
	sprig "github.com/go-task/slim-sprig/v3"
	"github.com/rkosegi/yaml-toolkit/dom"
	"github.com/rkosegi/yaml-toolkit/patch"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
	"strings"
	"testing"
)

var dummyExec = New().(*exec)

func newTestExec(d dom.ContainerBuilder) *exec {
	return New(WithData(d)).(*exec)
}

func parse[T any](t *testing.T, source string) *T {
	var x T
	err := yaml.NewDecoder(strings.NewReader(source)).Decode(&x)
	assert.NoError(t, err)
	if err != nil {
		return nil
	}
	return &x
}

func TestBoolExpressionEval(t *testing.T) {
	var (
		val bool
		err error
	)
	te := &templateEngine{
		fm: sprig.TxtFuncMap(),
	}
	expr := `{{ eq .Env "Development" }}`
	val, err = te.EvalBool(expr, map[string]interface{}{
		"Env": "Development",
	})
	assert.NoError(t, err)
	assert.Equal(t, true, val)

	val, err = te.EvalBool(expr, map[string]interface{}{
		"Env": "Production",
	})
	assert.NoError(t, err)
	assert.Equal(t, false, val)

	_, err = te.EvalBool(`{{`, map[string]interface{}{})
	assert.Error(t, err)
}

func TestParseStep(t *testing.T) {
	pl := parse[ActionSpec](t, `
---
name: root step
set:
  data:
    root:
      sub1:
        leaf1: 123
      sub2:
        - list_item1
  path: result
`)
	assert.NotNil(t, pl)
	assert.Contains(t, pl.String(), "root step")
	assert.Equal(t, "root step", pl.Name)
	assert.Equal(t, 123, pl.Operations.Set.Data["root"].(map[string]interface{})["sub1"].(map[string]interface{})["leaf1"])
	assert.Equal(t, "list_item1", pl.Operations.Set.Data["root"].(map[string]interface{})["sub2"].([]interface{})[0])
}

func TestExecute(t *testing.T) {
	var (
		gd dom.ContainerBuilder
		ex *exec
	)
	gd = b.Container()
	ex = newTestExec(gd)
	assert.NoError(t, ex.Execute(&ActionSpec{
		ActionMeta: ActionMeta{
			Name: "root step",
		},
		Operations: OpSpec{
			Set: &SetOp{
				Data: map[string]interface{}{
					"leaf": "abcd",
				},
			},
		},
		Children: ChildActions{
			"sub1": {
				Operations: OpSpec{
					Set: &SetOp{
						Path: "sub1.sub2",
						Data: map[string]interface{}{
							"leaf": "abcd",
						},
						Strategy: setStrategyPointer(SetStrategyReplace),
					},
				},
			},
		},
	}))
	assert.Equal(t, "abcd", gd.Lookup("leaf").(dom.Leaf).Value())
	assert.Equal(t, "abcd", gd.Lookup("sub1.sub2.leaf").(dom.Leaf).Value())

	gd = b.Container()
	ex = newTestExec(gd)
	assert.NoError(t, ex.Execute(&ActionSpec{
		ActionMeta: ActionMeta{
			Name: "root step",
		},
		Children: ChildActions{
			"sub_step1": {
				ActionMeta: ActionMeta{
					Name: "sub_step1",
				},
				Operations: OpSpec{
					Template: &TemplateOp{
						Template: "{{ mul 1 2 3 4 5 6 }}",
						Path:     "Results.Factorial",
					},
				},
			},
		},
	}))
	assert.Equal(t, "720", gd.Lookup("Results.Factorial").(dom.Leaf).Value())

}

func TestExecuteImport(t *testing.T) {
	var (
		gd dom.ContainerBuilder
		ex *exec
	)
	gd = b.Container()
	ex = newTestExec(gd)
	assert.NoError(t, ex.Execute(&ActionSpec{
		ActionMeta: ActionMeta{
			Name: "root step",
		},
		Operations: OpSpec{
			Import: &ImportOp{
				File: "../testdata/props1.properties",
				Path: "wrapped",
				Mode: ParseFileModeProperties,
			},
		},
	}))
	assert.Equal(t, "abcdef", gd.Lookup("wrapped.root.sub1.leaf2").(dom.Leaf).Value())
}

func TestExecuteImportInvalid(t *testing.T) {
	var (
		gd dom.ContainerBuilder
		ex *exec
	)
	gd = b.Container()
	ex = newTestExec(gd)
	assert.Error(t, ex.Execute(&ActionSpec{
		ActionMeta: ActionMeta{
			Name: "root step",
		},
		Operations: OpSpec{
			Import: &ImportOp{},
		},
	}))
}

func TestExecutePatch(t *testing.T) {
	var (
		gd dom.ContainerBuilder
		ex *exec
	)
	gd = b.Container()
	ex = newTestExec(gd)
	assert.NoError(t, ex.Execute(&ActionSpec{
		ActionMeta: ActionMeta{
			Name: "root step",
		},
		Operations: OpSpec{
			Patch: &PatchOp{
				Op:    patch.OpAdd,
				Path:  "/root",
				Value: map[string]interface{}{"leaf": "abcd"},
			},
		},
	}))
	assert.Equal(t, "abcd", gd.Lookup("root.leaf").(dom.Leaf).Value())
}

func TestExecuteInnerSteps(t *testing.T) {
	var (
		gd dom.ContainerBuilder
		ex *exec
	)
	gd = b.Container()
	ex = newTestExec(gd)
	assert.NoError(t, ex.Execute(&ActionSpec{
		ActionMeta: ActionMeta{
			Name: "root step",
		},
		Children: ChildActions{
			"step20": {
				ActionMeta: ActionMeta{
					Order: 20,
					Name:  "step 20",
				},
				Operations: OpSpec{
					Set: &SetOp{
						Data: map[string]interface{}{
							"root.sub": 123,
						},
						Strategy: setStrategyPointer(SetStrategyReplace),
					},
				},
			},
			"step30": {
				ActionMeta: ActionMeta{
					Order: 30,
					Name:  "step 30",
				},
				Operations: OpSpec{
					Set: &SetOp{
						Data: map[string]interface{}{
							"root.sub": 456,
						},
						Strategy: setStrategyPointer(SetStrategyReplace),
					},
				},
			},
		},
	}))
	assert.Equal(t, 456, gd.Lookup("root.sub").(dom.Leaf).Value())
}

func TestExecuteInnerStepsFail(t *testing.T) {
	assert.Error(t, dummyExec.Execute(&ActionSpec{
		ActionMeta: ActionMeta{
			Name: "root step",
		},
		Children: ChildActions{
			"step20": {
				ActionMeta: ActionMeta{
					Order: 20,
					Name:  "step 20",
				},
				Operations: OpSpec{
					Set: &SetOp{},
				},
			},
		},
	}))
}

func TestExecuteInnerStepsSkipped(t *testing.T) {
	assert.NoError(t, dummyExec.Execute(&ActionSpec{
		ActionMeta: ActionMeta{
			Name: "root step",
		},
		Children: ChildActions{
			"step20": {
				ActionMeta: ActionMeta{
					When: strPointer("{{ .Data.Skip | default false }}"),
					Name: "step 20",
				},
			},
		},
	}))
}

func TestExecuteInnerStepsWhenInvalid(t *testing.T) {
	assert.Error(t, dummyExec.Execute(&ActionSpec{
		ActionMeta: ActionMeta{
			Name: "root step",
		},
		Children: ChildActions{
			"step20": {
				ActionMeta: ActionMeta{
					When: strPointer("{{ .Data.Unknown.Field }}"),
					Name: "step 20",
				},
			},
		},
	}))
}

func TestExecuteForEachFileGlob(t *testing.T) {
	var (
		gd dom.ContainerBuilder
		ex *exec
	)
	gd = b.Container()
	ex = newTestExec(gd)
	fe := &ForEachOp{
		Glob: strPointer("../testdata/doc?.yaml"),
		Action: OpSpec{
			Import: &ImportOp{
				File: "{{ .forEach }}",
				Path: "import.files.{{ b64enc (osBase .forEach) }}",
				Mode: ParseFileModeYaml,
			},
		},
	}

	ss := &ActionSpec{
		ActionMeta: ActionMeta{
			Name: "root step",
		},
		Operations: OpSpec{
			ForEach: fe,
		},
	}
	assert.NoError(t, ex.Execute(ss))
	assert.Equal(t, 2, len(gd.Lookup("import.files").(dom.Container).Children()))
	assert.Contains(t, fe.String(), "doc?.yaml")
}
