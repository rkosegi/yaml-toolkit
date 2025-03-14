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
	"strings"

	"github.com/rkosegi/yaml-toolkit/dom"
)

// TemplateOp can be used to render value from data at runtime.
// Global data tree is available under .Data
type TemplateOp struct {
	// template to render
	Template string `yaml:"template"`
	// path within global data tree where to set result at
	Path string `yaml:"path" clone:"template"`
	// Trim when true, whitespace is trimmed off the value
	Trim *bool `yaml:"trim,omitempty"`
}

func (ts *TemplateOp) String() string {
	return fmt.Sprintf("Template[Path=%s]", ts.Path)
}

func (ts *TemplateOp) Do(ctx ActionContext) error {
	if len(ts.Template) == 0 {
		return ErrTemplateEmpty
	}
	if len(ts.Path) == 0 {
		return ErrPathEmpty
	}
	ss := ctx.Snapshot()
	val, err := ctx.TemplateEngine().Render(ts.Template, ss)
	if safeBoolDeref(ts.Trim) {
		val = strings.TrimSpace(val)
	}
	ctx.Data().AddValueAt(ctx.TemplateEngine().RenderLenient(ts.Path, ss), dom.LeafNode(val))
	return err
}

func (ts *TemplateOp) CloneWith(ctx ActionContext) Action {
	return &TemplateOp{
		Template: ts.Template,
		Trim:     ts.Trim,
		Path:     ctx.TemplateEngine().RenderLenient(ts.Path, ctx.Snapshot()),
	}
}
