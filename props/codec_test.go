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

package props

import (
	"bytes"
	"strings"
	"testing"

	"github.com/rkosegi/yaml-toolkit/common"
	"github.com/rkosegi/yaml-toolkit/dom"
	"github.com/stretchr/testify/assert"
)

func TestEncoderFn(t *testing.T) {
	m := map[string]interface{}{
		"a.b.c": 1,
		"x.y.z": "Hi!",
	}
	var buff bytes.Buffer
	err := EncoderFn(&buff, m)

	assert.NoError(t, err)
	assert.Contains(t, buff.String(), "a.b.c=1\n")
	assert.Contains(t, buff.String(), "x.y.z=Hi!\n")
	assert.Error(t, EncoderFn(common.FailingWriter(), m))
}

func TestDomEncoderFn(t *testing.T) {
	m := map[string]interface{}{
		"a.b.c": 1,
		"x.y.z": "Hi!",
	}
	var buff bytes.Buffer
	c := dom.Builder().FromMap(m)
	err := DomEncoderFn(&buff, c)
	assert.NoError(t, err)
	assert.Contains(t, buff.String(), "a.b.c=1\n")
	assert.Contains(t, buff.String(), "x.y.z=Hi!\n")
	assert.Error(t, DomEncoderFn(common.FailingWriter(), c))
}

func TestDecoderFn(t *testing.T) {
	m := make(map[string]interface{})
	err := DecoderFn(strings.NewReader("a.b=1\nx.y=Hi!\n"), &m)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(m))
	assert.Equal(t, "1", m["a"].(map[string]interface{})["b"])
	assert.Equal(t, "Hi!", m["x"].(map[string]interface{})["y"])

	err = DecoderFn(common.FailingReader(), m)
	assert.Error(t, err)
}

func TestParseListPathComponent(t *testing.T) {
	name, indexes, ok := ParseListPathComponent("list[0][3][1]")
	assert.True(t, ok)
	assert.Equal(t, 3, len(indexes))
	assert.Equal(t, 0, indexes[0])
	assert.Equal(t, 3, indexes[1])
	assert.Equal(t, 1, indexes[2])
	assert.Equal(t, "list", name)
	_, _, ok = ParseListPathComponent("not a list")
	assert.False(t, ok)
}
