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

package jmespath

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQuery(t *testing.T) {
	p := NewParser()
	q, err := p.Parse("a")
	assert.NoError(t, err)
	assert.NotNil(t, q)
	r := q.Select(map[string]interface{}{
		"a": []interface{}{
			"b", "d",
		},
	})
	assert.Len(t, r, 2)
	assert.Equal(t, "b", r[0])

	r = q.Select(map[string]interface{}{
		"a": "b",
	})
	assert.Len(t, r, 1)
	assert.Equal(t, "b", r[0])

	_, err = p.Parse("syntax error")
	assert.Error(t, err)
}
