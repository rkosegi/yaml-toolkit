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

package pipeline

import (
	"testing"

	sprig "github.com/go-task/slim-sprig/v3"
	"github.com/stretchr/testify/assert"
)

func TestSortActionNames(t *testing.T) {
	ca := ChildActions{
		"a": ActionSpec{ActionMeta: ActionMeta{
			Order: 90,
		}},
		"b": ActionSpec{ActionMeta: ActionMeta{
			Order: 50,
		}},
		"c": ActionSpec{ActionMeta: ActionMeta{
			Order: 10,
		},
		},
	}
	assert.Equal(t, "c,b,a", actionNames(ca))
	assert.Contains(t, ca.String(), "c,b,a")
}

func TestChildActionsCloneWith(t *testing.T) {
	a := ChildActions{
		"step1": ActionSpec{
			Operations: OpSpec{
				Set: &SetOp{
					Data: map[string]interface{}{
						"abcd": 123,
					},
					Path:     "{{ .sub1.leaf1 }}",
					Strategy: setStrategyPointer(SetStrategyReplace),
				},
			},
		},
	}.CloneWith(&actContext{
		exec: &exec{
			d: b.FromMap(map[string]interface{}{
				"sub1": map[string]interface{}{
					"leaf1": "root.sub2",
				},
			}), t: &templateEngine{fm: sprig.TxtFuncMap()},
		},
	})
	assert.NotNil(t, a)
	assert.Equal(t, "root.sub2", a.(ChildActions)["step1"].Operations.Set.Path)
}
