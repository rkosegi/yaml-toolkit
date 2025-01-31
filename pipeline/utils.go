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
	"regexp"
	"slices"
	"strings"

	"github.com/rkosegi/yaml-toolkit/dom"
)

func strTruncIfNeeded(in string, size int) string {
	if len(in) <= size {
		return in
	}
	return in[0:size]
}

func parseFile(path string, mode ParseFileMode) (dom.Node, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(mode) == 0 {
		mode = ParseFileModeText
	}
	val, err := mode.toValue(data)
	if err != nil {
		return nil, err
	}
	return val, nil
}

func safeStrDeref(in *string) string {
	if in == nil {
		return ""
	}
	return *in
}

func setStrategyPointer(s SetStrategy) *SetStrategy {
	return &s
}

func safeRegexpDeref(re *regexp.Regexp) string {
	if re == nil {
		return ""
	}
	return re.String()
}

func safeStrListSize(in *[]string) int {
	if in == nil {
		return 0
	}
	return len(*in)
}

func nonEmpty(in *string) bool {
	return in != nil && len(*in) > 0
}

func sortActionNames(actions ChildActions) []string {
	var keys []string
	for n := range actions {
		keys = append(keys, n)
	}
	slices.SortFunc(keys, func(a, b string) int {
		return actions[a].Order - actions[b].Order
	})
	return keys
}

func actionNames(actions ChildActions) string {
	return strings.Join(sortActionNames(actions), ",")
}

func safeCopyIntSlice(in *[]int) *[]int {
	if in == nil {
		return nil
	}
	r := make([]int, len(*in))
	copy(r, *in)
	return &r
}

func safeRenderStrPointer(str *string, te TemplateEngine, data map[string]interface{}) *string {
	if str == nil {
		return nil
	}
	s := te.RenderLenient(*str, data)
	return &s
}

func safeRenderStrSlice(args *[]string, te TemplateEngine, data map[string]interface{}) *[]string {
	if args == nil {
		return nil
	}
	r := make([]string, len(*args))
	for i, arg := range *args {
		r[i] = te.RenderLenient(arg, data)
	}
	return &r
}
