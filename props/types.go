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

package props

// LookupFn looks up value corresponding to given key.
type LookupFn func(key string) *string

// Resolver allows to resolve placeholder in property values. Nested placeholders are supported.
type Resolver interface {
	// Resolve resolves placeholder. Panics if there is circular reference.
	Resolve(in string) string
}

// ResolverBuilder is fluent builder for Resolver interface
type ResolverBuilder interface {
	// Prefix sets placeholder prefix. When not set, "${" will be used.
	Prefix(string) ResolverBuilder
	// Suffix sets placeholder suffix. When not set, "}" will be used.
	Suffix(string) ResolverBuilder
	// ValueSeparator sets value separator. When not set, ":" will be used.
	ValueSeparator(string) ResolverBuilder
	// LookupFunc sets LookupFn that is used to perform lookups.
	// Call to this function must be made prior to invoking MustBuild function.
	LookupFunc(LookupFn) ResolverBuilder
	// MustBuild builds Resolver from this instance. Panics if mandatory fields has not been set.
	MustBuild() Resolver
}
