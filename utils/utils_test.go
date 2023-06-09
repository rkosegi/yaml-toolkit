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
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestToPath(t *testing.T) {
	assert.Equal(t, "abc", ToPath("", "abc"))
	assert.Equal(t, "abc.def", ToPath("abc", "def"))
}

func TestNewYamlEncoder(t *testing.T) {
	assert.NotNil(t, NewYamlEncoder(bytes.NewBuffer(make([]byte, 0))))
}
