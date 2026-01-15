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
	"errors"
	"io"
	"testing"

	"github.com/rkosegi/yaml-toolkit/common"
	"github.com/rkosegi/yaml-toolkit/dom"
	"github.com/stretchr/testify/assert"
)

type failingBiCodec struct{}

func (f failingBiCodec) Encoder() dom.EncoderFunc {
	return func(w io.Writer, v interface{}) error {
		return errors.New("failing bi encoder")
	}
}

func (f failingBiCodec) Decoder() dom.DecoderFunc {
	return func(r io.Reader, v interface{}) error {
		return errors.New("failing bi decoder")
	}
}

func TestTranscode(t *testing.T) {
	var (
		err error
		x   bytes.Buffer
	)
	type Inner struct {
		MyField1 string `json:"my-field_1"`
		MyField2 int    `json:"my_field2"`
	}
	type data1 struct {
		Inner Inner `json:"inner data"`
	}

	t.Run("json to yaml as object", func(t *testing.T) {
		x.Reset()
		in := &data1{
			Inner: Inner{
				MyField1: "Hello",
				MyField2: 42,
			},
		}

		err = TranscodeJson2Yaml[data1](in, &x)
		assert.NoError(t, err)

		var node dom.Node
		node, err = dom.DecodeReader(&x, dom.DefaultYamlDecoder)
		assert.NoError(t, err)
		assert.NotNil(t, node)

		assert.True(t, node.IsContainer())
		assert.True(t, node.AsContainer().Child("inner data").IsContainer())
		assert.Equal(t, "Hello", node.AsContainer().
			Child("inner data").AsContainer().
			Child("my-field_1").AsLeaf().Value())
		assert.Equal(t, 42, node.AsContainer().
			Child("inner data").AsContainer().
			Child("my_field2").AsLeaf().Value())
	})

	t.Run("json to yaml as list", func(t *testing.T) {
		x.Reset()

		in := &[]Inner{
			{
				MyField1: "Hi",
				MyField2: 42,
			},
			{
				MyField1: "Ola",
				MyField2: 350,
			},
		}

		err = TranscodeJson2Yaml[[]Inner](in, &x)
		assert.NoError(t, err)
		var node dom.Node
		node, err = dom.DecodeReader(&x, dom.DefaultYamlDecoder)
		assert.NoError(t, err)
		assert.NotNil(t, node)

		assert.True(t, node.IsList())
		assert.Len(t, node.AsList().Items(), 2)
		assert.Equal(t, "Ola", node.AsList().Get(1).AsContainer().Child("my-field_1").AsLeaf().Value())
	})

	t.Run("negative cases", func(t *testing.T) {
		assert.Error(t, Transcode[data1](&data1{}, func(w io.Writer, v interface{}) error {
			return errors.New("")
		}, dom.DefaultYamlDecoder, dom.DefaultJsonEncoder, common.FailingWriter()))

		assert.Error(t, Transcode[data1](&data1{}, dom.DefaultJsonEncoder, func(r io.Reader, v interface{}) error {
			return errors.New("this is an error")
		}, dom.DefaultJsonEncoder, common.FailingWriter()))
	})
}

func TestTransform(t *testing.T) {
	type MyType struct {
		A string `json:"a"`
		B int
	}
	t.Run("list of map[string]interface{} to list of specific types", func(t *testing.T) {
		in := []interface{}{
			map[string]interface{}{
				"a": "hello",
				"b": 42,
			},
			map[string]interface{}{
				"a": "world",
				"b": 50,
			},
		}
		out, err := Transform[[]MyType](in, dom.DefaultJsonCodec())
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.Equal(t, "hello world", (*out)[0].A+" "+(*out)[1].A)
		assert.Equal(t, 92, (*out)[0].B+(*out)[1].B)
	})
	t.Run("invalid read", func(t *testing.T) {
		_, err := Transform[any](struct{}{}, &failingBiCodec{})
		assert.Error(t, err)
	})

	t.Run("must transform", func(t *testing.T) {
		t.Run("valid", func(t *testing.T) {
			mt := MustTransform[MyType](map[string]interface{}{
				"a": "hello",
				"b": 42,
				"c": 3.14,
			}, dom.DefaultYamlCodec())
			assert.NotNil(t, mt)
			assert.Equal(t, 42, mt.B)
			assert.Equal(t, "hello", mt.A)
		})
		t.Run("invalid", func(t *testing.T) {
			defer func() {
				recover()
			}()
			MustTransform[MyType]([]int{0}, dom.DefaultYamlCodec())
			assert.Fail(t, "transform must fail")
		})
	})
	t.Run("transform slice", func(t *testing.T) {
		t.Run("valid", func(t *testing.T) {
			mt := MustTransformSlice[map[string]interface{}, MyType]([]map[string]interface{}{
				{"a": "hello"},
				{"a": "world"},
			}, dom.DefaultYamlCodec())
			assert.NotNil(t, mt)
			assert.Len(t, mt, 2)
			assert.Equal(t, "hello", mt[0].A)
			assert.Equal(t, "world", mt[1].A)
		})
		t.Run("invalid", func(t *testing.T) {
			defer func() {
				recover()
			}()
			MustTransformSlice[string, MyType]([]string{"A"}, failingBiCodec{})
			assert.Fail(t, "transform slice must fail")
		})
	})
}
