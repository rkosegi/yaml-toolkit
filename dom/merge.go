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

package dom

import "math"

func defaultListMerger() MergeOption {
	return func(m *merger) {
		m.listMergeFn = m.mergeListsMeld
	}
}

var defOpts = []MergeOption{defaultListMerger()}

func ListsMergeAppend() MergeOption {
	return ListsMergeFunc(mergeListsAppend)
}

func ListsMergeFunc(fn func(_, _ List) List) MergeOption {
	return func(m *merger) {
		m.listMergeFn = fn
	}
}

type merger struct {
	listMergeFn func(_, _ List) List
}

func mergeListsAppend(l1, l2 List) List {
	l := &listBuilderImpl{}
	for i := 0; i < l1.Size(); i++ {
		l.Append(l1.Items()[i])
	}
	for i := 0; i < l2.Size(); i++ {
		l.Append(l2.Items()[i])
	}
	return l
}

func (mg *merger) mergeListsMeld(l1, l2 List) List {
	c1 := l1.Size()
	c2 := l2.Size()
	maxLen := int(math.Max(float64(c1), float64(c2)))
	minLen := int(math.Min(float64(c1), float64(c2)))
	l := &listBuilderImpl{}
	for i := 0; i < maxLen; i++ {
		l.Append(nilLeaf)
	}
	for i := 0; i < minLen; i++ {
		n1 := l1.Items()[i]
		n2 := l2.Items()[i]
		if n1.IsContainer() && n2.IsContainer() {
			l.Set(uint(i), mg.mergeContainers(n1.(Container), n2.(Container)))
		} else if n1.IsList() && n2.IsList() {
			l.Set(uint(i), mg.listMergeFn(n1.(List), n2.(List)))
		} else {
			l.Set(uint(i), coalesce(n1, n2))
		}
	}
	for i := minLen; i < maxLen; i++ {
		l.Set(uint(i), firstValidListItem(i, l1, l2))
	}
	return l
}

func (mg *merger) mergeContainers(c1, c2 Container) ContainerBuilder {
	merged := map[string]Node{}
	for k, v := range c1.Children() {
		merged[k] = v
	}
	for k, v := range c2.Children() {
		if n, exists := merged[k]; exists {
			if n.IsContainer() && v.IsContainer() {
				merged[k] = mg.mergeContainers(n.(Container), v.(Container))
			} else if n.IsList() && v.IsList() {
				merged[k] = mg.listMergeFn(n.(List), v.(List))
			} else {
				merged[k] = coalesce(n, v)
			}
		} else {
			merged[k] = v
		}
	}
	r := &containerBuilderImpl{}
	r.children = merged
	return r
}

func (mg *merger) init(opts ...MergeOption) {
	for _, opt := range defOpts {
		opt(mg)
	}
	for _, opt := range opts {
		opt(mg)
	}
}

func (mg *merger) mergeLists(l1, l2 List) List {
	return mg.listMergeFn(l1, l2)
}

func (mg *merger) mergeOverlay(m *overlayDocument) Container {
	var merged Container
	merged = &containerBuilderImpl{}
	for _, name := range m.names {
		merged = mg.mergeContainers(merged, m.overlays[name])
	}
	return merged
}
