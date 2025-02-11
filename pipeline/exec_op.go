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
	"io"
	"os"
	osx "os/exec"
	"slices"
	"strings"

	"github.com/rkosegi/yaml-toolkit/dom"
)

type ExecOp struct {
	// Program to execute
	Program string `yaml:"program,omitempty"`
	// Optional arguments for program
	Args *[]string `yaml:"args,omitempty"`
	// List of exit codes that are assumed to be valid
	ValidExitCodes *[]int `yaml:"validExitCodes,omitempty"`
	// Path to file where program's stdout will be written upon completion.
	// Any error occurred during write will result in panic.
	Stdout *string
	// Path to file where program's stderr will be written upon completion
	// Any error occurred during write will result in panic.
	Stderr *string
	// Path within the global data where to set exit code.
	SaveExitCodeTo *string `yaml:"saveExitCodeTo,omitempty"`
}

func (e *ExecOp) String() string {
	return fmt.Sprintf("Exec[Program=%s,Args=%d]", e.Program, safeStrListSize(e.Args))
}

func (e *ExecOp) Do(ctx ActionContext) error {
	var closables []io.Closer
	if e.ValidExitCodes == nil {
		e.ValidExitCodes = &[]int{}
	}
	if e.Args == nil {
		e.Args = &[]string{}
	}
	snapshot := ctx.Snapshot()
	prog := ctx.TemplateEngine().RenderLenient(e.Program, snapshot)
	args := *safeRenderStrSlice(e.Args, ctx.TemplateEngine(), snapshot)
	cmd := osx.Command(prog, args...)
	defer func() {
		for _, closer := range closables {
			_ = closer.Close()
		}
	}()
	type streamTgt struct {
		output *string
		target *io.Writer
	}
	for _, stream := range []streamTgt{
		{
			output: e.Stdout,
			target: &cmd.Stdout,
		},
		{
			output: e.Stderr,
			target: &cmd.Stderr,
		},
	} {
		if stream.output != nil {
			s := ctx.TemplateEngine().RenderLenient(*stream.output, snapshot)
			stream.output = &s
			out, err := os.OpenFile(*stream.output, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
			if err != nil {
				return err
			}
			*stream.target = out
			closables = append(closables, out)
		}
	}
	ctx.Logger().Log(fmt.Sprintf("prog=%s,args=[%s]", prog, strings.Join(args, ",")))
	err := cmd.Run()
	var exitErr *osx.ExitError
	if errors.As(err, &exitErr) {
		if e.SaveExitCodeTo != nil {
			ctx.Data().AddValueAt(*e.SaveExitCodeTo, dom.LeafNode(exitErr.ExitCode()))
		}
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
		Program:        ctx.TemplateEngine().RenderLenient(e.Program, ss),
		Args:           safeRenderStrSlice(e.Args, ctx.TemplateEngine(), ss),
		Stdout:         safeRenderStrPointer(e.Stdout, ctx.TemplateEngine(), ss),
		Stderr:         safeRenderStrPointer(e.Stderr, ctx.TemplateEngine(), ss),
		ValidExitCodes: safeCopyIntSlice(e.ValidExitCodes),
	}
}
