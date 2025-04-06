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
)

type AbortOp struct {
	Message string `yaml:"message" clone:"template"`
}

func (ao *AbortOp) Do(ctx ActionContext) error {
	msg := ctx.TemplateEngine().RenderLenient(ao.Message, ctx.Snapshot())
	return errors.New(fmt.Sprintf("abort: %s", msg))
}

func (ao *AbortOp) String() string {
	return fmt.Sprintf("Abort[message=%s]", ao.Message)
}

func (ao *AbortOp) CloneWith(ctx ActionContext) Action {
	return &AbortOp{
		Message: ctx.TemplateEngine().RenderLenient(ao.Message, ctx.Snapshot()),
	}
}
