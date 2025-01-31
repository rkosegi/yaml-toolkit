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

package path

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuilder(t *testing.T) {
	b := NewBuilder().
		Append(Simple("root")).
		Append(Wildcard()).
		Append(Numeric(1)).
		Append(AfterLast())
	p := b.Build()

	pc := p.Components()

	assert.False(t, p.IsEmpty())
	assert.Equal(t, 4, len(pc))
	assert.Equal(t, "root", pc[0].Value())
	assert.True(t, pc[1].IsWildcard())
	assert.True(t, pc[2].IsNumeric())
	assert.Equal(t, "1", pc[2].Value())
	assert.Equal(t, 1, pc[2].NumericValue())
	assert.True(t, pc[3].IsInsertAfterLast())
	assert.False(t, p.Last().IsNumeric())

	b.Reset()
	assert.True(t, b.Build().IsEmpty())
}

func TestBuilderAppendNoOption(t *testing.T) {
	defer func() {
		recover()
	}()
	BuildComponent()
	assert.Fail(t, "should not be here")
}

func TestPathGetLastEmpty(t *testing.T) {
	defer func() {
		recover()
	}()
	NewBuilder().Build().Last()
	assert.Fail(t, "should not be here")
}
