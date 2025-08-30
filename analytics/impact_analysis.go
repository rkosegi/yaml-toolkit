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
	"github.com/rkosegi/yaml-toolkit/common"
	"github.com/rkosegi/yaml-toolkit/dom"
)

type ImpactAnalysis interface {
	ResolveOverlayDocument(od dom.OverlayDocument, keys []string) map[string]dom.Coordinates
	ResolveDocumentSet(ds DocumentSet, keys []string) map[string]dom.Coordinates
}

type impactAnalysis struct {
	keyFilterFn          common.StringPredicateFn
	placeholderMatcherFn func(string) dom.SearchValueFunc
}

func (i *impactAnalysis) ResolveDocumentSet(ds DocumentSet, keys []string) map[string]dom.Coordinates {
	return i.ResolveOverlayDocument(ds.AsOne(), keys)
}

func (i *impactAnalysis) resolveKey(od dom.OverlayDocument, key string) dom.Coordinates {
	if x := od.Search(i.placeholderMatcherFn(key), ps.Serialize); x != nil {
		return x
	}
	return nil
}

func (i *impactAnalysis) ResolveOverlayDocument(od dom.OverlayDocument, keys []string) map[string]dom.Coordinates {
	res := make(map[string]dom.Coordinates)
	for _, key := range keys {
		r := i.resolveKey(od, key)
		if r != nil {
			res[key] = r
		}
	}
	return res
}

type ImpactAnalysisBuilder interface {
	Build() ImpactAnalysis
	// WithKeyFilter allows to override default key filter predicate
	WithKeyFilter(common.StringPredicateFn) ImpactAnalysisBuilder
}

type impactAnalysisBuilder struct {
	// method to override dom.SearchValueFunc which searches property value for presence of placeholder reference.
	placeholderMatcherFn func(string) dom.SearchValueFunc
	keyFilterFn          common.StringPredicateFn
}

func (iab *impactAnalysisBuilder) WithKeyFilter(keyFilterFn common.StringPredicateFn) ImpactAnalysisBuilder {
	iab.keyFilterFn = keyFilterFn
	return iab
}

// Build builds new instance of ImpactAnalysis using current state of builder.
// It's safe to call any builder method after Build() has been called, it won't affect existing builders,
// only newly created called afterward
func (iab *impactAnalysisBuilder) Build() ImpactAnalysis {
	return &impactAnalysis{
		placeholderMatcherFn: iab.placeholderMatcherFn,
		keyFilterFn:          iab.keyFilterFn,
	}
}

// NewImpactAnalysisBuilder returns new instance of ImpactAnalysisBuilder with default values set
func NewImpactAnalysisBuilder() ImpactAnalysisBuilder {
	return &impactAnalysisBuilder{
		placeholderMatcherFn: hasPlaceholderFunc,
		keyFilterFn:          matchAll,
	}
}
