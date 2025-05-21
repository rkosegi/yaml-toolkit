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
	"errors"
	"fmt"
	"strconv"

	"github.com/rkosegi/yaml-toolkit/dom"
)

type (
	LeafValueFormat string

	ConvertOp struct {
		// Path to the leaf value within the data which will be converted.
		Path string

		// Format specifies desired format for value.
		Format LeafValueFormat
	}
)

const (
	LeafValueFormatInt64   = LeafValueFormat("int64")
	LeafValueFormatFloat64 = LeafValueFormat("float64")
)

func (c *ConvertOp) String() string {
	return fmt.Sprintf("Convert[Path: %s, Format: %s]", c.Path, c.Format)
}

func (c *ConvertOp) Do(ctx ActionContext) error {
	if len(c.Format) == 0 {
		return errors.New("format must not be empty")
	}
	if len(c.Path) == 0 {
		return ErrPathEmpty
	}
	node := ctx.Data().Lookup(c.Path)
	if node != nil && node.IsLeaf() && node.AsLeaf().Value() != nil {
		val := fmt.Sprintf("%v", node.AsLeaf().Value())
		switch c.Format {
		case LeafValueFormatFloat64:
			if x, err := strconv.ParseFloat(val, 64); err != nil {
				return err
			} else {
				ctx.Data().AddValueAt(c.Path, dom.LeafNode(x))
			}
		case LeafValueFormatInt64:
			if x, err := strconv.ParseInt(val, 10, 64); err != nil {
				return err
			} else {
				ctx.Data().AddValueAt(c.Path, dom.LeafNode(x))
			}
		default:
			return fmt.Errorf("unknown format: %v", c.Format)
		}
	}
	return nil
}

func (c *ConvertOp) CloneWith(ctx ActionContext) Action {
	return &ConvertOp{
		Path:   ctx.TemplateEngine().RenderLenient(c.Path, ctx.Snapshot()),
		Format: c.Format,
	}
}
