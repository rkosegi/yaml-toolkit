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

type component struct {
	value string
	// rfc6901 - pointer after last list item "-"
	afterLast bool
	wildcard  bool
	num       int
	isNumeric bool
}

func (c component) IsInsertAfterLast() bool {
	return c.afterLast
}

func (c component) IsNumeric() bool {
	return c.isNumeric
}

func (c component) NumericValue() int {
	return c.num
}

func (c component) IsWildcard() bool {
	return c.wildcard
}

func (c component) Value() string {
	return c.value
}
