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

package common

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultFileEncoderProvider(t *testing.T) {
	for _, ext := range []string{"a.yaml", "b.yml", "c.json", "d.properties"} {
		t.Log("file:", ext)
		assert.NotNil(t, DefaultFileEncoderProvider(ext))
	}
	assert.Nil(t, DefaultFileEncoderProvider(".unknown"))
}

func TestDefaultFileDecoderProvider(t *testing.T) {
	for _, ext := range []string{"a.yaml", "b.yml", "c.json", "d.properties"} {
		t.Log("file:", ext)
		assert.NotNil(t, DefaultFileDecoderProvider(ext))
	}
	assert.Nil(t, DefaultFileDecoderProvider(".unknown"))
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
