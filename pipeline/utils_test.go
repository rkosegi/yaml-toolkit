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
	"os"
	"regexp"
	"testing"
)

func strPointer(str string) *string {
	return &str
}

func mockEmptyActCtx() ActionContext {
	return mockActCtx(b.Container())
}

func mockActCtx(data dom.ContainerBuilder) ActionContext {
	return New(WithData(data)).(*exec).newCtx(nil)
}

func removeFilesLater(t *testing.T, files ...*os.File) {
	t.Cleanup(func() {
		for _, f := range files {
			t.Logf("cleanup temporary file %s", f.Name())
			_ = os.Remove(f.Name())
		}
	})
}

func removeDirsLater(t *testing.T, dirs ...string) {
	t.Cleanup(func() {
		for _, f := range dirs {
			t.Logf("delete temporary directory %s", f)
			_ = os.RemoveAll(f)
		}
	})
}

func TestGetActionFromContext(t *testing.T) {
	ac := mockEmptyActCtx().(*actContext)
	ac.c = &ExportOp{}
	assert.NotNil(t, ac.Action())
}

func TestNonEmpty(t *testing.T) {
	assert.False(t, nonEmpty(strPointer("")))
	assert.False(t, nonEmpty(nil))
	assert.True(t, nonEmpty(strPointer("abcd")))
}

func TestSafeStrDeref(t *testing.T) {
	assert.Equal(t, "", safeStrDeref(nil))
	assert.Equal(t, "aa", safeStrDeref(strPointer("aa")))
}

func TestSafeStrListSize(t *testing.T) {
	assert.Equal(t, 0, safeStrListSize(nil))
	assert.Equal(t, 1, safeStrListSize(&([]string{"a"})))
}

func TestSafeRegexpDeref(t *testing.T) {
	assert.Equal(t, "", safeRegexpDeref(nil))
	assert.Equal(t, "abc", safeRegexpDeref(regexp.MustCompile(`abc`)))
}

func TestSafeRenderStrPointerNil(t *testing.T) {
	assert.Nil(t, safeRenderStrPointer(nil, mockEmptyActCtx().TemplateEngine(), nil))
}

func TestSafeRenderStrPointer(t *testing.T) {
	s := "{{ .X }}"
	d := b.FromMap(map[string]interface{}{
		"X": "abc",
	})
	c := mockActCtx(d)
	assert.Equal(t, "abc", *safeRenderStrPointer(&s, c.TemplateEngine(), c.Snapshot()))
}

func TestSafeCopyIntSlice(t *testing.T) {
	var x *[]int
	assert.Nil(t, safeCopyIntSlice(nil))
	a := []int{1, 2, 5}
	x = safeCopyIntSlice(&a)
	assert.Equal(t, 3, len(*x))
}
