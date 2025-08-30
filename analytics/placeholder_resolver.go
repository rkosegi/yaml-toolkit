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
	"fmt"
	"slices"
	"strings"

	"github.com/rkosegi/yaml-toolkit/common"
	"github.com/rkosegi/yaml-toolkit/dom"
	"github.com/rkosegi/yaml-toolkit/props"
)

type placeholderResolver struct {
	onResolutionFailureFn      OnResolutionFailureFn
	onPlaceholderEncounteredFn OnPlaceholderEncounteredFn
	keyFilterFn                common.StringPredicateFn
	placeholderMatcherFn       common.StringPredicateFn
}

var (
	pp = props.NewPathParser()
	ps = props.NewPathSerializer()
)

func (pr *placeholderResolver) Resolve(doc dom.OverlayDocument) *PlaceholderResolutionReport {
	c := doc.Merged()
	resolver := props.Builder().LookupFunc(func(key string) *string {
		if v := c.Get(pp.MustParse(key)); v != nil {
			x := fmt.Sprintf("%v", v.(dom.Leaf).Value())
			return &x
		}
		return nil
	}).MustBuild()

	failedKeys := make([]string, 0)
	actualValues := make(map[string]interface{})
	coordsMap := make(map[string]dom.Coordinates)
	for k, v := range c.Flatten(ps.Serialize) {
		if pr.keyFilterFn(k) {
			if ph := fmt.Sprintf("%v", v.Value()); pr.placeholderMatcherFn(ph) {
				pr.onPlaceholderEncounteredFn(k, ph)
				p2 := resolver.Resolve(ph)
				if ph == p2 {
					coords := doc.Search(dom.SearchEqual(ph), ps.Serialize)
					pr.onResolutionFailureFn(k, ph, coords)
					if !slices.Contains(failedKeys, ph) {
						failedKeys = append(failedKeys, k)
						actualValues[k] = v
						coordsMap[k] = coords
					}
				}
			}
		}
	}
	slices.Sort(failedKeys)
	return &PlaceholderResolutionReport{
		FailedKeys:   failedKeys,
		ActualValues: actualValues,
		Coordinates:  coordsMap,
	}
}

func NewPlaceholderResolverBuilder() PlaceholderResolverBuilder {
	return &placeholderResolverBuilder{
		onResolutionFailureFn:      func(string, string, dom.Coordinates) {},
		onPlaceholderEncounteredFn: func(string, string) {},
		keyFilterFn:                matchAll,
		placeholderMatcherFn:       possiblyContainsPlaceholder,
	}
}

type placeholderResolverBuilder struct {
	onResolutionFailureFn      OnResolutionFailureFn
	onPlaceholderEncounteredFn OnPlaceholderEncounteredFn
	keyFilterFn                common.StringPredicateFn
	placeholderMatcherFn       common.StringPredicateFn
}

func (rb *placeholderResolverBuilder) WithPlaceholderMatcher(matcherFn common.StringPredicateFn) PlaceholderResolverBuilder {
	rb.placeholderMatcherFn = matcherFn
	return rb
}

func (rb *placeholderResolverBuilder) WithKeyFilter(filterFn common.StringPredicateFn) PlaceholderResolverBuilder {
	rb.keyFilterFn = filterFn
	return rb
}

func (rb *placeholderResolverBuilder) Build() PlaceholderResolver {
	return &placeholderResolver{
		onResolutionFailureFn:      rb.onResolutionFailureFn,
		onPlaceholderEncounteredFn: rb.onPlaceholderEncounteredFn,
		keyFilterFn:                rb.keyFilterFn,
		placeholderMatcherFn:       rb.placeholderMatcherFn,
	}
}

// OnPlaceholderEncountered sets the callback function for situation where placeholder is encountered during resolution process.
func (rb *placeholderResolverBuilder) OnPlaceholderEncountered(fn OnPlaceholderEncounteredFn) PlaceholderResolverBuilder {
	rb.onPlaceholderEncounteredFn = fn
	return rb
}

func (rb *placeholderResolverBuilder) OnResolutionFailure(fn OnResolutionFailureFn) PlaceholderResolverBuilder {
	rb.onResolutionFailureFn = fn
	return rb
}

// possiblyContainsPlaceholder returns true if given string possibly contain property placeholder.
func possiblyContainsPlaceholder(in string) bool {
	idx := strings.Index(in, "${")
	if idx == -1 {
		return false
	}
	return strings.Index(in[idx:], "}") != -1
}
