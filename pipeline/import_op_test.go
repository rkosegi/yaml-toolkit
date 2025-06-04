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
	"testing"

	"github.com/rkosegi/yaml-toolkit/dom"
	"github.com/stretchr/testify/assert"
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
	assert.Equal(t, "c", gd.Lookup("step1.data.root.list1[2]").AsLeaf().Value())

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
	assert.Equal(t, 456, gd.Lookup("step1.data.level1.level2a.level3b").AsLeaf().Value())

	gd = b.Container()
	is = ImportOp{
		File: "../testdata/doc1.yaml",
		Mode: ParseFileModeText,
		Path: "step3",
	}
	assert.NoError(t, New(WithData(gd)).Execute(&is))
	assert.NotEmpty(t, gd.Lookup("step3").AsLeaf().Value())
	assert.Contains(t, is.String(), "path=step3,mode=text")

	gd = b.Container()
	is = ImportOp{
		File: "../testdata/doc1.yaml",
		Mode: ParseFileModeBinary,
		Path: "files.doc1",
	}
	assert.NoError(t, New(WithData(gd)).Execute(&is))
	assert.NotEmpty(t, gd.Lookup("files.doc1").AsLeaf().Value())

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
