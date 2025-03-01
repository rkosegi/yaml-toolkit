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

	"github.com/rkosegi/yaml-toolkit/dom"
	"github.com/stretchr/testify/assert"
)

type noopOp struct {
}

func (n *noopOp) String() string                   { return "" }
func (n *noopOp) Do(_ ActionContext) error         { return nil }
func (n *noopOp) CloneWith(_ ActionContext) Action { return &noopOp{} }

func TestExtOpDo(t *testing.T) {
	var ex *ExtOp
	d := b.Container()
	ctx := mockActCtxExt(d, map[string]Action{
		"dummyfn": &SetOp{
			Data: map[string]interface{}{
				"X": 123,
			},
		},
	})
	ex = &ExtOp{
		Function: "dummyfn",
	}
	assert.NoError(t, ex.Do(ctx))
	assert.Equal(t, 123, d.Lookup("X").(dom.Leaf).Value())
	ex = &ExtOp{
		Function: "non-existent",
	}
	assert.Error(t, ex.Do(ctx))

}
