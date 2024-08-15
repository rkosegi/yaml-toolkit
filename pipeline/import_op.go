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
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/rkosegi/yaml-toolkit/dom"
	"github.com/rkosegi/yaml-toolkit/props"
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

func (pfm ParseFileMode) toValue(content []byte) (dom.Node, error) {
	switch pfm {
	case ParseFileModeBinary:
		return dom.LeafNode(base64.StdEncoding.EncodeToString(content)), nil
	case ParseFileModeText:
		return dom.LeafNode(string(content)), nil
	case ParseFileModeYaml:
		return b.FromReader(bytes.NewReader(content), dom.DefaultYamlDecoder)
	case ParseFileModeJson:
		return b.FromReader(bytes.NewReader(content), dom.DefaultJsonDecoder)
	case ParseFileModeProperties:
		return b.FromReader(bytes.NewReader(content), props.DecoderFn)
	default:
		return nil, fmt.Errorf("invalid ParseFileMode: %v", pfm)
	}
}

func (ia *ImportOp) String() string {
	return fmt.Sprintf("Import[file=%s,path=%s,mode=%s]", ia.File, ia.Path, ia.Mode)
}

func (ia *ImportOp) Do(ctx ActionContext) error {
	val, err := parseFile(ctx.TemplateEngine().RenderLenient(ia.File, ctx.Snapshot()), ia.Mode)
	if err != nil {
		return err
	}
	p := ctx.TemplateEngine().RenderLenient(ia.Path, ctx.Snapshot())
	if len(p) > 0 {
		ctx.Data().AddValueAt(p, val)
	} else {
		if !val.IsContainer() {
			return ErrNotContainer
		} else {
			for k, v := range val.(dom.Container).Children() {
				ctx.Data().AddValueAt(k, v)
			}
		}
	}
	return nil
}

func (ia *ImportOp) CloneWith(ctx ActionContext) Action {
	return &ImportOp{
		Mode: ia.Mode,
		Path: ctx.TemplateEngine().RenderLenient(ia.Path, ctx.Snapshot()),
		File: ctx.TemplateEngine().RenderLenient(ia.File, ctx.Snapshot()),
	}
}
