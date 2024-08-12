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
	"os"
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

func TestTemplateEngineRenderToYaml(t *testing.T) {
	var (
		out string
		err error
	)
	out, err = renderTemplate("{{ toYaml . }}", map[string]interface{}{
		"x": map[string]interface{}{
			"z": "abc",
		},
		"y": 25,
	}, sprig.TxtFuncMap())
	assert.NoError(t, err)
	assert.Equal(t, "x:\n  z: abc\n\"y\": 25", out)
}

func TestTemplateFuncIsEmpty(t *testing.T) {
	type testCase struct {
		v   interface{}
		res bool
	}
	for _, v := range []testCase{
		{
			v:   "",
			res: true,
		},
		{
			v:   nil,
			res: true,
		},
		{
			v:   "a",
			res: false,
		},
		{
			v:   struct{}{},
			res: false,
		},
	} {
		assert.Equal(t, v.res, isEmptyFunc(v.v))
	}
}

func TestTemplateFuncUnflatten(t *testing.T) {
	r := unflattenFunc(map[string]interface{}{
		"a.b": 1,
		"c":   "hello",
	})
	assert.Equal(t, 2, len(r))
	assert.Equal(t, 1, r["a"].(map[string]interface{})["b"])
	assert.Equal(t, "hello", r["c"])
}

func TestTemplateFuncFileExists(t *testing.T) {
	assert.False(t, fileExistsFunc("/this/definitely/shouldn't exists"))
	f, err := os.CreateTemp("", "yt*.txt")
	assert.NoError(t, err)
	if err != nil {
		return
	}
	t.Cleanup(func() {
		t.Logf("cleanup temporary file %s", f.Name())
		_ = os.Remove(f.Name())
	})
	assert.True(t, fileExistsFunc(f.Name()))
}

func TestTemplateFuncMergeFiles(t *testing.T) {
	f1, err := os.CreateTemp("", "yt*.yaml")
	assert.NoError(t, err)
	assert.NoError(t, os.WriteFile(f1.Name(), []byte("A: 1"), 0o664))
	f2, err := os.CreateTemp("", "yt*.json")
	assert.NoError(t, err)
	assert.NoError(t, os.WriteFile(f2.Name(), []byte("{ \"B\": 2 }"), 0o664))
	res, err := mergeFilesFunc(f1.Name(), f2.Name())
	assert.NoError(t, err)
	assert.NotNil(t, res)
	t.Cleanup(func() {
		t.Logf("cleanup temporary file %s", f1.Name())
		_ = os.Remove(f1.Name())
		t.Logf("cleanup temporary file %s", f2.Name())
		_ = os.Remove(f2.Name())
	})
}

func TestTemplateFuncMergeFilesInvalid(t *testing.T) {
	f2, err := os.CreateTemp("", "yt*.json")
	assert.NoError(t, err)
	assert.NoError(t, os.WriteFile(f2.Name(), []byte("NOT_A_JSON"), 0o664))
	res, err := mergeFilesFunc(f2.Name())
	assert.Error(t, err)
	assert.Nil(t, res)
	t.Cleanup(func() {
		t.Logf("cleanup temporary file %s", f2.Name())
		_ = os.Remove(f2.Name())
	})
}

func TestTemplateFuncIsDir(t *testing.T) {
	d, err := os.MkdirTemp("", "yt*")
	assert.NoError(t, err)
	t.Cleanup(func() {
		t.Logf("deleting temporary directory %s", d)
		_ = os.RemoveAll(d)
	})
	assert.True(t, isDirFunc(d))
	assert.False(t, isDirFunc("/i hope/this/path/does/not/exist"))
}
