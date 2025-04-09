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

	"github.com/rkosegi/yaml-toolkit/patch"
)

// PatchOp performs RFC6902-style patch on global data document.
// Check patch package for more details
type PatchOp struct {
	Op   patch.Op `yaml:"op"`
	From string   `yaml:"from,omitempty" clone:"template"`
	Path string   `yaml:"path" clone:"template"`
	// Value is value to be used for op. This takes precedence over ValueFrom.
	Value *AnyVal `yaml:"value,omitempty"`
	// ValueFrom allow value to be read from data tree at given path.
	// Only considered when Value is nil.
	ValueFrom *string `yaml:"valueFrom,omitempty" clone:"template"`
}

func (ps *PatchOp) String() string {
	return fmt.Sprintf("Patch[Op=%s,Path=%s]", ps.Op, ps.Path)
}

func (ps *PatchOp) Do(ctx ActionContext) error {
	ss := ctx.Snapshot()
	oo := &patch.OpObj{
		Op: ps.Op,
	}
	path, err := patch.ParsePath(ctx.TemplateEngine().RenderLenient(ps.Path, ss))
	if err != nil {
		return err
	}
	oo.Path = path
	if ps.Value != nil {
		oo.Value = ps.Value.Value()
	} else if ps.ValueFrom != nil {
		oo.Value = ctx.Data().Lookup(ctx.TemplateEngine().RenderLenient(*ps.ValueFrom, ss))
	}
	if len(ps.From) > 0 {
		from, err := patch.ParsePath(ps.From)
		if err != nil {
			return err
		}
		oo.From = &from
	}
	ctx.Logger().Log(fmt.Sprintf("Patch[Op=%v,Path=%v]", oo.Op, oo.Path))
	return patch.Do(oo, ctx.Data())
}

func (ps *PatchOp) CloneWith(ctx ActionContext) Action {
	ss := ctx.Snapshot()
	return &PatchOp{
		Op:        ps.Op,
		Value:     ps.Value,
		ValueFrom: safeRenderStrPointer(ps.ValueFrom, ctx.TemplateEngine(), ss),
		From:      ctx.TemplateEngine().RenderLenient(ps.From, ss),
		Path:      ctx.TemplateEngine().RenderLenient(ps.Path, ss),
	}
}
