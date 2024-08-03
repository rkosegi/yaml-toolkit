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
)

var (
	ErrNoDataToSet   = errors.New("no data to set")
	ErrTemplateEmpty = errors.New("template cannot be empty")
	ErrPathEmpty     = errors.New("path cannot be empty")
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
	// Action return reference to actual Action
	Action() Action
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
