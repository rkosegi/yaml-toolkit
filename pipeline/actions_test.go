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
	"github.com/rkosegi/yaml-toolkit/patch"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestExecuteImportOp(t *testing.T) {
	var (
		is ImportOp
		gd dom.ContainerBuilder
	)
	gd = b.Container()
	is = ImportOp{
		File: "../testdata/doc1.json",
		Path: "step1.data",
		Mode: ParseFileModeJson,
	}

	assert.NoError(t, New(WithData(gd)).Execute(&is))
	assert.Equal(t, "c", gd.Lookup("step1.data.root.list1[2]").(dom.Leaf).Value())

	// parsing YAML file as JSON should lead to error
	is = ImportOp{
		File: "../testdata/doc1.yaml",
		Mode: ParseFileModeJson,
	}
	assert.Error(t, New(WithData(gd)).Execute(&is))

	gd = b.Container()
	is = ImportOp{
		File: "../testdata/doc1.yaml",
		Mode: ParseFileModeYaml,
		Path: "step1.data",
	}
	assert.NoError(t, New(WithData(gd)).Execute(&is))
	assert.Equal(t, 456, gd.Lookup("step1.data.level1.level2a.level3b").(dom.Leaf).Value())

	gd = b.Container()
	is = ImportOp{
		File: "../testdata/doc1.yaml",
		Mode: ParseFileModeText,
		Path: "step3",
	}
	assert.NoError(t, New(WithData(gd)).Execute(&is))
	assert.NotEmpty(t, gd.Lookup("step3").(dom.Leaf).Value())
	assert.Contains(t, is.String(), "path=step3,mode=text")

	gd = b.Container()
	is = ImportOp{
		File: "../testdata/doc1.yaml",
		Mode: ParseFileModeBinary,
		Path: "files.doc1",
	}
	assert.NoError(t, New(WithData(gd)).Execute(&is))
	assert.NotEmpty(t, gd.Lookup("files.doc1").(dom.Leaf).Value())

	gd = b.Container()
	is = ImportOp{
		File: "../testdata/doc1.json",
		Path: "files.doc1_json",
	}
	assert.NoError(t, New(WithData(gd)).Execute(&is))
	assert.Contains(t, is.String(), "path=files.doc1_json,mode=")

	is = ImportOp{
		File: "non-existent-file.ext",
		Path: "something",
	}
	assert.Error(t, New(WithData(gd)).Execute(&is))

	is = ImportOp{
		File: "../testdata/props1.properties",
		Mode: ParseFileModeProperties,
		Path: "props",
	}
	assert.NoError(t, New(WithData(gd)).Execute(&is))

	// no path provided and data is not a container - error
	is = ImportOp{
		File: "../testdata/props1.properties",
		Mode: ParseFileModeText,
	}
	assert.Error(t, New(WithData(gd)).Execute(&is))

	// import directly to root (with no path)
	is = ImportOp{
		File: "../testdata/doc1.json",
		Mode: ParseFileModeJson,
	}
	assert.NoError(t, New(WithData(gd)).Execute(&is))

	is = ImportOp{
		File: "../testdata/props1.properties",
		Path: "something",
		Mode: "invalid-mode",
	}
	assert.Error(t, New(WithData(gd)).Execute(&is))
}

func TestExecutePatchOp(t *testing.T) {
	var (
		ps PatchOp
		gd dom.ContainerBuilder
	)

	ps = PatchOp{
		Op:   patch.OpAdd,
		Path: "@#$%^&",
	}
	assert.Error(t, New(WithData(gd)).Execute(&ps))

	gd = b.Container()
	gd.AddValueAt("root.sub1.leaf2", dom.LeafNode("abcd"))
	ps = PatchOp{
		Op:   patch.OpReplace,
		Path: "/root/sub1",
		Value: map[string]interface{}{
			"leaf2": "xyz",
		},
	}
	assert.NoError(t, New(WithData(gd)).Execute(&ps))
	assert.Equal(t, "xyz", gd.Lookup("root.sub1.leaf2").(dom.Leaf).Value())
	assert.Contains(t, ps.String(), "Op=replace,Path=/root/sub1")

	gd = b.Container()
	gd.AddValueAt("root.sub1.leaf3", dom.LeafNode("abcd"))
	ps = PatchOp{
		Op:   patch.OpMove,
		From: "/root/sub1",
		Path: "/root/sub2",
	}
	assert.NoError(t, New(WithData(gd)).Execute(&ps))
	assert.Equal(t, "abcd", gd.Lookup("root.sub2.leaf3").(dom.Leaf).Value())

	gd = b.Container()
	gd.AddValueAt("root.sub1.leaf3", dom.LeafNode("abcd"))
	ps = PatchOp{
		Op:   patch.OpMove,
		From: "%#$&^^*&",
		Path: "/root/sub2",
	}
	assert.Error(t, New(WithData(gd)).Execute(&ps))
}

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
