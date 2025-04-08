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

package pipeline

import (
	"testing"

	"github.com/rkosegi/yaml-toolkit/dom"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestAnyVal(t *testing.T) {
	type testStruct struct {
		V *AnyVal `yaml:"v,omitempty"`
	}
	var (
		err error
		ts  testStruct
		// n  *yaml.Node
	)
	d1 := `v: 123`
	err = yaml.Unmarshal([]byte(d1), &ts)
	assert.NoError(t, err)
	assert.Equal(t, "123", ts.V.Value().(dom.Leaf).Value())
	d2 := `
v:
  a: 123
  b: XYZ
`
	err = yaml.Unmarshal([]byte(d2), &ts)
	assert.NoError(t, err)
	assert.Equal(t, "123", ts.V.Value().(dom.Container).Child("a").(dom.Leaf).Value())
}
