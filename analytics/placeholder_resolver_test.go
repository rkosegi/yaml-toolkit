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

package analytics

import (
	"github.com/rkosegi/yaml-toolkit/dom"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIsPossiblePlaceholder(t *testing.T) {
	assert.True(t, possiblyContainsPlaceholder("${abc}"))
	assert.False(t, possiblyContainsPlaceholder("}${"))
	assert.True(t, possiblyContainsPlaceholder("${a.b.c}"))
	assert.False(t, possiblyContainsPlaceholder(""))
	assert.False(t, possiblyContainsPlaceholder("abcd"))
}

func TestResolvePlaceholders(t *testing.T) {
	ds := NewDocumentSet()
	ds.AddUnnamedDocument(b.FromMap(map[string]interface{}{
		"key1.key2.key31": "${key1.key2.key32}",
		"key1.key2.key32": 3,
		"key1.key2.key33": "${key1.key2.key34}",
		"key1.key2.key40": "${key1.key2.key41}",
		"key1.key2.key41": "${key1.key2.key42}",
		"key1.key2.key42": "${unresolved}",
		"key1.key2.key43": "${unresolved}",
	}))
	res := NewPlaceholderResolverBuilder().
		OnResolutionFailure(func(key, value string, coordinates dom.Coordinates) {
			t.Logf("resolution failed,key=%s, value=%s, coordinates=%s", key, value, coordinates.String())
		}).
		WithPlaceholderMatcher(possiblyContainsPlaceholder).
		OnPlaceholderEncountered(func(key, ph string) {
			t.Logf("key=%s, placeholder=%s", key, ph)
		}).
		WithKeyFilter(func(key string) bool {
			return key != "skip.me"
		}).
		Build()
	rpt := res.Resolve(ds.AsOne())

	assert.Equal(t, 3, len(rpt.FailedKeys))
	res = NewPlaceholderResolverBuilder().
		Build()
	rpt = res.Resolve(ds.AsOne())
	assert.Equal(t, 3, len(rpt.FailedKeys))
}
