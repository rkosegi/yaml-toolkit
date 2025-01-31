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
	"os"
	"testing"

	sprig "github.com/go-task/slim-sprig/v3"
	"github.com/stretchr/testify/assert"
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
	removeFilesLater(t, f)
	assert.True(t, fileExistsFunc(f.Name()))
}

func TestTemplateFuncMergeFiles(t *testing.T) {
	f1, err := os.CreateTemp("", "yt*.yaml")
	assert.NoError(t, err)
	assert.NoError(t, os.WriteFile(f1.Name(), []byte("A: 1"), 0o664))
	f2, err := os.CreateTemp("", "yt*.json")
	assert.NoError(t, err)
	assert.NoError(t, os.WriteFile(f2.Name(), []byte("{ \"B\": 2 }"), 0o664))
	res, err := mergeFilesFunc([]string{f1.Name(), f2.Name()})
	assert.NoError(t, err)
	assert.NotNil(t, res)
	removeFilesLater(t, f1, f2)
}

func TestTemplateFuncMergeFilesInvalid(t *testing.T) {
	f2, err := os.CreateTemp("", "yt*.json")
	assert.NoError(t, err)
	assert.NoError(t, os.WriteFile(f2.Name(), []byte("NOT_A_JSON"), 0o664))
	res, err := mergeFilesFunc([]string{f2.Name()})
	assert.Error(t, err)
	assert.Nil(t, res)
	removeFilesLater(t, f2)
}

func TestTemplateFuncIsDir(t *testing.T) {
	d, err := os.MkdirTemp("", "yt*")
	assert.NoError(t, err)
	removeDirsLater(t, d)
	assert.True(t, isDirFunc(d))
	assert.False(t, isDirFunc("/i hope/this/path/does/not/exist"))
}

func TestTemplateFuncGlob(t *testing.T) {
	d, err := os.MkdirTemp("", "yt*")
	assert.NoError(t, err)
	assert.NoError(t, os.WriteFile(d+"/1.yaml", []byte{}, 0o664))
	removeDirsLater(t, d)
	files, err := globFunc(d + "/*.yaml")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(files))
}

func TestTemplateFuncDom2Yaml(t *testing.T) {
	type testCase struct {
		format string
		exp    string
	}
	d, err := os.MkdirTemp("", "yt*")
	assert.NoError(t, err)
	removeDirsLater(t, d)
	f1, err := os.CreateTemp(d, "yt*.yaml")
	assert.NoError(t, err)
	f2, err := os.CreateTemp(d, "yt*.yaml")
	assert.NoError(t, err)
	_, err = f1.Write([]byte("a: 1"))
	assert.NoError(t, err)
	_, err = f2.Write([]byte("b: 2\n"))
	assert.NoError(t, err)
	for _, tc := range []testCase{
		{
			format: "properties",
			exp:    "a=1",
		},
		{
			format: "yaml",
			exp:    "a: 1",
		},
		{
			format: "json",
			exp:    `"a": 1`,
		},
	} {
		var res string
		res, err = renderTemplate(`{{ mergeFiles ( glob ( printf "%s/*.yaml" .Temp ) ) | dom2`+tc.format+` | trim }}`,
			map[string]interface{}{
				"Temp": d,
			}, sprig.HermeticTxtFuncMap())
		t.Logf("Merged content using format '%s':\n%s", tc.format, res)
		assert.NoError(t, err)
		assert.Contains(t, res, tc.exp)
	}
	removeFilesLater(t, f1, f2)
}
