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
	"github.com/rkosegi/yaml-toolkit/dom"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestExecuteTemplateOp(t *testing.T) {
	var (
		err error
		ts  TemplateOp
		gd  dom.ContainerBuilder
	)

	gd = b.Container()
	gd.AddValueAt("root.leaf1", dom.LeafNode(123456))
	ts = TemplateOp{
		Template: `{{ (mul .Data.root.leaf1 2) | quote }}`,
		Path:     "result.x1",
	}
	assert.NoError(t, New(WithData(gd)).Execute(&ts))
	assert.Equal(t, "\"246912\"", gd.Lookup("result.x1").(dom.Leaf).Value())
	assert.Contains(t, ts.String(), "result.x1")

	// empty template error
	ts = TemplateOp{}
	err = New(WithData(gd)).Execute(&ts)
	assert.Error(t, err)
	assert.Equal(t, ErrTemplateEmpty, err)

	// empty path error
	ts = TemplateOp{
		Template: `TEST`,
	}
	err = New(WithData(gd)).Execute(&ts)
	assert.Error(t, err)
	assert.Equal(t, ErrPathEmpty, err)

	ts = TemplateOp{
		Template: `{{}}{{`,
		Path:     "result",
	}
	assert.Error(t, New(WithData(gd)).Execute(&ts))

	ts = TemplateOp{
		Template: `{{ invalid_func }}`,
		Path:     "result",
	}
	assert.Error(t, New(WithData(gd)).Execute(&ts))
}
