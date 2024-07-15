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

func TestExecuteSetOp(t *testing.T) {
	var (
		ss  SetOp
		gd  dom.ContainerBuilder
		err error
	)
	ss = SetOp{
		Data: map[string]interface{}{
			"sub1": 123,
		},
	}
	gd = b.Container()
	assert.NoError(t, New(WithData(gd)).Execute(&ss))
	assert.Equal(t, 123, gd.Lookup("sub1").(dom.Leaf).Value())

	ss = SetOp{
		Data: map[string]interface{}{
			"sub1": 123,
		},
		Path: "sub0",
	}
	gd = b.Container()
	assert.NoError(t, New(WithData(gd)).Execute(&ss))
	assert.Equal(t, 123, gd.Lookup("sub0.sub1").(dom.Leaf).Value())
	assert.Contains(t, ss.String(), "sub0")

	ss = SetOp{}
	err = New(WithData(gd)).Execute(&ss)
	assert.Error(t, err)
	assert.Equal(t, ErrNoDataToSet, err)
}
