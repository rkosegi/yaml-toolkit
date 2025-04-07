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

func TestActionMetaString(t *testing.T) {
	type testcase struct {
		am  *ActionMeta
		exp string
	}

	for _, tc := range []testcase{
		{
			am: &ActionMeta{
				Name: "test",
			},
			exp: "[name=test]",
		},
		{
			am: &ActionMeta{
				Order: 1,
			},
			exp: "[order=1]",
		},
		{
			am: &ActionMeta{
				When: strPointer("{{ false }}"),
			},
			exp: "[when={{ false }}]",
		},
	} {
		assert.Equal(t, tc.exp, tc.am.String())
	}
}
