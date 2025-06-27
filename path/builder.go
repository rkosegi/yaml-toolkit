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

type builder struct {
	components []Component
}

var emptyPath = &path{}

func (b *builder) Build() Path {
	return &path{components: b.components}
}

func (b *builder) Append(c Component) Builder {
	cs := make([]Component, len(b.components))
	copy(cs, b.components)
	cs = append(b.components, c)
	return &builder{components: cs}
}

// NewBuilder creates new Builder
func NewBuilder() Builder {
	return &builder{}
}

// ParentOf returns Path with last component stripped out.
// If given Path is empty, nil is returned.
func ParentOf(p Path) Path {
	switch len(p.Components()) {
	case 0:
		return nil
	case 1:
		return emptyPath
	default:
		c := make([]Component, len(p.(*path).components)-1)
		copy(c, p.(*path).components[:len(p.(*path).components)-1])
		return &path{components: c}
	}
}

// ChildOf creates path based on parent Path with additional child path Components.
func ChildOf(parent Path, cps ...Component) Path {
	cs := make([]Component, len(parent.(*path).components))
	copy(cs, parent.(*path).components)
	cs = append(cs, cps...)
	return &path{components: cs}
}
