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
	"io"
	"os"

	"github.com/rkosegi/yaml-toolkit/dom"
	"github.com/rkosegi/yaml-toolkit/props"
)

type OutputFormat string

const (
	OutputFormatYaml       = OutputFormat("yaml")
	OutputFormatJson       = OutputFormat("json")
	OutputFormatProperties = OutputFormat("properties")
	OutputFormatText       = OutputFormat("text")
)

// ExportOp allows to export data into file
type ExportOp struct {
	// File to export data onto
	File string `clone:"template"`
	// Path within data tree pointing to dom.Node to export. Empty path denotes whole document.
	// If path does not resolve, then empty document will be exported.
	// If output format is "text" then path must point to leaf.
	// Any other output format must point to dom.Container.
	// If neither of these conditions are met, then it is considered as if path does not resolve at all.
	Path string `clone:"template"`
	// Format of output file.
	Format OutputFormat `clone:"template"`
}

func (e *ExportOp) String() string {
	return fmt.Sprintf("Export[file=%s,format=%s,path=%s]", e.File, e.Format, e.Path)
}

func (e *ExportOp) Do(ctx ActionContext) (err error) {
	var (
		d      dom.Node
		enc    dom.EncoderFunc
		defVal dom.Node
	)
	defVal = b.Container()
	d = ctx.Data()
	if len(e.Path) > 0 {
		d = ctx.Data().Lookup(e.Path)
	}
	switch e.Format {
	case OutputFormatYaml:
		enc = dom.DefaultYamlEncoder
	case OutputFormatJson:
		enc = dom.DefaultJsonEncoder
	case OutputFormatProperties:
		enc = props.EncoderFn
	case OutputFormatText:
		enc = func(w io.Writer, v interface{}) error {
			_, err = fmt.Fprintf(w, "%v", v)
			return err
		}
		defVal = dom.LeafNode("")

	default:
		return fmt.Errorf("unknown output format: %s", e.Format)
	}
	if d == nil {
		d = defVal
	}
	fp := ctx.TemplateEngine().RenderLenient(e.File, ctx.Snapshot())
	ctx.Logger().Log("opening file", fp)
	f, err := os.OpenFile(fp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
	}()
	if e.Format == OutputFormatText {
		if !d.IsLeaf() {
			return fmt.Errorf("unsupported node for 'text' format: %v", d)
		}
		return enc(f, d.(dom.Leaf).Value())
	}
	return enc(f, dom.DefaultNodeEncoderFn(d.(dom.Container)))
}

func (e *ExportOp) CloneWith(ctx ActionContext) Action {
	ss := ctx.Snapshot()
	return &ExportOp{
		File:   ctx.TemplateEngine().RenderLenient(e.File, ss),
		Path:   ctx.TemplateEngine().RenderLenient(e.Path, ss),
		Format: OutputFormat(ctx.TemplateEngine().RenderLenient(string(e.Format), ss)),
	}
}
