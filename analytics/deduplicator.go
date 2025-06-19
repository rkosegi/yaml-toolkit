/*
Copyright 2025 Richard Kosegi

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
	"github.com/rkosegi/yaml-toolkit/fluent"
	"github.com/samber/lo"
)

var (
	emptyContainer = dom.Builder().Container().Seal()
)

func isPropertyInAllDocuments(k string, v dom.Leaf, docs map[string]map[string]dom.Leaf) bool {
	found := false
	for _, m := range docs {
		if x, ok := m[k]; ok && x.Value() == v.Value() {
			found = true
		} else {
			return false
		}
	}
	return found
}

func isContainerEmpty(c dom.Container) bool {
	return c == emptyContainer || len(c.Children()) == 0
}

type deduplicatorImpl struct{}

func (d *deduplicatorImpl) Deduplicate(od dom.OverlayDocument) (dom.OverlayDocument, dom.Container) {
	var cd = d.FindCommon(od)
	if isContainerEmpty(cd) {
		return od, emptyContainer
	}

	flatten := cd.Flatten()
	out := NewDocumentSet()
	for name, doc := range od.Layers() {
		_ = out.AddDocument(name, fluent.NewMorpher().AddWithFilter(doc, func(path string, leaf dom.Leaf) bool {
			if x, ok := flatten[path]; ok && x.Value() == leaf.Value() {
				return false
			}
			return true
		}).Result())
	}

	return out.AsOne(), cd
}

func (d *deduplicatorImpl) FindCommon(od dom.OverlayDocument) dom.Container {
	// no reason to deduplicate one document
	if len(od.LayerNames()) < 2 {
		return emptyContainer
	}
	fdocs := lo.MapEntries(od.Layers(), func(key string, value dom.Container) (string, map[string]dom.Leaf) {
		return key, value.Flatten()
	})
	// get first document
	d1 := fdocs[od.LayerNames()[0]]
	res := dom.Builder().Container()
	for k, v := range d1 {
		if isPropertyInAllDocuments(k, v, fdocs) {
			res.AddValueAt(k, v)
		}
	}
	return res.Seal()
}

func NewDeduplicator() Deduplicator {
	return &deduplicatorImpl{}
}
