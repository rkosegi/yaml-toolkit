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
	"github.com/rkosegi/yaml-toolkit/dom"
	"slices"
	"strings"
)

func matchAll(string) bool {
	return true
}

func hasPlaceholderFunc(ph string) dom.SearchValueFunc {
	return func(val interface{}) bool {
		if x, ok := val.(string); ok {
			return strings.Contains(x, fmt.Sprintf("${%s}", ph)) ||
				(strings.HasPrefix(x, fmt.Sprintf("${%s:", ph)) && strings.HasSuffix(x, "}"))
		}
		return false
	}
}

type dependencyResolverBuilder struct {
	// predicate to filter out keys to search
	keyFilterFn StringPredicateFn
	// callback function that is invoked when placeholder is found in property
	onPlaceholderEncounteredFn func(string, dom.Coordinates)
	// factory method to provide dom.SearchValueFunc which searches property value for presence of placeholder reference.
	placeholderMatcherFn func(string) dom.SearchValueFunc
}

func (drb *dependencyResolverBuilder) OnPlaceholderEncountered(fn func(string, dom.Coordinates)) DependencyResolverBuilder {
	drb.onPlaceholderEncounteredFn = fn
	return drb
}

func (drb *dependencyResolverBuilder) PlaceholderMatcher(matcherFunc func(string) dom.SearchValueFunc) DependencyResolverBuilder {
	drb.placeholderMatcherFn = matcherFunc
	return drb
}

func (drb *dependencyResolverBuilder) Build() DependencyResolver {
	return &dependencyResolver{
		onPlaceholderEncounteredFn: drb.onPlaceholderEncounteredFn,
		placeholderMatcherFn:       drb.placeholderMatcherFn,
		keyFilterFn:                drb.keyFilterFn,
	}
}

type dependencyResolver struct {
	// placeholderMatcherFn       SearchPlaceholderFunc
	onPlaceholderEncounteredFn func(string, dom.Coordinates)
	placeholderMatcherFn       func(string) dom.SearchValueFunc
	keyFilterFn                StringPredicateFn
}

func unique(in []string) []string {
	ret := make([]string, 0)
	for _, s := range in {
		if !slices.Contains(ret, s) {
			ret = append(ret, s)
		}
	}
	return ret
}

func subtract(from []string, what []string) []string {
	ret := make([]string, 0)
	for _, i := range from {
		if !slices.Contains(what, i) {
			ret = append(ret, i)
		}
	}
	return ret
}

func (dr *dependencyResolver) Resolve(srcDoc dom.OverlayDocument,
	refDocs ...dom.OverlayDocument) *DependencyResolutionReport {
	var (
		allKeys []string
		used    []string
	)

	m := make(map[string]dom.Coordinates)
	for k := range srcDoc.Merged().Flatten() {
		if dr.keyFilterFn(k) {
			allKeys = append(allKeys, k)
			for _, d := range append([]dom.OverlayDocument{srcDoc}, refDocs...) {
				if x := d.Search(dr.placeholderMatcherFn(k)); x != nil {
					used = append(used, k)
					dr.onPlaceholderEncounteredFn(k, x)
					m[k] = append(m[k], x...)
				}
			}
		}
	}
	used = unique(used)
	orphan := subtract(allKeys, used)
	slices.Sort(orphan)
	return &DependencyResolutionReport{
		Map:        m,
		AllKeys:    allKeys,
		OrphanKeys: orphan,
	}
}

func NewDependencyResolverBuilder() DependencyResolverBuilder {
	return &dependencyResolverBuilder{
		onPlaceholderEncounteredFn: func(string, dom.Coordinates) {},
		placeholderMatcherFn:       hasPlaceholderFunc,
		keyFilterFn:                matchAll,
	}
}

// DefaultDependencyResolver returns dependency resolver with default settings
func DefaultDependencyResolver() DependencyResolver {
	return NewDependencyResolverBuilder().Build()
}
