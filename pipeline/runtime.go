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

package pipeline

import (
	"fmt"
	"iter"
	"slices"

	"github.com/rkosegi/yaml-toolkit/dom"
	te "github.com/rkosegi/yaml-toolkit/pipeline/template_engine"
	"github.com/samber/lo"
)

type dataCtx struct {
	d     dom.ContainerBuilder
	dirty bool
	ss    *map[string]interface{}
}

func (dc *dataCtx) Data() dom.ContainerBuilder    { return dc.d }
func (dc *dataCtx) Factory() dom.ContainerFactory { return b }
func (dc *dataCtx) Snapshot() map[string]interface{} {
	if dc.dirty || dc.ss == nil {
		dc.ss = ptr(dom.DefaultNodeEncoderFn(dc.Data()).(map[string]interface{}))
		dc.dirty = false
	}
	return *dc.ss
}
func (dc *dataCtx) InvalidateSnapshot() {
	dc.dirty = true
}

type runtimeCtx struct {
	teng te.TemplateEngine
	sys  dom.ContainerBuilder
}

func (r *runtimeCtx) TemplateEngine() te.TemplateEngine {
	return r.teng
}

func (r *runtimeCtx) Ext() ExtInterface {
	return r
}

func (r *runtimeCtx) DefineAction(name string, spec ActionSpec) {
	// TODO restrict name to something sane, eg no "." or "/"
	// TODO is re-define (defining same name again) allowed?
	r.sys.AddValueAt(fmt.Sprintf("registry.action.define.%s", name), dom.LeafNode(spec))
}

func (r *runtimeCtx) GetAction(name string) (ActionSpec, bool) {
	if ref := r.getRef("registry.action.define", name); ref != nil {
		return ref.(ActionSpec), true
	}
	return ActionSpec{}, false
}

func (r *runtimeCtx) RegisterActionFactory(name string, factory ActionFactory) {
	r.sys.AddValueAt(fmt.Sprintf("registry.action.factory.%s", name), dom.LeafNode(factory))
}

func (r *runtimeCtx) GetActionFactory(name string) ActionFactory {
	if ref := r.getRef("registry.action.factory", name); ref != nil {
		return ref.(ActionFactory)
	}
	return nil
}

func (r *runtimeCtx) RegisterService(name string, impl Service) {
	r.sys.AddValueAt(fmt.Sprintf("registry.service.impl.%s", name), dom.LeafNode(impl))
}

func (r *runtimeCtx) GetService(name string) Service {
	if ref := r.getRef("registry.service.impl", name); ref != nil {
		return ref.(Service)
	}
	return nil
}

func (r *runtimeCtx) EnumServices() iter.Seq[Service] {
	if c := r.sys.Lookup("registry.service.impl"); c != nil {
		return slices.Values(
			lo.Map(
				lo.Values(c.AsContainer().Children()), func(item dom.Node, _ int) Service {
					return item.AsLeaf().Value().(Service)
				},
			),
		)
	}
	return slices.Values([]Service{})
}

func (r *runtimeCtx) getRef(prefix string, name string) any {
	if n := r.sys.Lookup(fmt.Sprintf("%s.%s", prefix, name)); n != nil {
		if n.IsLeaf() && n.AsLeaf().Value() != nil {
			return n.AsLeaf().Value()
		}
	}
	return nil
}

func newRuntimeCtx() *runtimeCtx {
	rt := &runtimeCtx{
		sys: b.Container(),
	}
	return rt
}
