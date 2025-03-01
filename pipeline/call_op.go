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
	Name string
	Args map[string]interface{}
}

func (c *CallOp) String() string {
	return fmt.Sprintf("Call[Name=%s, Args=%d]", c.Name, len(c.Args))
}

func (c *CallOp) Do(ctx ActionContext) error {
	if spec, exists := ctx.Ext().Get(c.Name); !exists {
		return fmt.Errorf("callable '%s' is not registered", c.Name)
	} else {
		ctx.Data().AddValue("args", dom.DefaultNodeDecoderFn(c.Args))
		defer ctx.Data().Remove("args")
		return ctx.Executor().Execute(spec)
	}
}

func (c *CallOp) CloneWith(_ ActionContext) Action {
	return &CallOp{
		Name: c.Name,
		Args: c.Args,
	}
}
