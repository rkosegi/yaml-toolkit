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
	"strings"

	"github.com/rkosegi/yaml-toolkit/dom"
	"gopkg.in/yaml.v3"
)

var (
	ErrNoDataToSet   = errors.New("no data to set")
	ErrTemplateEmpty = errors.New("template cannot be empty")
	ErrPathEmpty     = errors.New("path cannot be empty")
	ErrFileEmpty     = errors.New("file cannot be empty")
	ErrOutputEmpty   = errors.New("output cannot be empty")
	ErrNotContainer  = errors.New("data element must be container when no path is provided")
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
	// RenderMapLenient attempts to render every leaf value in provided map
	RenderMapLenient(input map[string]interface{}, data map[string]interface{}) map[string]interface{}
	EvalBool(template string, data map[string]interface{}) (bool, error)
}

// Logger interface allows arbitrary messages to be logged by actions.
type Logger interface {
	// Log logs given values.
	// Format of values passed to this method is undefined.
	Log(v ...interface{})
}

// Listener interface allows hook into execution of Action.
type Listener interface {
	// OnBefore is called just before action is executed
	OnBefore(ctx ActionContext)
	// OnAfter is called sometime after action is executed, regardless of result.
	// Any error returned by invoking Do() method is returned as last parameter.
	OnAfter(ctx ActionContext, err error)
	// OnLog is called whenever action invokes Log method on Logger instance
	OnLog(ctx ActionContext, v ...interface{})
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
	// Action return reference to actual Action
	Action() Action
	// Logger gets reference to Logger interface
	Logger() Logger
	// Ext allows access to extensions interface
	Ext() ExtInterface
}

// Action is implemented by actions within ActionSpec
type Action interface {
	fmt.Stringer
	// Do will perform this action.
	// This function is invoked by Executor implementation and as such it's not meant to be called by end user directly.
	Do(ctx ActionContext) error
	Cloneable
}

// Cloneable interface allows to customize default clone behavior by providing implementation of CloneWith function.
type Cloneable interface {
	// CloneWith creates fresh clone of this Action with values of its fields templated.
	// Data for template can be obtained by calling Snapshot() on provided context.
	CloneWith(ctx ActionContext) Action
}

// ActionFactory can be used to create instances of Action
type ActionFactory interface {
	// NewForArgs creates new instance of Action for given set of arguments.
	NewForArgs(args map[string]interface{}) Action
}

// ArgsSetter is implemented by ext Action if it wishes to receive arguments.
type ArgsSetter interface {
	// SetArgs sets arguments from map
	SetArgs(args map[string]interface{})
}

// AnyVal can represent any DOM value (leaf, list, container)
type AnyVal struct {
	v dom.Node
}

func (pv *AnyVal) UnmarshalYAML(node *yaml.Node) error {
	pv.v = dom.YamlNodeDecoder()(node)
	return nil
}

// Value get actual value
func (pv *AnyVal) Value() dom.Node {
	return pv.v
}

// ValOrRef is either immediate leaf value or reference to a dom.Leaf in data tree at given path.
// If Path references non-existent node, or node pointed to is not a dom.Leaf, empty value is returned
type ValOrRef struct {
	isRef bool
	// Ref is resolved reference, is any
	Ref string
	// Val is value of dom.Leaf pointed to by ref after Resolve(ctx) has been called, or immediate value
	// if Ref is empty
	Val string
}

func (pv *ValOrRef) CloneWith(ctx ActionContext) *ValOrRef {
	return &ValOrRef{
		isRef: pv.isRef,
		Ref:   ctx.TemplateEngine().RenderLenient(pv.Ref, ctx.Snapshot()),
		Val:   ctx.TemplateEngine().RenderLenient(pv.Val, ctx.Snapshot()),
	}
}

func (pv *ValOrRef) UnmarshalYAML(node *yaml.Node) error {
	switch node.Kind {
	case yaml.MappingNode:
		m := make(map[string]interface{})
		_ = node.Decode(&m)
		if x, ok := m["ref"]; !ok {
			return errors.New("missing 'ref' field")
		} else {
			pv.isRef = true
			pv.Ref = x.(string)
		}
		return nil

	case yaml.ScalarNode:
		pv.Val = node.Value
		return nil
	}
	return fmt.Errorf("invalid value: '%v'", node)
}

func (pv *ValOrRef) Resolve(ctx ActionContext) string {
	ss := ctx.Snapshot()
	if pv.isRef {
		if n := ctx.Data().Lookup(pv.Ref); n != nil && n.IsLeaf() {
			v := fmt.Sprintf("%v", n.(dom.Leaf).Value())
			return ctx.TemplateEngine().RenderLenient(v, ss)
		}
		return ""
	} else {
		return ctx.TemplateEngine().RenderLenient(pv.Val, ss)
	}
}

func (pv *ValOrRef) String() string {
	return fmt.Sprintf("[Ref=%v,Val=%v]", pv.Ref, pv.Val)
}

type ValOrRefSlice []*ValOrRef

func (pv *ValOrRefSlice) String() string {
	var strs []string
	for _, val := range *pv {
		strs = append(strs, val.String())
	}
	return "[" + strings.Join(strs, ",") + "]"
}

func anyValFromMap(m map[string]interface{}) *AnyVal {
	return &AnyVal{v: dom.DefaultNodeDecoderFn(m)}
}

// ChildActions is map of named actions that are executed as a part of parent action
type ChildActions map[string]ActionSpec
