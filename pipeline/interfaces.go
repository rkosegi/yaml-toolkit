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
	"io"
	"iter"

	"github.com/rkosegi/yaml-toolkit/dom"
	te "github.com/rkosegi/yaml-toolkit/pipeline/template_engine"
)

// Executor interface is used by external callers to execute arbitrary Action or run full PipelineOp.
type Executor interface {
	// Run runs pipeline
	Run(po *PipelineOp) error

	// Execute executes single Action
	Execute(act Action) error

	// Runtime gets reference to RuntimeServices
	Runtime() RuntimeServices
}

// Logger interface allows arbitrary messages to be logged by actions.
type Logger interface {
	// Log logs given values.
	// Format of values passed to this method is undefined.
	Log(v ...interface{})
}

// Listener interface allows to hook into execution of Action.
type Listener interface {
	// OnBefore is called just before action is executed
	OnBefore(ctx ActionContext)
	// OnAfter is called sometime after action is executed, regardless of result.
	// Any error returned by invoking Do() method is returned as last parameter.
	OnAfter(ctx ActionContext, err error)
	// OnLog is called whenever action invokes Log method on Logger instance
	OnLog(ctx ActionContext, v ...interface{})
}

// ExtInterface enables actions to use extensions.
type ExtInterface interface {
	// DefineAction defines named ActionSpec in action registry.
	DefineAction(string, ActionSpec)
	// GetAction retrieves previously registered ActionSpec from registry.
	GetAction(string) (ActionSpec, bool)
	// RegisterActionFactory adds named ActionFactory to registry, so it can be retrieved later via GetActionFactory() call.
	RegisterActionFactory(string, ActionFactory)
	// GetActionFactory retrieves previously added ActionFactory of given name.
	GetActionFactory(string) ActionFactory
	// RegisterService registers named service with registry.
	RegisterService(string, Service)
	// GetService gets a reference to previously registered service from registry
	GetService(string) Service
	// EnumServices allows to enumerate all registered services
	EnumServices() iter.Seq[Service]
}

// ActionContext is created by Executor implementation for sole purpose of invoking Action's Do function.
type ActionContext interface {
	ClientContext
	// Action return reference to actual Action
	Action() Action
	// Executor returns reference to executor
	Executor() Executor
}

type ServiceContext interface {
	ClientContext
}

type ClientContext interface {
	DataServices
	RuntimeServices
	// Logger gets reference to Logger interface
	Logger() Logger
}

type DataServices interface {
	// Data exposes data document
	Data() dom.ContainerBuilder
	// Snapshot is read-only view of Data() in point in time.
	// This value is cached for performance reasons.
	Snapshot() map[string]interface{}
	// InvalidateSnapshot marks cached snapshot as dirty so next snapshot request will cause retrieval.
	InvalidateSnapshot()
	// Factory give access to factory to create new documents
	Factory() dom.ContainerFactory
}

type RuntimeServices interface {
	// TemplateEngine return reference to TemplateEngine
	TemplateEngine() te.TemplateEngine

	// Ext allows access to extensions interface
	Ext() ExtInterface
}

type Service interface {
	// Configure configures this instance of Service.
	// Returned value must be non-nil to allow for fluent-style invocation.
	Configure(ctx ServiceContext, cfg StrKeysAnyValues) Service
	// Init initializes this instance.
	// Implementation of this function should be idempotent (so runtime can call it multiple times without any problem).
	Init() error
	io.Closer
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
	NewForArgs(args StrKeysAnyValues) Action
}
