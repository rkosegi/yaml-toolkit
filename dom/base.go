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

package dom

type base struct {
	desc string
}

func (b *base) Desc() string {
	return b.desc
}

func (b *base) IsContainer() bool {
	return false
}

func (b *base) IsLeaf() bool {
	return false
}

func (b *base) IsList() bool {
	return false
}

func (b *base) AsLeaf() Leaf {
	panic("not a leaf: " + b.Desc())
}

func (b *base) AsContainer() Container {
	panic("not a container: " + b.Desc())
}

func (b *base) AsList() List {
	panic("not a list: " + b.Desc())
}
