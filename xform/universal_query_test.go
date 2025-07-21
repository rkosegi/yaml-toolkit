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

package xform

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestParseUniversalQuery(t *testing.T) {
	type data struct {
		P *UniversalQuery `yaml:"query"`
	}
	var (
		str string
		err error
		d   data
	)
	str = ``
	err = yaml.Unmarshal([]byte(str), &d)
	assert.NoError(t, err)
	assert.Nil(t, d.P)

	str = `
query: $.a.b
`
	err = yaml.Unmarshal([]byte(str), &d)
	assert.NoError(t, err)
	assert.Equal(t, "rfc9535", string(d.P.Syntax))

	for _, inv := range []string{
		`query: invalid`, `
query:
  value: ok
  syntax: unknown`, `
query:
  value: 1.3
`,
		`
query:
  novalue:`,
		`
---
query:
  value: *x
`,
		`
query:
  - 1`,
	} {
		err = yaml.Unmarshal([]byte(inv), &d)
		t.Logf("err = %v", err)
		assert.Error(t, err)
	}

}
