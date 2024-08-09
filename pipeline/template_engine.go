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
	"bytes"
	"fmt"
	"github.com/rkosegi/yaml-toolkit/analytics"
	"github.com/rkosegi/yaml-toolkit/dom"
	"github.com/rkosegi/yaml-toolkit/utils"
	"os"
	"strconv"
	"strings"
	"text/template"
)

type templateEngine struct {
	fm template.FuncMap
}

func tplFunc(tmpl *template.Template) func(string, interface{}) (string, error) {
	return func(tpl string, data interface{}) (string, error) {
		t, _ := tmpl.Clone()
		t, err := tmpl.New(tmpl.Name()).Parse(tpl)
		if err != nil {
			return "", err
		}
		var buf strings.Builder
		err = t.Execute(&buf, data)
		return buf.String(), err
	}
}

func toYamlFunc(v interface{}) (string, error) {
	var buf strings.Builder
	err := utils.NewYamlEncoder(&buf).Encode(v)
	return strings.TrimSuffix(buf.String(), "\n"), err
}

// isEmptyFunc returns true if given argument is nil, or empty string
func isEmptyFunc(v interface{}) bool {
	if v == nil {
		return true
	}
	if str, ok := v.(string); ok {
		if str == "" {
			return true
		}
	}
	return false
}

// un-flatten map
func unflattenFunc(v map[string]interface{}) map[string]interface{} {
	return utils.Unflatten(v)
}

// fileExistsFunc checks if files exists.
// Any error is swallowed and will cause function to return false, as if file does not exist.
func fileExistsFunc(f string) bool {
	_, err := os.Stat(f)
	if err != nil {
		return false
	}
	return true
}

// mergeFilesFunc merges 0 or more files into single map[string]interface{}
func mergeFilesFunc(files ...string) (map[string]interface{}, error) {
	ds := analytics.NewDocumentSet()
	result := make(map[string]interface{})
	for _, f := range files {
		err := ds.AddDocumentFromFile(f, analytics.DefaultFileDecoderProvider(f))
		if err != nil {
			return nil, err
		}
	}
	for k, v := range ds.AsOne().Merged(dom.ListsMergeAppend()).Flatten() {
		result[k] = fmt.Sprintf("%v", v.Value())
	}
	return result, nil
}

func renderTemplate(tmplStr string, data interface{}, fm template.FuncMap) (string, error) {
	tmpl := template.New("tmpl").Funcs(fm)
	tmpl.Funcs(template.FuncMap{
		"tpl":        tplFunc(tmpl),
		"toYaml":     toYamlFunc,
		"isEmpty":    isEmptyFunc,
		"unflatten":  unflattenFunc,
		"fileExists": fileExistsFunc,
		"mergeFiles": mergeFilesFunc,
	})
	_, err := tmpl.Parse(tmplStr)
	if err != nil {
		return "", err
	}
	var out bytes.Buffer
	err = tmpl.Execute(&out, data)
	if err != nil {
		return "", err
	}
	return out.String(), nil
}

func possiblyTemplate(in string) bool {
	openIdx := strings.Index(in, "{{")
	if openIdx == -1 {
		return false
	}
	closeIdx := strings.Index(in[openIdx:], "}}")
	return closeIdx > 0
}

func renderLenientTemplate(tmpl string, data map[string]interface{}, fm template.FuncMap) string {
	if possiblyTemplate(tmpl) {
		if val, err := renderTemplate(tmpl, data, fm); err != nil {
			return tmpl
		} else {
			return val
		}
	}
	return tmpl
}

func (te templateEngine) RenderLenient(tmpl string, data map[string]interface{}) string {
	return renderLenientTemplate(tmpl, data, te.fm)
}

func (te templateEngine) Render(tmpl string, data map[string]interface{}) (string, error) {
	return renderTemplate(tmpl, data, te.fm)
}

func (te templateEngine) EvalBool(template string, data map[string]interface{}) (bool, error) {
	val, err := te.Render(template, data)
	if err != nil {
		return false, err
	}
	return strconv.ParseBool(strings.TrimSpace(val))
}
