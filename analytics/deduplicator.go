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
	"github.com/rkosegi/yaml-toolkit/path"
)

// Current implementation might not be correct in all scenarios.
// It works by flattening container to map of paths to leaves.
// Alternative implementation could use different technique like calling Node's Equals() on each visited Node.
// If it returns true, whole subtree can be considered "same" in both containers and walk branch is halted.
// When false, then walk process continues by descending into sub-elements.
//
// Some notes about (un)expected results.
// For some of documents types, users (wrongly) expect equality where they shouldn't.
// Good example are k8s manifests, specifically Volumes and VolumeMounts within PodSpec.
// Those constructs use lists, but yet, order of these items is rather irrelevant at runtime.
// However, if you shuffle items within the list, original list and shuffled one will NOT be equal.

var (
	emptyContainer = dom.ContainerNode().Seal()
	matchAllNodes  = func(path.Path, dom.Node, dom.Node) bool {
		return true
	}
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

type deduplicatorImpl struct {
	filterFn dom.NodeVisitorFn
}

type collector struct {
	sk map[string]path.Path
	pk map[string]dom.Leaf
}

func (c *collector) collect(d dom.Container, matchFn dom.NodeVisitorFn) {
	c.sk = make(map[string]path.Path)
	c.pk = make(map[string]dom.Leaf)
	d.Walk(func(p path.Path, parent dom.Node, node dom.Node) bool {
		if !matchFn(p, parent, node) {
			return false
		}
		return c.visit(p, parent, node)
	})
}

func (c *collector) visit(p path.Path, _ dom.Node, node dom.Node) bool {
	if node.IsLeaf() {
		sp := p.String()
		c.sk[sp] = p
		c.pk[sp] = node.AsLeaf()
	}
	return true
}

func (c *collector) hasNode(sp string, node dom.Node) bool {
	if _, ok := c.pk[sp]; ok && c.pk[sp].Equals(node) {
		return true
	}
	return false
}

func (d *deduplicatorImpl) Deduplicate(od dom.OverlayDocument) (dom.OverlayDocument, dom.Container) {
	var cd = d.FindCommon(od)
	if isContainerEmpty(cd) {
		return od, emptyContainer
	}
	c := collector{}
	c.collect(cd, matchAllNodes)

	out := NewDocumentSet()
	for name, doc := range od.Layers() {
		_ = out.AddDocument(name, fluent.NewMorpher().Copy(doc,
			fluent.CopyModeReplace(fluent.CopyParamFilterFunc(func(p path.Path, parent dom.Node, node dom.Node) bool {
				return node.IsLeaf() && !c.hasNode(p.String(), node)
			}))).Result())
	}
	return out.AsOne(), cd
}

func (d *deduplicatorImpl) FindCommon(od dom.OverlayDocument) dom.Container {
	// no reason to deduplicate zero or one document
	if len(od.LayerNames()) < 2 {
		return emptyContainer
	}

	// cache mapping between string representation of path and actual Path
	km := make(map[string]path.Path)
	oLayers := make(map[string]map[string]dom.Leaf)
	for layerName, layer := range od.Layers() {
		c := &collector{}
		c.collect(layer, d.filterFn)
		for sp, pv := range c.sk {
			km[sp] = pv
		}
		oLayers[layerName] = c.pk
	}

	// get first document
	d1 := oLayers[od.LayerNames()[0]]
	res := dom.ContainerNode()
	for k, v := range d1 {
		if isPropertyInAllDocuments(k, v, oLayers) {
			res.Set(km[k], v)
		}
	}
	return res.Seal()
}

type DeduplicationOpt func(*deduplicatorImpl)

func DeduplicationOptFilterFn(fn dom.NodeVisitorFn) DeduplicationOpt {
	return func(impl *deduplicatorImpl) {
		impl.filterFn = fn
	}
}

func NewDeduplicator(opts ...DeduplicationOpt) Deduplicator {
	di := &deduplicatorImpl{}
	DeduplicationOptFilterFn(matchAllNodes)(di)
	for _, opt := range opts {
		opt(di)
	}
	return di
}
