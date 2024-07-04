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

// Builder provides convenient way to construct Path using fluent builder pattern.
type Builder interface {
	// Append adds path component to the end of path.
	Append(opts ...AppendOpt) Builder

	// Reset clears internal state of builder.
	// Path instances created using this builder previously are unaffected by this operation.
	Reset() Builder

	// Build creates Path using current state.
	// It's safe to call this function multiple times or re-use it afterward,
	// it will always create fresh Path everytime.
	Build() Path
}

type Component interface {
	// IsInsertAfterLast returns true if this path component points to non-existent item after last element in the list.
	// This is required by JSON pointer (rfc6901) during append to the end of list.
	IsInsertAfterLast() bool

	// IsNumeric returns true if name is numeric value according to rfc6901
	IsNumeric() bool

	// NumericValue gets number that points to array element with the zero-based index.
	// Only valid if IsNumeric returns true.
	NumericValue() int

	// IsWildcard returns true if this path element represent wildcard match, ie equal to any value.
	// This is used e.g. in property paths such as "x.y.z.*.w"
	IsWildcard() bool

	// Value returns canonical value of this component.
	Value() string
}

type Path interface {
	// Components returns copy of path components in this path
	Components() []Component

	// IsEmpty returns true if Path does not have any components.
	IsEmpty() bool

	// Last gets very last path Component, panics if path is empty.
	Last() Component
}

type AppendOpt func(*component)

// Parser interface is implemented by different Path syntax parsers.
type Parser interface {
	// Parse parses source string into Path.
	// Any error encountered during parse is returned to caller.
	Parse(string) (Path, error)
	// MustParse parses source string into Path.
	// Any error encountered during parse will cause panic.
	MustParse(string) Path
}

// Serializer is interface that allows to serialize Path into lexical form
type Serializer interface {
	// Serialize serializes path into lexical representation
	Serialize(Path) string
}
