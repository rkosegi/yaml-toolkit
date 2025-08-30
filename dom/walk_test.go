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

package dom

import (
	"testing"

	"github.com/rkosegi/yaml-toolkit/path"
	"github.com/stretchr/testify/assert"
)

func TestWalkDfs(t *testing.T) {
	c := getTestDoc(t, "doc1")
	t.Run("walk everything except list", func(t *testing.T) {
		cnt := 0
		walkStart(func(p path.Path, parent Node, node Node) bool {
			if parent.IsList() {
				return false
			}
			cnt++
			return true
		}, c, WalkOptDFS())
		assert.Equal(t, 6, cnt)
	})

	t.Run("walk only one level", func(t *testing.T) {
		cnt := 0
		walkStart(func(p path.Path, parent Node, node Node) bool {
			t.Log(p.String())
			if len(p.Components()) > 1 {
				return false
			}
			cnt++
			return true
		}, c, WalkOptDFS())
		assert.Equal(t, 1, cnt)
	})
}
