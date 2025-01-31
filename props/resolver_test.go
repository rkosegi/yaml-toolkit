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

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	b = Builder().Prefix("${").Suffix("}").ValueSeparator(":")
)

func TestRemoveFromSlice(t *testing.T) {
	assert.Equal(t, 2, len(removeFromSlice([]string{"a", "b"}, "")))
	assert.Equal(t, 1, len(removeFromSlice([]string{"a", "b"}, "a")))
}

func TestIndexAfter(t *testing.T) {
	assert.Equal(t, notFound, indexAfter("Hello", "x", 20))
	assert.Equal(t, notFound, indexAfter("Hello", "o", 20))
	assert.Equal(t, 4, indexAfter("Hello", "o", 3))
}

func TestInvalidBuild(t *testing.T) {
	assert.Panics(t, func() {
		b.MustBuild()
	})
}

func TestResolve(t *testing.T) {
	l := MapLookup(map[string]string{
		"prop1":    "abc",
		"prop2":    "${prop1}:${prop5}@${prop3}",
		"prop5":    "123456",
		"prop3":    "localhost:50000",
		"circ1":    "${circ1}",
		"prop4":    "${propX} ${propY:123}",
		"username": "user1",
		"password": "pwd1",
	})
	r := b.LookupFunc(l).MustBuild()
	assert.Equal(t, "abc", r.Resolve("${prop1}"))
	assert.Equal(t, "abc:123456@localhost:50000", r.Resolve("${prop2}"))
	assert.Panics(t, func() {
		r.Resolve("${circ1}")
	})
	assert.Equal(t, "${propX} 123", r.Resolve("${prop4}"))
	assert.Equal(t, "${propY}", r.Resolve("${propY}"))
	assert.Equal(t, "${prop", r.Resolve("${prop"))
	assert.Equal(t, "username=\"user1\" password=\"pwd1\";",
		r.Resolve("${jaas-config:username=\"${username}\" password=\"${password}\";}"))
}
