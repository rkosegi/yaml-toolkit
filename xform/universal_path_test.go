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

func TestParseUniversalPath(t *testing.T) {
	type data struct {
		P *UniversalPath `yaml:"path"`
	}
	var (
		str string
		err error
		d   data
	)
	str = `
---
path:
  - a
`
	err = yaml.Unmarshal([]byte(str), &d)
	assert.Error(t, err)

	str = `
---
path:
  value: *x
`
	err = yaml.Unmarshal([]byte(str), &d)
	assert.Error(t, err)

	str = `
---
path: root.sub1.sublist[0].sub3
`
	err = yaml.Unmarshal([]byte(str), &d)
	assert.NoError(t, err)
	assert.NotNil(t, d.P.Value)
	assert.Equal(t, 5, len(d.P.Value.Components()))
	assert.Equal(t, "sub1", d.P.Value.Components()[1].Value())

	str = `
---
path:
  value: invalid
  syntax: rfc6901
`
	err = yaml.Unmarshal([]byte(str), &d)
	assert.Error(t, err)

	str = `
---
path:
  syntax: rfc6901
`
	err = yaml.Unmarshal([]byte(str), &d)
	assert.Error(t, err)

	str = `
---
path:
  value: 1
  syntax: rfc6901
`
	err = yaml.Unmarshal([]byte(str), &d)
	assert.Error(t, err)

	str = `
---
path:
  value: root
  syntax: unknown
`
	err = yaml.Unmarshal([]byte(str), &d)
	assert.Error(t, err)
}
