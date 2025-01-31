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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParsePath(t *testing.T) {
	var p Path
	p = ParsePath("a.b.c")
	assert.Equal(t, 3, len(p))
	assert.Equal(t, "a", p[0].String())
	p = ParsePath("x[1].b.c[3]")
	assert.Equal(t, 5, len(p))
	assert.Equal(t, "x", p[0].String())
	assert.Equal(t, "1", p[1].String())
	assert.Equal(t, "b", p[2].String())
}
