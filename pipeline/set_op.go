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
	"fmt"
)

func (sa *SetOp) String() string {
	return fmt.Sprintf("Set[Path=%s]", sa.Path)
}

func (sa *SetOp) Do(ctx ActionContext) error {
	gd := ctx.Data()
	if sa.Data == nil {
		return ErrNoDataToSet
	}
	data := ctx.Factory().FromMap(sa.Data)
	if len(sa.Path) > 0 {
		gd.AddValueAt(sa.Path, data)
	} else {
		for k, v := range data.Children() {
			gd.AddValueAt(k, v)
		}
	}
	return nil
}

func (sa *SetOp) CloneWith(ctx ActionContext) Action {
	return &SetOp{
		Data: sa.Data,
		Path: ctx.TemplateEngine().RenderLenient(sa.Path, ctx.Snapshot()),
	}
}
