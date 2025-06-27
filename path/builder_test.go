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
		Append(Numeric(1)).
		Append(AfterLast())
	p := b.Build()

	pc := p.Components()

	assert.False(t, p.IsEmpty())
	assert.Equal(t, 3, len(pc))
	assert.Equal(t, "root", pc[0].Value())
	assert.True(t, pc[1].IsNumeric())
	assert.Equal(t, "1", pc[1].Value())
	assert.Equal(t, 1, pc[1].NumericValue())
	assert.True(t, pc[2].IsInsertAfterLast())
	assert.False(t, p.Last().IsNumeric())

}

func TestPathGetLastEmpty(t *testing.T) {
	defer func() {
		recover()
	}()
	NewBuilder().Build().Last()
	assert.Fail(t, "should not be here")
}

func TestParentOf(t *testing.T) {
	// parent of empty path is nil
	assert.Nil(t, ParentOf(NewBuilder().Build()))
	assert.Len(t, ParentOf(NewBuilder().Append(Simple("a")).
		Append(Simple("b")).Build()).Components(), 1)
	assert.Equal(t, emptyPath, ParentOf(NewBuilder().Append(Simple("A")).Build()))
}

func TestChildOf(t *testing.T) {
	np := ChildOf(NewBuilder().Append(Simple("a")).Build(), Simple("b"), Simple("c"))
	assert.Len(t, np.Components(), 3)
	assert.Equal(t, "c", np.Last().Value())
}
