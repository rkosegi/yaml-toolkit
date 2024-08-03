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
)

type setHandlerFn func(path string, orig, other dom.ContainerBuilder)

func setOpMergeIfContainersReplaceOtherwise(orig, other dom.ContainerBuilder) {
	for k, v := range other.Children() {
		origChild := orig.Child(k)
		if origChild != nil && origChild.IsContainer() && v.IsContainer() {
			orig.AddValue(k, origChild.(dom.ContainerBuilder).Merge(v.(dom.Container)))
		} else {
			orig.AddValue(k, v)
		}
	}
}

var setHandlerFnMap = map[SetStrategy]setHandlerFn{
	SetStrategyMerge: func(path string, orig, other dom.ContainerBuilder) {
		if len(path) > 0 {
			dest := orig.Lookup(path)
			if dest != nil && dest.IsContainer() {
				orig.AddValueAt(path, dest.(dom.ContainerBuilder).Merge(other))
			} else {
				orig.AddValueAt(path, other)
			}
		} else {
			setOpMergeIfContainersReplaceOtherwise(orig, other)
		}
	},
	SetStrategyReplace: func(path string, orig, other dom.ContainerBuilder) {
		if len(path) > 0 {
			orig.AddValueAt(path, other)
		} else {
			for k, v := range other.Children() {
				orig.AddValueAt(k, v)
			}
		}
	},
}

type SetStrategy string

const (
	SetStrategyReplace = SetStrategy("replace")
	SetStrategyMerge   = SetStrategy("merge")
)

// SetOp sets data in global data document at given path.
type SetOp struct {
	// Arbitrary data to put into data tree
	Data map[string]interface{} `yaml:"data"`

	// Path at which to put data.
	// If omitted, then data are merged into root of document
	Path string `yaml:"path,omitempty"`

	// Strategy defines how that are handled when conflict during set/add of data occur.
	Strategy *SetStrategy `yaml:"strategy,omitempty"`
}

func (sa *SetOp) String() string {
	return fmt.Sprintf("Set[Path=%s]", sa.Path)
}

func (sa *SetOp) Do(ctx ActionContext) error {
	gd := ctx.Data()
	if sa.Data == nil {
		return ErrNoDataToSet
	}
	if sa.Strategy == nil {
		sa.Strategy = setStrategyPointer(SetStrategyMerge)
	}
	handler, exists := setHandlerFnMap[*sa.Strategy]
	if !exists {
		return fmt.Errorf("SetOp: unknown SetStrategy %s", *sa.Strategy)
	}
	data := ctx.Factory().FromMap(sa.Data)
	handler(sa.Path, gd, data)
	return nil
}

func (sa *SetOp) CloneWith(ctx ActionContext) Action {
	return &SetOp{
		Data: sa.Data,
		Path: ctx.TemplateEngine().RenderLenient(sa.Path, ctx.Snapshot()),
	}
}
