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

import "fmt"

func (as OpSpec) toList() []Action {
	actions := make([]Action, 0)
	if as.Set != nil {
		actions = append(actions, as.Set)
	}
	if as.Import != nil {
		actions = append(actions, as.Import)
	}
	if as.Patch != nil {
		actions = append(actions, as.Patch)
	}
	if as.Template != nil {
		actions = append(actions, as.Template)
	}
	if as.ForEach != nil {
		actions = append(actions, as.ForEach)
	}
	return actions
}

func (as OpSpec) Do(ctx ActionContext) error {
	for _, a := range as.toList() {
		err := ctx.Executor().Execute(a)
		if err != nil {
			return err
		}
	}
	return nil
}

func (as OpSpec) CloneWith(ctx ActionContext) Action {
	r := OpSpec{}
	if as.ForEach != nil {
		r.ForEach = as.ForEach.CloneWith(ctx).(*ForEachOp)
	}
	if as.Import != nil {
		r.Import = as.Import.CloneWith(ctx).(*ImportOp)
	}
	if as.Patch != nil {
		r.Patch = as.Patch.CloneWith(ctx).(*PatchOp)
	}
	if as.Set != nil {
		r.Set = as.Set.CloneWith(ctx).(*SetOp)
	}
	if as.Template != nil {
		r.Template = as.Template.CloneWith(ctx).(*TemplateOp)
	}
	return r
}

func (as OpSpec) String() string {
	return fmt.Sprintf("OpSpec[ForEach=%v,Import=%v,Patch=%v,Set=%v,Template=%v]",
		as.ForEach, as.Import, as.Patch, as.Set, as.Template)
}
