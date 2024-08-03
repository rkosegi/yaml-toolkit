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
	"github.com/rkosegi/yaml-toolkit/dom"
	"github.com/rkosegi/yaml-toolkit/props"
	"os"
)

type OutputFormat string

const (
	OutputFormatYaml       = OutputFormat("yaml")
	OutputFormatJson       = OutputFormat("json")
	OutputFormatProperties = OutputFormat("properties")
)

// ExportOp allows to export data into file
type ExportOp struct {
	// File to export data onto
	File string
	// Path within data tree pointing to dom.Container to export. Empty path denotes whole document.
	// If path does not resolve or resolves to dom.Node that is not dom.Container,
	// then empty document will be exported.
	Path string
	// Format of output file.
	Format OutputFormat
}

func (e *ExportOp) String() string {
	return fmt.Sprintf("Export[file=%s,format=%s,path=%s]", e.File, e.Format, e.Path)
}

func (e *ExportOp) Do(ctx ActionContext) (err error) {
	var (
		d   dom.Node
		enc dom.EncoderFunc
	)
	d = ctx.Data()
	if len(e.Path) > 0 {
		d = ctx.Data().Lookup(e.Path)
	}
	// use empty container, since path lookup didn't yield anything useful
	if d == nil || !d.IsContainer() {
		d = b.Container()
	}
	switch e.Format {
	case OutputFormatYaml:
		enc = dom.DefaultYamlEncoder
	case OutputFormatJson:
		enc = dom.DefaultJsonEncoder
	case OutputFormatProperties:
		enc = props.EncoderFn

	default:
		return fmt.Errorf("unknown output format: %s", e.Format)
	}

	f, err := os.OpenFile(e.File, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
	}()
	return enc(f, dom.DefaultNodeMappingFn(d.(dom.Container)))
}

func (e *ExportOp) CloneWith(ctx ActionContext) Action {
	ss := ctx.Snapshot()
	return &ExportOp{
		File:   ctx.TemplateEngine().RenderLenient(e.File, ss),
		Path:   ctx.TemplateEngine().RenderLenient(e.Path, ss),
		Format: OutputFormat(ctx.TemplateEngine().RenderLenient(string(e.Format), ss)),
	}
}
