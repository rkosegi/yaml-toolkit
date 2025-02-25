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

	"github.com/rkosegi/yaml-toolkit/dom"
)

type ForEachOp struct {
	Glob *string   `yaml:"glob,omitempty"`
	Item *[]string `yaml:"item,omitempty"`
	// Action to perform for every item
	Action ActionSpec `yaml:"action"`
}

func (fea *ForEachOp) Do(ctx ActionContext) error {
	if nonEmpty(fea.Glob) {
		if matches, err := filepath.Glob(*fea.Glob); err != nil {
			return err
		} else {
			for _, match := range matches {
				err = fea.performWithItem(ctx, match)
				if err != nil {
					return err
				}
			}
		}
	} else if fea.Item != nil {
		for _, item := range *fea.Item {
			err := fea.performWithItem(ctx, item)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (fea *ForEachOp) performWithItem(ctx ActionContext, item string) (err error) {
	ctx.Data().AddValue("forEach", dom.LeafNode(item))
	defer ctx.Data().Remove("forEach")

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
	return fmt.Sprintf("ForEach[Glob=%s,Items=%d]", safeStrDeref(fea.Glob), safeStrListSize(fea.Item))
}

func (fea *ForEachOp) CloneWith(ctx ActionContext) Action {
	cp := new(ForEachOp)
	cp.Glob = fea.Glob
	cp.Item = fea.Item
	cp.Action = ActionSpec{}.CloneWith(ctx).(ActionSpec)
	return cp
}
