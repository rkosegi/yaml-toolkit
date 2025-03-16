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
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/rkosegi/yaml-toolkit/dom"
	"github.com/stretchr/testify/assert"
)

type testingLogger struct {
	t      *testing.T
	indent int
}

func (l *testingLogger) OnLog(_ ActionContext, v ...interface{}) {
	l.t.Logf("%s[LOG] %v\n", strings.Repeat(" ", l.indent), v)
}

func (l *testingLogger) OnBefore(ctx ActionContext) {
	l.t.Logf("%s[START] %v\n", strings.Repeat(" ", l.indent), ctx.Action().String())
	l.indent++
}

func (l *testingLogger) OnAfter(ctx ActionContext, err error) {
	l.indent--
	l.t.Logf("%s[END] %v\n", strings.Repeat(" ", l.indent), ctx.Action().String())
	if err != nil {
		l.t.Logf("%s[ERROR] %v\n", strings.Repeat(" ", l.indent), err)
	}
}

func strPointer(str string) *string {
	return &str
}

func newMockActBuilder() *mockActCtxBuilder {
	return &mockActCtxBuilder{d: b.Container()}
}

type mockActCtxBuilder struct {
	d    dom.ContainerBuilder
	opts []Opt
}

func (mcb *mockActCtxBuilder) ext(ea map[string]Action) *mockActCtxBuilder {
	mcb.opts = append(mcb.opts, WithExtActions(ea))
	return mcb
}

func (mcb *mockActCtxBuilder) data(d dom.ContainerBuilder) *mockActCtxBuilder {
	mcb.d = d
	return mcb
}

func (mcb *mockActCtxBuilder) testLogger(t *testing.T) *mockActCtxBuilder {
	mcb.opts = append(mcb.opts, WithListener(&testingLogger{t: t}))
	return mcb
}

func (mcb *mockActCtxBuilder) build() ActionContext {
	mcb.opts = append(mcb.opts, WithData(mcb.d))
	return New(mcb.opts...).(*exec).newCtx(nil)
}

func mockEmptyActCtx() ActionContext {
	return newMockActBuilder().build()
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
	c := newMockActBuilder().data(d).build()
	assert.Equal(t, "abc", *safeRenderStrPointer(&s, c.TemplateEngine(), c.Snapshot()))
}

func TestStrTruncIfNeeded(t *testing.T) {
	assert.Equal(t, "Hello", strTruncIfNeeded("Hello world", 5))
	assert.Equal(t, "Hell", strTruncIfNeeded("Hell", 5))
}

func TestSafeCopyIntSlice(t *testing.T) {
	var x *[]int
	assert.Nil(t, safeCopyIntSlice(nil))
	a := []int{1, 2, 5}
	x = safeCopyIntSlice(&a)
	assert.Equal(t, 3, len(*x))
}

func TestSafeBoolDeref(t *testing.T) {
	assert.False(t, safeBoolDeref(nil))
	assert.True(t, safeBoolDeref(ptr(true)))
}
