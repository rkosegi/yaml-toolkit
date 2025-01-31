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
	"strings"
	"testing"

	"github.com/rkosegi/yaml-toolkit/dom"
	"github.com/stretchr/testify/assert"
)

func TestDefaultDependencyResolver(t *testing.T) {
	ds := NewDocumentSet()
	loadDocsIntoSet(t, ds)
	DefaultDependencyResolver().Resolve(ds.TaggedSubset("defaults"))

	rpt := DefaultDependencyResolver().
		Resolve(ds.TaggedSubset("env/dev", "defaults"),
			ds.TaggedSubset("source"))
	assert.NotNil(t, rpt)
}

func TestDependencyResolver(t *testing.T) {
	var rpt *DependencyResolutionReport
	ds := NewDocumentSet()
	loadDocsIntoSet(t, ds)
	res := NewDependencyResolverBuilder().
		PlaceholderMatcher(hasPlaceholderFunc).
		OnPlaceholderEncountered(func(key string, coordinates dom.Coordinates) {
			t.Logf("key=%s, coords=%s", key, coordinates.String())
		}).Build()
	rpt = res.Resolve(ds.TaggedSubset("env/dev", "defaults"), ds.TaggedSubset("source"))
	t.Logf("orphaned=%s", strings.Join(rpt.OrphanKeys, ","))
	assert.Equal(t, 0, len(rpt.OrphanKeys))
	rpt = res.Resolve(ds.TaggedSubset("env/invalid", "defaults"), ds.TaggedSubset("source"))
	assert.Equal(t, 2, len(rpt.OrphanKeys))
	assert.Equal(t, "defaults.connection.retryCount", rpt.OrphanKeys[0])
	assert.Equal(t, "defaults.connection.timeout", rpt.OrphanKeys[1])
}
