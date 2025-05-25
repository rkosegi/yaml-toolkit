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
	"strconv"
	"strings"

	"github.com/rkosegi/yaml-toolkit/dom"
	"gopkg.in/yaml.v3"
)

type ParseTextAs string

const (
	// ParseTextAsNone do not parse, it's just a text (dom.Leaf)
	ParseTextAsNone ParseTextAs = "none"
	// ParseTextAsYaml parse text as a YAML source into dom.Node
	ParseTextAsYaml ParseTextAs = "yaml"
	// ParseTextAsFloat64 parse text as float64 number into dom.Leaf
	ParseTextAsFloat64 ParseTextAs = "float64"
	// ParseTextAsInt64 parse text as int64 number into dom.Leaf
	ParseTextAsInt64 ParseTextAs = "int64"
)

// TemplateOp can be used to render value from data at runtime.
type TemplateOp struct {
	// template to render
	Template string `yaml:"template"`
	// path within global data tree where to set result at
	Path string `yaml:"path" clone:"template"`
	// How to treat rendered text after template engine completes successfully.
	// It's responsibility of template to produce source that is parseable by chosen mode
	ParseAs *ParseTextAs `yaml:"parseAs,omitempty" clone:"text"`
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
	if err != nil {
		return err
	}
	if safeBoolDeref(ts.Trim) {
		val = strings.TrimSpace(val)
	}
	if ts.ParseAs == nil {
		ts.ParseAs = ptr(ParseTextAsNone)
	}
	var node dom.Node
	switch *ts.ParseAs {
	case ParseTextAsYaml:
		var yn yaml.Node
		if err = yaml.Unmarshal([]byte(val), &yn); err != nil {
			return err
		}
		node = dom.YamlNodeDecoder()(&yn)
	case ParseTextAsNone:
		node = dom.LeafNode(val)
	case ParseTextAsFloat64:
		var x float64
		if x, err = strconv.ParseFloat(val, 64); err != nil {
			return err
		} else {
			node = dom.LeafNode(x)
		}
	case ParseTextAsInt64:
		var x int64
		if x, err = strconv.ParseInt(val, 10, 64); err != nil {
			return err
		} else {
			node = dom.LeafNode(x)
		}
	default:
		return fmt.Errorf("unknown ParseAs mode: %v", *ts.ParseAs)
	}
	ctx.Data().AddValueAt(ctx.TemplateEngine().RenderLenient(ts.Path, ss), node)
	ctx.InvalidateSnapshot()
	return err
}

func (ts *TemplateOp) CloneWith(ctx ActionContext) Action {
	return &TemplateOp{
		Template: ts.Template,
		Trim:     ts.Trim,
		Path:     ctx.TemplateEngine().RenderLenient(ts.Path, ctx.Snapshot()),
	}
}
