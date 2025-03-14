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

package diff

import (
	"bytes"
	"strings"
	"testing"

	"github.com/rkosegi/yaml-toolkit/dom"
	"github.com/stretchr/testify/assert"
)

func TestApply(t *testing.T) {
	doc, err := dom.Builder().FromReader(strings.NewReader(`
leaf0: 1234
leafX: null
level1:
  level2: 123
`), dom.DefaultYamlDecoder)
	if err != nil {
		t.Fatal(err)
	}
	Apply(doc, []Modification{
		{
			Type:  ModAdd,
			Path:  "level1.level22.leaf22",
			Value: "abc",
		},
	})
	var buf bytes.Buffer
	err = doc.Serialize(&buf, dom.DefaultNodeEncoderFn, dom.DefaultYamlEncoder)
	assert.Nil(t, err)
	assert.Equal(t, `leaf0: 1234
leafX: null
level1:
  level2: 123
  level22:
    leaf22: abc
`, buf.String())
}

func TestApplyListElements(t *testing.T) {
	doc, err := dom.Builder().FromReader(strings.NewReader(`
leafX: null
list1:
  - abc
`), dom.DefaultYamlDecoder)
	if err != nil {
		t.Fatal(err)
	}
	Apply(doc, []Modification{
		{
			Type:  ModAdd,
			Path:  "list1[1][2][3].item_obj1.sub.sublist[1].efgh[0]",
			Value: "abc",
		},
	})
	assert.True(t, doc.Child("list1").(dom.List).Items()[1].IsList())
}

func TestApplyNoop(t *testing.T) {
	doc, err := dom.Builder().FromReader(strings.NewReader(`
leaf0: 1234
level1:
  level2: 123
`), dom.DefaultYamlDecoder)
	if err != nil {
		t.Fatal(err)
	}
	Apply(doc, []Modification{
		{
			Type:  ModDelete,
			Path:  "level1.level23.leaf22",
			Value: nil,
		},
	})
	var buf bytes.Buffer
	err = doc.Serialize(&buf, dom.DefaultNodeEncoderFn, dom.DefaultYamlEncoder)
	assert.Nil(t, err)
	assert.Equal(t, `leaf0: 1234
level1:
  level2: 123
`, buf.String())
}
