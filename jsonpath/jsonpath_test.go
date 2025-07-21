/*
Copyright 2025 Richard Kosegi

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

package jsonpath

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQuery(t *testing.T) {
	p := NewParser()
	q, err := p.Parse("$.a.b")
	assert.NoError(t, err)
	assert.NotNil(t, q)
	r := q.Select(map[string]interface{}{
		"a": map[string]interface{}{
			"b": "c",
			"d": 1,
		},
	})
	assert.Len(t, r, 1)
	assert.Equal(t, "c", r[0])

	_, err = p.Parse("syntax error")
	assert.Error(t, err)
}
