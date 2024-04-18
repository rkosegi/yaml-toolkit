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
	"github.com/rkosegi/yaml-toolkit/utils"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
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
	assert.Error(t, EncoderFn(utils.FailingWriter(), m))
}

func TestDecoderFn(t *testing.T) {
	m := make(map[string]interface{})
	err := DecoderFn(strings.NewReader("a.b.c=1\nx.y.z=Hi!\n"), m)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(m))
	assert.Contains(t, m, "a.b.c")
	assert.Contains(t, m, "x.y.z")

	err = DecoderFn(utils.FailingReader(), m)
	assert.Error(t, err)
}
