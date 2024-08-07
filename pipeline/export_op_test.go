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
	"testing"
)

func TestExportOpDo(t *testing.T) {
	var (
		eo  *ExportOp
		err error
	)
	f, err := os.CreateTemp("", "yt_export*.json")
	assert.NoError(t, err)
	if err != nil {
		return
	}
	t.Cleanup(func() {
		t.Logf("cleanup temporary file %s", f.Name())
		_ = os.Remove(f.Name())
	})
	t.Logf("created temporary file: %s", f.Name())
	eo = &ExportOp{
		File:   f.Name(),
		Path:   "root.sub1",
		Format: OutputFormatJson,
	}
	assert.Contains(t, eo.String(), f.Name())
	d := b.Container()
	d.AddValueAt("root.sub1.sub2", dom.LeafNode(123))
	err = eo.Do(mockActCtx(d))
	assert.NoError(t, err)
	fi, err := os.Stat(f.Name())
	assert.NotNil(t, fi)
	assert.NoError(t, err)

	eo = &ExportOp{
		File:   f.Name(),
		Path:   "root.sub1.sub2",
		Format: OutputFormatText,
	}
	err = eo.Do(mockActCtx(d))
	assert.NoError(t, err)

	eo = &ExportOp{
		File:   f.Name(),
		Path:   "root.sub1",
		Format: OutputFormatText,
	}
	err = eo.Do(mockActCtx(d))
	assert.Error(t, err)
}

func TestExportOpDoInvalidDirectory(t *testing.T) {
	eo := &ExportOp{
		File:   "/invalid/dir/file.yaml",
		Format: OutputFormatYaml,
	}
	assert.Error(t, eo.Do(mockEmptyActCtx()))
}

func TestExportOpDoInvalidOutFormat(t *testing.T) {
	eo := &ExportOp{
		Format: "invalid-format",
	}
	assert.Error(t, eo.Do(mockEmptyActCtx()))
}

func TestExportOpDoNonExistentPath(t *testing.T) {
	f, err := os.CreateTemp("", "yt_export*.json")
	assert.NoError(t, err)
	if err != nil {
		return
	}
	t.Cleanup(func() {
		t.Logf("cleanup temporary file %s", f.Name())
		_ = os.Remove(f.Name())
	})
	eo := &ExportOp{
		File:   f.Name(),
		Path:   "this.path.does.not.exist",
		Format: OutputFormatProperties,
	}
	assert.NoError(t, eo.Do(mockEmptyActCtx()))
}

func TestExportOpCloneWith(t *testing.T) {
	eo := &ExportOp{
		File:   "/tmp/out.{{ .Format }}",
		Path:   "root.sub10.{{ .Sub }}",
		Format: "{{ .Format }}",
	}
	d := b.Container()
	d.AddValueAt("Format", dom.LeafNode("yaml"))
	d.AddValueAt("Sub", dom.LeafNode("sub20"))
	eo = eo.CloneWith(mockActCtx(d)).(*ExportOp)
	assert.Equal(t, "root.sub10.sub20", eo.Path)
	assert.Equal(t, OutputFormatYaml, eo.Format)
	assert.Equal(t, "/tmp/out.yaml", eo.File)
}
