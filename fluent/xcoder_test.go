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
