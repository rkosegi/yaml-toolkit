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

func TestExecuteTemplateOp(t *testing.T) {
	var (
		err error
		ts  TemplateOp
		gd  dom.ContainerBuilder
	)

	gd = b.Container()
	gd.AddValueAt("root.leaf1", dom.LeafNode(123456))
	ts = TemplateOp{
		Template: `{{ (mul .root.leaf1 2) | quote }}`,
		Path:     "result.x1",
		Trim:     ptr(true),
	}
	assert.NoError(t, New(WithData(gd)).Execute(&ts))
	assert.Equal(t, "\"246912\"", gd.Lookup("result.x1").(dom.Leaf).Value())
	assert.Contains(t, ts.String(), "result.x1")

	// empty template error
	ts = TemplateOp{}
	err = New(WithData(gd)).Execute(&ts)
	assert.Error(t, err)
	assert.Equal(t, ErrTemplateEmpty, err)

	// empty path error
	ts = TemplateOp{
		Template: `TEST`,
	}
	err = New(WithData(gd)).Execute(&ts)
	assert.Error(t, err)
	assert.Equal(t, ErrPathEmpty, err)

	ts = TemplateOp{
		Template: `{{}}{{`,
		Path:     "result",
	}
	assert.Error(t, New(WithData(gd)).Execute(&ts))

	ts = TemplateOp{
		Template: `{{ invalid_func }}`,
		Path:     "result",
	}
	assert.Error(t, New(WithData(gd)).Execute(&ts))
}

func TestExecuteTemplateOpAsYaml(t *testing.T) {
	var (
		err error
		ts  TemplateOp
		gd  dom.ContainerBuilder
	)

	// 1, render yaml source manually
	gd = b.Container()
	ts = TemplateOp{
		Template: `
items:
{{- range (split "," "a,b,c") }}
{{ printf "- %s" . }}
{{- end }}`,
		Path:    "Out",
		ParseAs: ptr(ParseTextAsYaml),
	}
	err = New(WithData(gd)).Execute(&ts)
	assert.NoError(t, err)
	assert.Equal(t, 3, gd.Lookup("Out.items").(dom.List).Size())

	// 2, render using template function
	gd = b.Container()
	ts = TemplateOp{
		Template: `
items:
{{ (split "," "a,b,c") | list | toYaml }}
`,
		Path:    "Out",
		ParseAs: ptr(ParseTextAsYaml),
	}
	err = New(WithData(gd)).Execute(&ts)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(gd.Lookup("Out.items").(dom.List).Items()[0].(dom.Container).Children()))

	// 3, render invalid
	gd = b.Container()
	ts = TemplateOp{
		Template: `*** this is not a YAML ***`,
		Path:     "Out",
		ParseAs:  ptr(ParseTextAsYaml),
	}
	err = New(WithData(gd)).Execute(&ts)
	assert.Error(t, err)
}

func TestExecuteTemplateOpAsInvalid(t *testing.T) {
	assert.Error(t, New().Execute(&TemplateOp{
		Template: `---\nOla: Hi`,
		Path:     "Out",
		ParseAs:  ptr(ParseTextAs("invalid")),
	}))
}

func TestExecuteTemplateOpAsFloat64(t *testing.T) {
	var (
		gd  dom.ContainerBuilder
		ts  *TemplateOp
		err error
	)
	gd = b.Container()
	ts = &TemplateOp{
		Template: `{{ maxf 1.5 3 4.5 }}`,
		Path:     "Out",
		ParseAs:  ptr(ParseTextAsFloat64),
	}
	err = New(WithData(gd)).Execute(ts)
	assert.NoError(t, err)
	assert.Equal(t, 4.5, gd.Lookup("Out").(dom.Leaf).Value())

	gd.AddValueAt("X", dom.LeafNode("Ou"))
	ts = &TemplateOp{
		Template: `XYZ`,
		Path:     "Out",
		ParseAs:  ptr(ParseTextAsFloat64),
	}
	err = New(WithData(gd)).Execute(ts)
	assert.Error(t, err)
}

func TestExecuteTemplateOpAsInt64(t *testing.T) {
	var (
		gd  dom.ContainerBuilder
		ts  *TemplateOp
		err error
	)
	gd = b.Container()
	ts = &TemplateOp{
		Template: `{{ max 1 3 5 }}`,
		Path:     "Out",
		ParseAs:  ptr(ParseTextAsInt64),
	}
	err = New(WithData(gd)).Execute(ts)
	assert.NoError(t, err)
	assert.Equal(t, int64(5), gd.Lookup("Out").(dom.Leaf).Value())

	gd.AddValueAt("X", dom.LeafNode("Ou"))
	ts = &TemplateOp{
		Template: `XYZ`,
		Path:     "Out",
		ParseAs:  ptr(ParseTextAsInt64),
	}
	err = New(WithData(gd)).Execute(ts)
	assert.Error(t, err)
}
