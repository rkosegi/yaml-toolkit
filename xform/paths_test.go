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

package xform

import (
	"testing"

	"github.com/rkosegi/yaml-toolkit/props"
	"github.com/stretchr/testify/assert"
)

func TestPropPath2Pointer(t *testing.T) {
	xp := PropPath2Pointer(props.ParsePath("root.sub1.list1[1][2].sub2.list2[3]"))
	assert.Equal(t, "/root/sub1/list1/1/2/sub2/list2/3", xp.String())
}
