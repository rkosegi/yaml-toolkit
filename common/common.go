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

package common

import (
	"path/filepath"

	"github.com/rkosegi/yaml-toolkit/dom"
	"github.com/rkosegi/yaml-toolkit/props"
)

type FileEncoderProvider func(file string) dom.EncoderFunc

// DefaultFileEncoderProvider is FileEncoderProvider that uses file suffix to choose dom.EncoderFunc
func DefaultFileEncoderProvider(file string) dom.EncoderFunc {
	switch filepath.Ext(file) {
	case ".yaml", ".yml":
		return dom.DefaultYamlEncoder
	case ".json":
		return dom.DefaultJsonEncoder
	case ".properties":
		return props.EncoderFn
	default:
		return nil
	}
}

// FileDecoderProvider resolves dom.DecoderFunc for given file.
// If file is not recognized, nil is returned.
type FileDecoderProvider func(file string) dom.DecoderFunc

// DefaultFileDecoderProvider is FileDecoderProvider that uses file suffix to choose dom.DecoderFunc
func DefaultFileDecoderProvider(file string) dom.DecoderFunc {
	switch filepath.Ext(file) {
	case ".yaml", ".yml":
		return dom.DefaultYamlDecoder
	case ".json":
		return dom.DefaultJsonDecoder
	case ".properties":
		return props.DecoderFn
	default:
		return nil
	}
}
