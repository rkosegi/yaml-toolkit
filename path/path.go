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

import "strconv"

type path struct {
	components []component
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
	for i := range p.components {
		c[i] = p.components[i]
	}
	return c
}

func AfterLast() AppendOpt {
	return func(c *component) {
		c.afterLast = true
		c.value = "-"
	}
}

func Wildcard() AppendOpt {
	return func(c *component) {
		c.wildcard = true
		c.isNumeric = false
		c.afterLast = false
	}
}

func Numeric(val int) AppendOpt {
	return func(c *component) {
		c.value = strconv.Itoa(val)
		c.isNumeric = true
		c.wildcard = false
		c.afterLast = false
		c.num = val
	}
}

func Simple(value string) AppendOpt {
	return func(c *component) {
		c.value = value
	}
}
