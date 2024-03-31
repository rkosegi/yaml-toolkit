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

package patch

import (
	"github.com/rkosegi/yaml-toolkit/dom"
	"strings"
)

var (
	b = dom.Builder()
)

func makeTestContainer() dom.ContainerBuilder {
	c, _ := b.FromReader(strings.NewReader(`
root:
  list:
    - item1
    - sub20: 123
      sub21: 456
    - item3
    - item4
  sub1:
    prop: 456
`), dom.DefaultYamlDecoder)
	return c
}
