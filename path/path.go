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

package path

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"
)

type path struct {
	components []Component
	s          string
}

func (p path) Last() Component {
	if len(p.components) == 0 {
		panic("empty path")
	}
	return p.components[len(p.components)-1]
}

func (p path) IsEmpty() bool {
	return len(p.components) == 0
}

func (p path) Components() []Component {
	c := make([]Component, len(p.components))
	copy(c, p.components)
	return c
}

func (p path) buildString() string {
	pcs := len(p.components)
	if pcs == 0 {
		return "[]"
	}
	var r []interface{}
	for i := 0; i < pcs; i++ {
		if p.components[i].IsNumeric() {
			r = append(r, p.components[i].NumericValue())
		} else {
			r = append(r, p.components[i].Value())
		}
	}
	// maybe use MarshalJSON() here?
	var out bytes.Buffer
	_ = json.NewEncoder(&out).Encode(r)
	return strings.TrimSpace(out.String())
}

// String returns string representation of this path, that can be used as a key into map.
func (p path) String() string {
	return p.s
}

func AfterLast() Component {
	return &component{
		afterLast: true,
		value:     "-",
	}
}

func Numeric(val int) Component {
	return &component{
		value:     strconv.Itoa(val),
		isNumeric: true,
		afterLast: false,
		num:       val,
	}
}

func Simple(value string) Component {
	return &component{
		value: value,
	}
}
