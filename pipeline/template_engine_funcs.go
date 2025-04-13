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
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/rkosegi/yaml-toolkit/analytics"
	"github.com/rkosegi/yaml-toolkit/common"
	"github.com/rkosegi/yaml-toolkit/diff"
	"github.com/rkosegi/yaml-toolkit/dom"
	"github.com/rkosegi/yaml-toolkit/props"
	"github.com/rkosegi/yaml-toolkit/utils"
)

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

// isDirFunc checks if provided path points to directory.
// Any error is swallowed and will cause function to return false.
func isDirFunc(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return false
	}
	return fi.IsDir()
}

// globFunc exposes filepath.Glob
func globFunc(pattern string) ([]string, error) {
	return filepath.Glob(pattern)
}

// mergeFilesFunc merges 0 or more files into dom.Container
func mergeFilesFunc(files []string) (dom.Container, error) {
	ds := analytics.NewDocumentSet()
	for _, f := range files {
		err := ds.AddDocumentFromFile(f, common.DefaultFileDecoderProvider(f))
		if err != nil {
			return nil, err
		}
	}
	return ds.AsOne().Merged(dom.ListsMergeAppend()), nil
}

func dom2str(c dom.Container, encFn dom.EncoderFunc) (string, error) {
	var buf strings.Builder
	err := c.Serialize(&buf, dom.DefaultNodeEncoderFn, encFn)
	return buf.String(), err
}

func dom2yamlFunc(c dom.Container) (string, error) {
	return dom2str(c, dom.DefaultYamlEncoder)
}

func dom2jsonFunc(c dom.Container) (string, error) {
	return dom2str(c, dom.DefaultJsonEncoder)
}

func dom2propertiesFunc(c dom.Container) (string, error) {
	return dom2str(c, props.EncoderFn)
}

// domDiffFunc computes difference between 2 container nodes.
// Both of nodes must be of dom.Container type, otherwise result is empty slice
func domDiffFunc(left, right dom.Node) ([]diff.Modification, error) {
	if left != nil && left.IsContainer() && left.SameAs(right) {
		return *diff.Diff(left.(dom.Container), right.(dom.Container)), nil
	}
	return []diff.Modification{}, nil
}

// urlParseQuery just delegates call to url.ParseQuery
func urlParseQuery(qry string) (url.Values, error) {
	return url.ParseQuery(qry)
}

// fileGlobFunc just delegates call to filepath.Glob
func fileGlobFunc(pattern string) ([]string, error) {
	return filepath.Glob(pattern)
}
