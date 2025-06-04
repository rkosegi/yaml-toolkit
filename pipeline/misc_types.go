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
	"strings"

	"github.com/rkosegi/yaml-toolkit/dom"
	te "github.com/rkosegi/yaml-toolkit/pipeline/template_engine"
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
	// Ref is resolved reference, if any
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
			v := fmt.Sprintf("%v", n.AsLeaf().Value())
			return ctx.TemplateEngine().RenderLenient(v, ss)
		}
		return ""
	} else {
		return ctx.TemplateEngine().RenderLenient(pv.Val, ss)
	}
}

func (pv *ValOrRef) String() string {
	var (
		sb    strings.Builder
		parts []string
	)
	sb.WriteByte('[')
	if len(pv.Ref) > 0 {
		parts = append(parts, fmt.Sprintf("Ref=%v", pv.Ref))
	}
	if len(pv.Val) > 0 {
		parts = append(parts, fmt.Sprintf("Val=%v", pv.Val))
	}
	sb.WriteString(strings.Join(parts, ","))
	sb.WriteByte(']')
	return sb.String()
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

// StrKeysStrValues is key-value map with strings.
type StrKeysStrValues map[string]string

// AsAnyValuesMap converts this map to StrKeysAnyValues suitable for template related operations.
func (skv StrKeysStrValues) AsAnyValuesMap() StrKeysAnyValues {
	out := make(StrKeysAnyValues, len(skv))
	for k, v := range skv {
		out[k] = v
	}
	return out
}

// RenderValues renders values of this map using provided TemplateEngine and data.
func (skv StrKeysStrValues) RenderValues(teng te.TemplateEngine, data StrKeysAnyValues) StrKeysStrValues {
	out := make(map[string]string, len(skv))
	for k, v := range skv {
		out[k] = teng.RenderLenient(v, data)
	}
	return out
}

// StrKeysAnyValues is a map with string keys and values with any type
type StrKeysAnyValues map[string]any
