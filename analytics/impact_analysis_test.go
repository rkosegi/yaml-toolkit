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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAnalyseImpact(t *testing.T) {
	ds := NewDocumentSet()
	loadDocsIntoSet(t, ds)
	ia := NewImpactAnalysisBuilder().
		WithKeyFilter(matchAll).
		Build()

	res := ia.ResolveDocumentSet(ds, []string{
		"env.port",
		"defaults.connection.retryCount",
		"unknown.invalid.key",
	})

	assert.Equal(t, 2, len(res))
	assert.Equal(t, 1, len(res["env.port"]))
}
