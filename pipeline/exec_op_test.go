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

func TestExecOpDoEmptyCommand(t *testing.T) {
	eo := &ExecOp{}
	assert.Error(t, eo.Do(mockEmptyActCtx()))
}

func TestExecOpDo(t *testing.T) {
	var (
		eo *ExecOp
	)
	fout, err := os.CreateTemp("", "yt.*.txt")
	ferr, err := os.CreateTemp("", "yt.*.txt")
	removeFilesLater(t, fout, ferr)
	assert.NoError(t, err)
	eo = &ExecOp{
		Program: "sh",
		Args:    &[]string{"-c", "echo abcd"},
		Stdout:  strPointer(fout.Name()),
		Stderr:  strPointer(ferr.Name()),
	}
	assert.NoError(t, eo.Do(mockEmptyActCtx()))

	eo = &ExecOp{
		Program: "sh",
		Args:    &[]string{"-c", "echo abcd"},
		Stdout:  strPointer("/"),
	}
	assert.Error(t, eo.Do(mockEmptyActCtx()))

	eo = &ExecOp{
		Program:        "sh",
		Args:           &[]string{"-c", "exit 3"},
		ValidExitCodes: &[]int{3},
	}
	assert.NoError(t, eo.Do(mockEmptyActCtx()))
	eo = &ExecOp{
		Program:        "sh",
		Args:           &[]string{"-c", "exit 4"},
		ValidExitCodes: &[]int{3},
	}
	assert.Contains(t, eo.String(), "sh")
	assert.Contains(t, eo.String(), "=2")
	assert.Error(t, eo.Do(mockEmptyActCtx()))
}

func TestExecOpCloneWith(t *testing.T) {
	eo := &ExecOp{
		Program: "{{ .Shell }}",
	}
	d := b.Container()
	d.AddValue("Shell", dom.LeafNode("/bin/bash"))
	eo = eo.CloneWith(mockActCtx(d)).(*ExecOp)
	assert.Equal(t, "/bin/bash", eo.Program)
}
