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

func TestLogOpDo(t *testing.T) {
	eo := &LogOp{}
	assert.NoError(t, eo.Do(mockEmptyActCtx()))
}

func TestLogOpCloneWith(t *testing.T) {
	eo := &LogOp{
		Message: "Output format: {{ .Format }}",
	}
	assert.Contains(t, eo.String(), "Log[")
	d := b.Container()
	d.AddValue("Format", dom.LeafNode("toml"))
	eo = eo.CloneWith(newMockActBuilder().data(d).build()).(*LogOp)
	assert.Equal(t, "Output format: toml", eo.Message)
}
