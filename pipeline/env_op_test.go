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
	"github.com/rkosegi/yaml-toolkit/dom"
	"github.com/stretchr/testify/assert"
	"regexp"
	"testing"
)

func createRePtr(in string) *regexp.Regexp {
	x := regexp.MustCompile(in)
	return x
}

func TestEnvOpDo(t *testing.T) {
	eo := EnvOp{
		envGetter: func() []string {
			return []string{
				"MOCK1=val1",
				"MOCK2=val2",
				"XYZ=123",
			}
		},
		Path:    "Sub",
		Include: createRePtr(`MOCK\d+`),
		Exclude: createRePtr("XYZ"),
	}
	d := b.Container()
	err := eo.Do(mockActCtx(d))
	assert.NoError(t, err)
	assert.Equal(t, "val1", d.Lookup("Sub.Env.MOCK1").(dom.Leaf).Value())
	assert.Equal(t, "val2", d.Lookup("Sub.Env.MOCK2").(dom.Leaf).Value())
	assert.Contains(t, eo.String(), "Sub")
}

func TestEnvOpCloneWith(t *testing.T) {
	eo := &EnvOp{
		Path: "{{ .NewPath }}",
	}
	d := b.Container()
	d.AddValue("NewPath", dom.LeafNode("root"))
	eo = eo.CloneWith(mockActCtx(d)).(*EnvOp)
	assert.Equal(t, "root", eo.Path)
}

func filterStrSlice(in []string, fn StringPredicateFn) []string {
	result := make([]string, 0)
	for _, e := range in {
		if fn(e) {
			result = append(result, e)
		}
	}
	return result
}

func TestStringMatchFunc(t *testing.T) {
	in := []string{"a", "b", "c"}
	var res []string
	res = filterStrSlice(in, MatchAny())
	assert.Equal(t, in, res)
	res = filterStrSlice(in, MatchNone())
	assert.Equal(t, 0, len(res))
	res = filterStrSlice(in, MatchRe(regexp.MustCompile(`a`)))
	assert.Equal(t, 1, len(res))
	res = filterStrSlice(in, MatchRe(regexp.MustCompile(`[ac]`)))
	assert.Equal(t, 2, len(res))
}
