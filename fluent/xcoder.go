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

package fluent

import (
	"bytes"
	"io"

	"github.com/rkosegi/yaml-toolkit/dom"
)

// Transcode type into one serialized form then to another.
// Common use case for this code is when you have a type that has JSON tags with custom serialized names,
// no YAML tags. When you serialize such object directly into YAML, custom field names are lost.
func Transcode[T any](in *T, enc dom.EncoderFunc, dec dom.DecoderFunc, outEnc dom.EncoderFunc, w io.Writer) error {
	var (
		data bytes.Buffer
		err  error
	)
	// 1, encode T into first serialized form.
	// This step will honor custom tags, so field naming is retained.
	if err = enc(&data, in); err != nil {
		return err
	}
	// 2, decode serialized form into dummy object
	var x interface{}
	if err = dec(&data, &x); err != nil {
		return err
	}
	// 3, encode dummy object into final, desired form
	return outEnc(w, x)
}

// TranscodeJson2Yaml takes T and encodes it into JSON form,
// then decodes it into intermediate object which is then serialized into YAML form.
func TranscodeJson2Yaml[T any](in *T, w io.Writer) error {
	return Transcode[T](in, dom.DefaultJsonEncoder, dom.DefaultJsonDecoder, dom.DefaultYamlEncoder, w)
}

// Transform can transform arbitrary object to specific type, using provided codec.
// Common use case is to transform i.e. []interface{} to []MyType
func Transform[T any](in any, codec dom.FormatBiCodec) (*T, error) {
	var (
		data bytes.Buffer
		err  error
		out  T
	)
	if err = codec.Encoder()(&data, in); err != nil {
		return nil, err
	}
	err = codec.Decoder()(&data, &out)
	return &out, err
}
