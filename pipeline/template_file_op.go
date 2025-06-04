/*
Copyright 2025 Richard Kosegi

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
	"os"

	"github.com/rkosegi/yaml-toolkit/dom"
)

// TemplateFileOp can be used to render template from file and write result to output.
type TemplateFileOp struct {
	// File is path to file with template
	File string `yaml:"file" clone:"template"`
	// Output is path to output file
	Output string `yaml:"output" clone:"template"`
	// Path is path within the global data where data are read from (must be container).
	// When omitted, then root of global data is assumed.
	Path *string `yaml:"path,omitempty" clone:"template"`
}

func (tfo *TemplateFileOp) String() string {
	return fmt.Sprintf("TemplateFile[File=%s,Output=%s]", tfo.File, tfo.Output)
}

func (tfo *TemplateFileOp) Do(ctx ActionContext) error {
	if len(tfo.File) == 0 {
		return ErrFileEmpty
	}
	if len(tfo.Output) == 0 {
		return ErrOutputEmpty
	}
	var (
		data dom.Container
		ss   map[string]interface{}
	)
	data = ctx.Data()
	if tfo.Path != nil {
		if n := ctx.Data().Lookup(*tfo.Path); n != nil && n.IsContainer() {
			data = n.AsContainer()
		} else {
			return fmt.Errorf("path does not point to a container: %s", *tfo.Path)
		}
	}
	ss = dom.DefaultNodeEncoderFn(data).(map[string]interface{})
	inFile := ctx.TemplateEngine().RenderLenient(tfo.File, ss)
	ctx.Logger().Log("reading template file", inFile)
	tmpl, err := os.ReadFile(inFile)
	if err != nil {
		return err
	}
	val, err := ctx.TemplateEngine().Render(string(tmpl), ss)
	if err != nil {
		return err
	}
	outFile := ctx.TemplateEngine().RenderLenient(tfo.Output, ss)
	ctx.Logger().Log("writing rendered template", outFile)
	return os.WriteFile(outFile, []byte(val), 0644)
}

func (tfo *TemplateFileOp) CloneWith(ctx ActionContext) Action {
	ss := ctx.Snapshot()
	return &TemplateFileOp{
		File:   ctx.TemplateEngine().RenderLenient(tfo.File, ss),
		Output: ctx.TemplateEngine().RenderLenient(tfo.Output, ss),
	}
}
