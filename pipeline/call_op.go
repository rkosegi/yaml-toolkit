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
	"fmt"

	"github.com/rkosegi/yaml-toolkit/dom"
)

type CallOp struct {
	// Name is name of callable previously registered using DefineOp.
	// Attempt to use name that was not registered will result in error
	Name string
	// ArgsPath is optional path within the global data where arguments are stored prior to execution.
	// When omitted, then default value of "args" is assumed. Note that passing arguments to nested callable
	// is only possible if path is different, otherwise inner's arguments will overwrite outer's one.
	// Template is accepted as possible value.
	ArgsPath *string `yaml:"argsPath,omitempty"`
	// Arguments to be passed to callable.
	// Leaf values are recursively templated just before call is executed.
	Args map[string]interface{}
}

func (c *CallOp) String() string {
	return fmt.Sprintf("Call[Name=%s, Args=%d]", c.Name, len(c.Args))
}

func (c *CallOp) Do(ctx ActionContext) error {
	snap := ctx.Snapshot()
	ap := "args"
	if c.ArgsPath != nil {
		ap = *c.ArgsPath
	}
	ap = ctx.TemplateEngine().RenderLenient(ap, snap)
	if spec, exists := ctx.Ext().GetAction(c.Name); !exists {
		return fmt.Errorf("callable '%s' is not registered", c.Name)
	} else {
		ctx.Data().AddValueAt(ap, dom.DefaultNodeDecoderFn(
			ctx.TemplateEngine().RenderMapLenient(c.Args, snap)),
		)
		defer func() {
			ctx.Data().Remove(ap)
			ctx.InvalidateSnapshot()
		}()
		return ctx.Executor().Execute(spec)
	}
}

func (c *CallOp) CloneWith(_ ActionContext) Action {
	return &CallOp{
		Name:     c.Name,
		Args:     c.Args,
		ArgsPath: c.ArgsPath,
	}
}
