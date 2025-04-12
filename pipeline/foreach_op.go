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
	"path/filepath"
	"strings"

	"github.com/rkosegi/yaml-toolkit/dom"
)

// ForEachOp can be used to repeat actions over list of items.
// Those items could be
//  1. files specified by globbing pattern
//  2. result of query from data tree (list values)
//  3. specified strings
type ForEachOp struct {
	// Glob is pattern that will be used to match files on file system.
	// Matched files will be used as iteration items.
	Glob *ValOrRef `yaml:"glob,omitempty"`
	// Query is path within the data tree that will be attempted
	Query *ValOrRef `yaml:"query,omitempty"`
	// Item is list of specified strings to iterate over
	Item *ValOrRefSlice `yaml:"item,omitempty"`
	// Action to perform for every item
	Action ActionSpec `yaml:"action"`
	// Variable is name of variable to hold current iteration item.
	// When omitted, default value of "forEach" will be used
	Variable *string `yaml:"var,omitempty"`
}

func (fea *ForEachOp) Do(ctx ActionContext) error {
	if fea.Glob != nil {
		if matches, err := filepath.Glob(fea.Glob.Resolve(ctx)); err != nil {
			return err
		} else {
			for _, match := range matches {
				err = fea.performWithItem(ctx, dom.LeafNode(match))
				if err != nil {
					return err
				}
			}
		}
	} else if fea.Query != nil {
		if n := ctx.Data().Lookup(fea.Query.Resolve(ctx)); n != nil {
			if n.IsList() {
				for _, item := range n.(dom.List).Items() {
					if err := fea.performWithItem(ctx, item); err != nil {
						return err
					}
				}
			} else if n.IsContainer() {
				for item := range n.(dom.Container).Children() {
					if err := fea.performWithItem(ctx, dom.LeafNode(item)); err != nil {
						return err
					}
				}
			} else {
				if err := fea.performWithItem(ctx, n.(dom.Leaf)); err != nil {
					return err
				}
			}
		}
	} else if fea.Item != nil {
		for _, item := range *fea.Item {
			err := fea.performWithItem(ctx, dom.LeafNode(item.Resolve(ctx)))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (fea *ForEachOp) performWithItem(ctx ActionContext, item dom.Node) (err error) {
	vp := "forEach"
	if fea.Variable != nil {
		vp = *fea.Variable
	}
	ctx.Data().AddValue(vp, item)
	defer ctx.Data().Remove(vp)

	for _, act := range fea.Action.Operations.toList() {
		act = act.CloneWith(ctx)
		err = ctx.Executor().Execute(act)
		if err != nil {
			return err
		}
	}
	return ctx.Executor().Execute(fea.Action.Children)
}

func (fea *ForEachOp) String() string {
	var (
		sb    strings.Builder
		parts []string
	)
	sb.WriteString("ForEach[")
	if fea.Glob != nil {
		parts = append(parts, fmt.Sprintf("Glob=%v", *fea.Glob))
	}
	if fea.Item != nil {
		parts = append(parts, fmt.Sprintf("Items=%v", fea.Item))
	}
	if fea.Query != nil {
		parts = append(parts, fmt.Sprintf("Query=%v", *fea.Query))
	}
	sb.WriteString(strings.Join(parts, ","))
	sb.WriteString("]")
	return sb.String()
}

func (fea *ForEachOp) CloneWith(ctx ActionContext) Action {
	cp := new(ForEachOp)
	cp.Glob = fea.Glob
	cp.Item = fea.Item
	cp.Query = fea.Query
	cp.Action = ActionSpec{}.CloneWith(ctx).(ActionSpec)
	return cp
}
