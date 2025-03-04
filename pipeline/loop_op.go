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

import "reflect"

// LoopOp is similar to loop statement.
type LoopOp struct {
	// Init is called just before any loop execution takes place
	Init *ActionSpec `yaml:"init,omitempty"`

	// Test is condition that is tested before each iteration.
	// When evaluated to true, execution will proceed with next iteration,
	// false terminates loop immediately
	Test string `yaml:"test,omitempty"`

	// Action is action that is executed every loop iteration
	Action ActionSpec `yaml:"action,omitempty"`

	// PostAction is action that is executed after every loop iteration.
	// This is right place to modify loop variables, such as incrementing counter
	PostAction *ActionSpec `yaml:"postAction,omitempty"`
}

func (l *LoopOp) String() string {
	return "Loop[]"
}

func (l *LoopOp) doAction(ctx ActionContext, act Action) (err error) {
	if act != nil && !reflect.ValueOf(act).IsNil() {
		return ctx.Executor().Execute(act)
	}
	return nil
}

func (l *LoopOp) Do(ctx ActionContext) (err error) {
	if err = l.doAction(ctx, l.Init); err != nil {
		return err
	}

	for {
		var next bool
		if next, err = ctx.TemplateEngine().EvalBool(l.Test, ctx.Snapshot()); err != nil {
			return err
		}
		if next {
			if err = l.doAction(ctx, l.PostAction); err != nil {
				return err
			}
			if err = l.Action.Do(ctx); err != nil {
				return err
			}
		} else {
			return nil
		}
	}
}

func (l *LoopOp) CloneWith(ctx ActionContext) Action {
	lc := new(LoopOp)
	lc.Test = l.Test
	lc.Action = l.Action.CloneWith(ctx).(ActionSpec)
	if l.Init != nil {
		lc.Init = ptr(l.Init.CloneWith(ctx).(ActionSpec))
	}
	if l.PostAction != nil {
		lc.PostAction = ptr(l.PostAction.CloneWith(ctx).(ActionSpec))
	}
	return lc
}
