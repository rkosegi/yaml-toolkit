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
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/rkosegi/yaml-toolkit/dom"
	"github.com/rkosegi/yaml-toolkit/patch"
	"github.com/rkosegi/yaml-toolkit/props"
)

func (am ActionMeta) String() string {
	return fmt.Sprintf("[name=%s,order=%d,when=%s]", am.Name, am.Order, safeStrDeref(am.When))
}

func (s ActionSpec) CloneWith(ctx ActionContext) Action {
	return ActionSpec{
		ActionMeta: s.ActionMeta,
		Operations: s.Operations.CloneWith(ctx).(OpSpec),
		Children:   s.Children.CloneWith(ctx).(ChildActions),
	}
}

func (s ActionSpec) String() string {
	return fmt.Sprintf("ActionSpec[meta=%v]", s.ActionMeta)
}

func (s ActionSpec) Do(ctx ActionContext) error {
	for _, a := range []Action{s.Operations, s.Children} {
		if s.When != nil {
			if ok, err := ctx.TemplateEngine().EvalBool(*s.When, ctx.Snapshot()); err != nil {
				return err
			} else if !ok {
				return nil
			}
		}
		err := ctx.Executor().Execute(a)
		if err != nil {
			return err
		}
	}
	return nil
}

func (pfm ParseFileMode) toValue(content []byte) (dom.Node, error) {
	switch pfm {
	case ParseFileModeBinary:
		return dom.LeafNode(base64.StdEncoding.EncodeToString(content)), nil
	case ParseFileModeText:
		return dom.LeafNode(string(content)), nil
	case ParseFileModeYaml:
		return b.FromReader(bytes.NewReader(content), dom.DefaultYamlDecoder)
	case ParseFileModeJson:
		return b.FromReader(bytes.NewReader(content), dom.DefaultJsonDecoder)
	case ParseFileModeProperties:
		return b.FromReader(bytes.NewReader(content), props.DecoderFn)
	default:
		return nil, fmt.Errorf("invalid ParseFileMode: %v", pfm)
	}
}

func (ia *ImportOp) String() string {
	return fmt.Sprintf("Import[file=%s,path=%s,mode=%s]", ia.File, ia.Path, ia.Mode)
}

func (ia *ImportOp) Do(ctx ActionContext) error {
	val, err := parseFile(ia.File, ia.Mode)
	if err != nil {
		return err
	}
	if len(ia.Path) > 0 {
		ctx.Data().AddValueAt(ia.Path, val)
	} else {
		if !val.IsContainer() {
			return ErrNotContainer
		} else {
			for k, v := range val.(dom.Container).Children() {
				ctx.Data().AddValueAt(k, v)
			}
		}
	}
	return nil
}

func (ia *ImportOp) CloneWith(ctx ActionContext) Action {
	return &ImportOp{
		Mode: ia.Mode,
		Path: ctx.TemplateEngine().RenderLenient(ia.Path, ctx.Snapshot()),
		File: ctx.TemplateEngine().RenderLenient(ia.File, ctx.Snapshot()),
	}
}

func (ps *PatchOp) String() string {
	return fmt.Sprintf("Patch[Op=%s,Path=%s]", ps.Op, ps.Path)
}

func (ps *PatchOp) Do(ctx ActionContext) error {
	oo := &patch.OpObj{
		Op: ps.Op,
	}
	path, err := patch.ParsePath(ps.Path)
	if err != nil {
		return err
	}
	oo.Path = path
	oo.Value = b.FromMap(ps.Value)
	if len(ps.From) > 0 {
		from, err := patch.ParsePath(ps.From)
		if err != nil {
			return err
		}
		oo.From = &from
	}
	return patch.Do(oo, ctx.Data())
}

func (ps *PatchOp) CloneWith(ctx ActionContext) Action {
	return &PatchOp{
		Op:    ps.Op,
		Value: ps.Value,
		From:  ctx.TemplateEngine().RenderLenient(ps.From, ctx.Snapshot()),
		Path:  ctx.TemplateEngine().RenderLenient(ps.Path, ctx.Snapshot()),
	}
}

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

func (ts *TemplateOp) String() string {
	return fmt.Sprintf("Template[WriteTo=%s]", ts.WriteTo)
}

func (ts *TemplateOp) Do(ctx ActionContext) error {
	if len(ts.Template) == 0 {
		return ErrTemplateEmpty
	}
	if len(ts.WriteTo) == 0 {
		return ErrWriteToEmpty
	}
	val, err := ctx.TemplateEngine().Render(ts.Template, map[string]interface{}{
		"Data": ctx.Snapshot(),
	})
	ctx.Data().AddValueAt(ts.WriteTo, dom.LeafNode(val))
	return err
}

func (ts *TemplateOp) CloneWith(ctx ActionContext) Action {
	return &TemplateOp{
		Template: ts.Template,
		WriteTo:  ctx.TemplateEngine().RenderLenient(ts.WriteTo, ctx.Snapshot()),
	}
}
