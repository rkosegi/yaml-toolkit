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
	"errors"
	"fmt"
	"github.com/rkosegi/yaml-toolkit/dom"
	"github.com/rkosegi/yaml-toolkit/patch"
	"regexp"
)

var (
	ErrNoDataToSet   = errors.New("no data to set")
	ErrTemplateEmpty = errors.New("template cannot be empty")
	ErrWriteToEmpty  = errors.New("writeTo cannot be empty")
	ErrNotContainer  = errors.New("data element must be container when no path is provided")
)

// ParseFileMode defines how the file is parsed before is put into data tree
type ParseFileMode string

const (
	// ParseFileModeBinary File is read and encoded using base64 string into data tree
	ParseFileModeBinary ParseFileMode = "binary"

	// ParseFileModeText File is read as-is and is assumed it represents utf-8 encoded byte stream
	ParseFileModeText ParseFileMode = "text"

	// ParseFileModeYaml File is parsed as YAML document and put as child node into data tree
	ParseFileModeYaml ParseFileMode = "yaml"

	// ParseFileModeJson File is parsed as JSON document and put as child node into data tree
	ParseFileModeJson ParseFileMode = "json"

	// ParseFileModeProperties File is parsed as Java properties into map[string]interface{} and put as child node into data tree
	ParseFileModeProperties ParseFileMode = "properties"
)

// Executor interface is used by external callers to execute Action items
type Executor interface {
	Execute(act Action) error
}

type TemplateEngine interface {
	Render(template string, data map[string]interface{}) (string, error)
	// RenderLenient attempts to render given template using provided data, while swallowing any error.
	// Value of template is first checked by simple means if it is actually template to avoid unnecessary errors.
	// Use with caution.
	RenderLenient(template string, data map[string]interface{}) string
	EvalBool(template string, data map[string]interface{}) (bool, error)
}

// Listener interface allows hook into execution of Action.
type Listener interface {
	// OnBefore is called just before act is executed
	OnBefore(ctx ActionContext)
	// OnAfter is called sometime after act is executed, regardless of result.
	// Any error returned by invoking Do() method is returned as last parameter.
	OnAfter(ctx ActionContext, err error)
}

// ActionContext is created by Executor implementation for sole purpose of invoking Action's Do function.
type ActionContext interface {
	// Data exposes data document
	Data() dom.ContainerBuilder
	// Snapshot is read-only view of Data() in point in time
	Snapshot() map[string]interface{}
	// Factory give access to factory to create new documents
	Factory() dom.ContainerFactory
	// Executor returns reference to executor
	Executor() Executor
	// TemplateEngine return reference to TemplateEngine
	TemplateEngine() TemplateEngine
}

// Action is implemented by actions within ActionSpec
type Action interface {
	fmt.Stringer
	// Do will perform this action.
	// This function is invoked by Executor implementation and as such it's not meant to be called by end user directly.
	Do(ctx ActionContext) error
	// CloneWith creates fresh clone of this Action with values of its fields templated.
	// Data for template can be obtained by calling Snapshot() on provided context.
	CloneWith(ctx ActionContext) Action
}

// ChildActions is map of named actions that are executed as a part of parent action
type ChildActions map[string]ActionSpec

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

	// Env adds OS environment variables into data document
	Env *EnvOp `yaml:"env,omitempty"`

	// Export exports data document into file
	Export *ExportOp `yaml:"export,omitempty"`
	// ForEach execute same operation in a loop for every configured item
	ForEach *ForEachOp `yaml:"forEach,omitempty"`
}

type ActionSpec struct {
	ActionMeta `yaml:",inline"`
	// Operations to perform
	Operations OpSpec `yaml:",inline"`
	// Children element is an optional map of child actions that will be executed
	// as a part of this action (after any of OpSpec in Operations are performed).
	// Exact order of execution is given by Order field value (lower the value, sooner the execution will take place).
	Children ChildActions `yaml:"steps,omitempty"`
}

// ActionMeta holds action's metadata used by Executor
type ActionMeta struct {
	// Name of this step, should be unique within current scope
	Name string `yaml:"name,omitempty"`

	// Optional ordinal number that controls order of execution within parent step
	Order int `yaml:"order,omitempty"`

	// Optional expression to make execution of this action conditional.
	// Execution of this step is skipped when this expression is evaluated to false.
	// If value of this field is omitted, then this action is executed.
	When *string `yaml:"when,omitempty"`
}

// EnvOp is used to import OS environment variables into data
type EnvOp struct {
	// Optional regexp which defines what to include. Only item names matching this regexp are added into data document.
	Include *regexp.Regexp `yaml:"include,omitempty"`

	// Optional regexp which defines what to exclude. Only item names NOT matching this regexp are added into data document.
	// Exclusion is considered after inclusion regexp is processed.
	Exclude *regexp.Regexp `yaml:"exclude,omitempty"`

	// Optional path within data tree under which "Env" container will be put.
	// When omitted, then "Env" goes to root of data.
	Path string `yaml:"path,omitempty"`

	// for mock purposes only. this could be used to override os.Environ() to arbitrary func
	envGetter func() []string
}

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

type ForEachOp struct {
	Glob *string   `yaml:"glob,omitempty"`
	Item *[]string `yaml:"item,omitempty"`
	// Action to perform for every item
	Action OpSpec `yaml:"action"`
}

// ImportOp reads content of file into data tree at given path
type ImportOp struct {
	// File to read
	File string `yaml:"file"`

	// Path at which to import data.
	// If omitted, then data are merged into root of document
	Path string `yaml:"path"`

	// How to parse file
	Mode ParseFileMode `yaml:"mode,omitempty"`
}

// PatchOp performs RFC6902-style patch on global data document.
// Check patch package for more details
type PatchOp struct {
	Op    patch.Op               `yaml:"op"`
	From  string                 `yaml:"from,omitempty"`
	Path  string                 `yaml:"path"`
	Value map[string]interface{} `yaml:"value,omitempty"`
}

// SetOp sets data in global data document at given path.
type SetOp struct {
	// Arbitrary data to put into data tree
	Data map[string]interface{} `yaml:"data"`

	// Path at which to put data.
	// If omitted, then data are merged into root of document
	Path string `yaml:"path,omitempty"`
}

// TemplateOp can be used to render value from data at runtime.
// Global data tree is available under .Data
type TemplateOp struct {
	// template to render
	Template string `yaml:"template"`
	// path within global data tree where to set result at
	WriteTo string `yaml:"writeTo"`
}
