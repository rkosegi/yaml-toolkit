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

package pipeline

import (
	"os"
	"testing"

	"github.com/rkosegi/yaml-toolkit/dom"
	"github.com/stretchr/testify/assert"
)

func TestTemplateFileOp(t *testing.T) {
	var (
		err error
		x   []byte
		f   *os.File
		tfo *TemplateFileOp
	)
	d := b.Container()
	ctx := newMockActBuilder().testLogger(t).data(d).build()
	tfo = &TemplateFileOp{}
	assert.Error(t, tfo.Do(ctx))
	tfo = &TemplateFileOp{File: "/tmp/in.tmpl"}
	assert.Error(t, tfo.Do(ctx))

	f, err = os.CreateTemp(t.TempDir(), "yt")
	t.Cleanup(func() {
		_ = os.Remove(f.Name())
	})
	assert.NoError(t, err)

	tfo = &TemplateFileOp{
		File:   "../testdata/invalid.template",
		Output: f.Name(),
	}
	assert.Error(t, tfo.Do(ctx))

	tfo = &TemplateFileOp{
		File:   "../testdata/simple.template",
		Output: f.Name(),
		Path:   ptr("invalid-path"),
	}
	assert.Error(t, tfo.Do(ctx))

	d.AddValueAt("tmpl1.name", dom.LeafNode("tester"))
	tfo = &TemplateFileOp{
		File:   "../testdata/simple.template",
		Output: f.Name(),
		Path:   ptr("tmpl1"),
	}
	assert.NoError(t, tfo.Do(ctx))
	assert.NoError(t, f.Close())

	x, err = os.ReadFile(f.Name())
	assert.NoError(t, err)
	assert.Contains(t, string(x), "tester")

	tfo = &TemplateFileOp{
		File:   "non-existent-file",
		Output: "whatever",
	}
	assert.Error(t, tfo.Do(ctx))

}
