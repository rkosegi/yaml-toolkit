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
)

// OpSpec is specification of operation.
type OpSpec struct {
	// Set sets data in data document.
	Set *SetOp `yaml:"set,omitempty"`

	// Patch performs RFC6902-style patch on data document.
	Patch *PatchOp `yaml:"patch,omitempty"`

	// Import loads content of file into data document.
	Import *ImportOp `yaml:"import,omitempty"`

	// Template allows to render value at runtime
	Template *TemplateOp `yaml:"template,omitempty"`

	// Call calls previously defined callable
	Call *CallOp `yaml:"call,omitempty"`

	// Define defines callable ActionSpec
	Define *DefineOp `yaml:"define,omitempty"`

	// Env adds OS environment variables into data document
	Env *EnvOp `yaml:"env,omitempty"`

	// Exec executes program
	Exec *ExecOp `yaml:"exec,omitempty"`

	// Export exports data document into file
	Export *ExportOp `yaml:"export,omitempty"`

	// Ext allows runtime-registered extension action to be executed
	Ext *ExtOp `yaml:"ext,omitempty"`

	// ForEach execute same operation in a loop for every configured item
	ForEach *ForEachOp `yaml:"forEach,omitempty"`

	// Log logs arbitrary message to logger
	Log *LogOp `yaml:"log,omitempty"`

	// Loop allows for execution to be done in a loop
	Loop *LoopOp `yaml:"loop,omitempty"`

	// Abort is able to signal error, so that pipeline can abort execution
	Abort *AbortOp `yaml:"abort,omitempty"`
}

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
	if as.Call != nil {
		actions = append(actions, as.Call)
	}
	if as.Define != nil {
		actions = append(actions, as.Define)
	}
	if as.Export != nil {
		actions = append(actions, as.Export)
	}
	if as.Ext != nil {
		actions = append(actions, as.Ext)
	}
	if as.Env != nil {
		actions = append(actions, as.Env)
	}
	if as.Exec != nil {
		actions = append(actions, as.Exec)
	}
	if as.ForEach != nil {
		actions = append(actions, as.ForEach)
	}
	if as.Log != nil {
		actions = append(actions, as.Log)
	}
	if as.Loop != nil {
		actions = append(actions, as.Loop)
	}
	if as.Abort != nil {
		actions = append(actions, as.Abort)
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
	if as.Call != nil {
		r.Call = as.Call.CloneWith(ctx).(*CallOp)
	}
	if as.Define != nil {
		r.Define = as.Define.CloneWith(ctx).(*DefineOp)
	}
	if as.Export != nil {
		r.Export = as.Export.CloneWith(ctx).(*ExportOp)
	}
	if as.Ext != nil {
		r.Ext = as.Ext.CloneWith(ctx).(*ExtOp)
	}
	if as.Env != nil {
		r.Env = as.Env.CloneWith(ctx).(*EnvOp)
	}
	if as.Exec != nil {
		r.Exec = as.Exec.CloneWith(ctx).(*ExecOp)
	}
	if as.Log != nil {
		r.Log = as.Log.CloneWith(ctx).(*LogOp)
	}
	if as.Loop != nil {
		r.Loop = as.Loop.CloneWith(ctx).(*LoopOp)
	}
	if as.Abort != nil {
		r.Abort = as.Abort.CloneWith(ctx).(*AbortOp)
	}
	return r
}

func (as OpSpec) String() string {
	var sb strings.Builder
	parts := make([]string, 0)
	sb.WriteString("OpSpec[")
	if as.Abort != nil {
		parts = append(parts, fmt.Sprintf("Abort=%v", as.Abort.String()))
	}
	if as.Call != nil {
		parts = append(parts, fmt.Sprintf("Call=%v", as.Call.String()))
	}
	if as.Define != nil {
		parts = append(parts, fmt.Sprintf("Define=%v", as.Define.String()))
	}
	if as.Env != nil {
		parts = append(parts, fmt.Sprintf("Env=%v", as.Env.String()))
	}
	if as.Export != nil {
		parts = append(parts, fmt.Sprintf("Export=%v", as.Export.String()))
	}
	if as.Ext != nil {
		parts = append(parts, fmt.Sprintf("Ext=%v", as.Ext.String()))
	}
	if as.Exec != nil {
		parts = append(parts, fmt.Sprintf("Exec=%v", as.Exec.String()))
	}
	if as.ForEach != nil {
		parts = append(parts, fmt.Sprintf("ForEach=%v", as.ForEach.String()))
	}
	if as.Import != nil {
		parts = append(parts, fmt.Sprintf("Import=%v", as.Import.String()))
	}
	if as.Log != nil {
		parts = append(parts, fmt.Sprintf("Log=%v", as.Log.String()))
	}
	if as.Loop != nil {
		parts = append(parts, fmt.Sprintf("Loop=%v", as.Loop.String()))
	}
	if as.Patch != nil {
		parts = append(parts, fmt.Sprintf("Patch=%v", as.Patch.String()))
	}
	if as.Set != nil {
		parts = append(parts, fmt.Sprintf("Set=%v", as.Set.String()))
	}
	if as.Template != nil {
		parts = append(parts, fmt.Sprintf("Template=%v", as.Template.String()))
	}
	sb.WriteString(strings.Join(parts, ","))
	sb.WriteString("]")
	return sb.String()
}
