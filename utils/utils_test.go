/*
Copyright 2023 Richard Kosegi

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

package utils

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToPath(t *testing.T) {
	assert.Equal(t, "abc", ToPath("", "abc"))
	assert.Equal(t, "abc.def", ToPath("abc", "def"))
}

func TestNewYamlEncoder(t *testing.T) {
	assert.NotNil(t, NewYamlEncoder(bytes.NewBuffer(make([]byte, 0))))
}

func TestParseListPathComponent(t *testing.T) {
	name, indexes, ok := ParseListPathComponent("list[0][3][1]")
	assert.True(t, ok)
	assert.Equal(t, 3, len(indexes))
	assert.Equal(t, 0, indexes[0])
	assert.Equal(t, 3, indexes[1])
	assert.Equal(t, 1, indexes[2])
	assert.Equal(t, "list", name)
	_, _, ok = ParseListPathComponent("not a list")
	assert.False(t, ok)
}

func TestToListPath(t *testing.T) {
	assert.Equal(t, "[2]", ToListPath("", 2))
	assert.Equal(t, "item[3]", ToListPath("item", 3))
}

func TestFailingReader(t *testing.T) {
	_, err := FailingReader().Read([]byte{})
	assert.Error(t, err)
}

func TestFailingWriter(t *testing.T) {
	_, err := FailingWriter().Write([]byte{})
	assert.Error(t, err)
}

func TestUnique(t *testing.T) {
	assert.Equal(t, 3, len(Unique([]string{"a", "a", "b", "x", "x"})))
}

func TestUnflattenMap(t *testing.T) {
	out := Unflatten(map[string]interface{}{
		"a.b": 123,
		"a.c": "0876",
		"x.y": "hello",
		"x.z": nil,
		"s":   3.14,
	})
	assert.Equal(t, 123, out["a"].(map[string]interface{})["b"])
	assert.Equal(t, "0876", out["a"].(map[string]interface{})["c"])
	assert.Equal(t, "hello", out["x"].(map[string]interface{})["y"])
	assert.Equal(t, nil, out["x"].(map[string]interface{})["z"])
	assert.Equal(t, 3.14, out["s"])
}
