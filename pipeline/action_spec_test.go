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

	"github.com/stretchr/testify/assert"
)

func TestActionSpecShouldPropagateError(t *testing.T) {
	var as *ActionSpec
	as = &ActionSpec{}
	assert.Nil(t, as.ErrorPropagation)
	as = &ActionSpec{
		ActionMeta: ActionMeta{
			ErrorPropagation: ptr(ErrorPropagationPolicy("unknown")),
		},
	}
	assert.True(t, as.shouldPropagateError())
	as = &ActionSpec{
		ActionMeta: ActionMeta{
			ErrorPropagation: ptr(ErrorPropagationPolicyIgnore),
		},
	}
	assert.False(t, as.shouldPropagateError())

	as = &ActionSpec{
		ActionMeta: ActionMeta{
			ErrorPropagation: ptr(ErrorPropagationPolicyIgnore),
		},
		Operations: OpSpec{
			Abort: &AbortOp{Message: "abort"},
		},
	}
	assert.NoError(t, as.Do(newMockActBuilder().build()))
}
