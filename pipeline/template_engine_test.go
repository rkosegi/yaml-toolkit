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
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPossiblyTemplate(t *testing.T) {
	assert.True(t, possiblyTemplate("{{ . }}"))
	assert.True(t, possiblyTemplate("{{data}}"))
	assert.True(t, possiblyTemplate("{{}}"))
	assert.False(t, possiblyTemplate("{{"))
	assert.False(t, possiblyTemplate("345678"))
}

func TestTemplateEngineRenderLenient(t *testing.T) {
	te := &templateEngine{
		fm: sprig.TxtFuncMap(),
	}
	assert.Equal(t, "AAA", te.RenderLenient("AAA", nil))
	assert.Equal(t, "{{ data }}", te.RenderLenient("{{ data }}", nil))
	assert.Equal(t, "123", te.RenderLenient("{{ .data }}", map[string]interface{}{
		"data": 123,
	}))
}

func TestRenderTemplate(t *testing.T) {
	var (
		out string
		err error
	)
	// invalid template syntax
	_, err = renderTemplate("{{", map[string]interface{}{}, sprig.TxtFuncMap())
	assert.Error(t, err)

	// valid template, valid data
	out, err = renderTemplate("{{ .X }}", map[string]interface{}{
		"X": "abcd",
	}, sprig.TxtFuncMap())
	assert.NoError(t, err)
	assert.Equal(t, "abcd", out)

	// invalid data
	_, err = renderTemplate("{{ .a }}", "", sprig.TxtFuncMap())
	assert.Error(t, err)
}

func TestTemplateEngineRenderTpl(t *testing.T) {
	var (
		out string
		err error
	)
	out, err = renderTemplate("{{ tpl .T . }}", map[string]interface{}{
		"T": "{{ add .X 3 }}",
		"X": 10,
	}, sprig.TxtFuncMap())
	assert.NoError(t, err)
	assert.Equal(t, "13", out)
}

func TestTemplateEngineRenderTplInvalid(t *testing.T) {
	var (
		err error
	)
	_, err = renderTemplate("{{ tpl .T . }}", map[string]interface{}{
		"T": "{{",
	}, sprig.TxtFuncMap())
	assert.Error(t, err)
}
