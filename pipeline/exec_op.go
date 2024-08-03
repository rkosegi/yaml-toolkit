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
	"errors"
	"fmt"
	osx "os/exec"
	"slices"
)

type ExecOp struct {
	// Program to execute
	Program string `yaml:"program,omitempty"`
	// Optional arguments for program
	Args *[]string `yaml:"args,omitempty"`
	// List of exit codes that are assumed to be valid
	ValidExitCodes *[]int `yaml:"validExitCodes,omitempty"`
}

func (e *ExecOp) String() string {
	return fmt.Sprintf("Exec[Program=%s,Args=%d]", e.Program, safeStrListSize(e.Args))
}

func (e *ExecOp) Do(_ ActionContext) error {
	if e.ValidExitCodes == nil {
		e.ValidExitCodes = &[]int{}
	}
	if e.Args == nil {
		e.Args = &[]string{}
	}
	cmd := osx.Command(e.Program, *e.Args...)
	err := cmd.Run()
	var exitErr *osx.ExitError
	if errors.As(err, &exitErr) {
		if !slices.Contains(*e.ValidExitCodes, exitErr.ExitCode()) {
			return err
		}
	} else {
		return err
	}
	return nil
}

func (e *ExecOp) CloneWith(ctx ActionContext) Action {
	ss := ctx.Snapshot()
	return &ExecOp{
		Program: ctx.TemplateEngine().RenderLenient(e.Program, ss),
		Args:    safeRenderStrSlice(e.Args, ctx.TemplateEngine(), ss),
	}
}
