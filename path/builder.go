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
	components []component
}

func (b *builder) Reset() Builder {
	b.components = nil
	return b
}

func (b *builder) Build() Path {
	c := make([]component, len(b.components))
	copy(c, b.components)
	return &path{components: c}
}

func (b *builder) Append(opts ...AppendOpt) Builder {
	b.components = append(b.components, *buildComponent(opts...))
	return b
}

// NewBuilder creates new Builder
func NewBuilder() Builder {
	return &builder{}
}

func buildComponent(opts ...AppendOpt) *component {
	if len(opts) == 0 {
		panic("no append option provided by caller")
	}
	c := &component{}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func BuildComponent(opts ...AppendOpt) Component {
	return buildComponent(opts...)
}
