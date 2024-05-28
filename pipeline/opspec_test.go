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
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestOpSpecCloneWith(t *testing.T) {
	o := OpSpec{
		Set: &SetOp{
			Data: map[string]interface{}{
				"a": 1,
			},
			Path: "{{ .Path }}",
		},
		Patch: &PatchOp{
			Path: "{{ .Path3 }}",
		},
		ForEach: &ForEachOp{
			Item:   &([]string{"left", "right"}),
			Action: OpSpec{},
		},
		Template: &TemplateOp{
			WriteTo: "{{ .Path }}",
		},
		Import: &ImportOp{
			Path: "{{ .Path }}",
			Mode: ParseFileModeYaml,
		},
	}

	a := o.CloneWith(mockActCtx(b.FromMap(map[string]interface{}{
		"Path":  "root.sub2",
		"Path3": "/root/sub3",
	}))).(OpSpec)
	t.Log(a.String())
	assert.Equal(t, "root.sub2", a.Set.Path)
	assert.Equal(t, "root.sub2", a.Import.Path)
	assert.Equal(t, "/root/sub3", a.Patch.Path)
	assert.Equal(t, "root.sub2", a.Template.WriteTo)
}
