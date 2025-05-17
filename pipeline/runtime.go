package pipeline

import (
	"fmt"
	"iter"
	"slices"

	"github.com/rkosegi/yaml-toolkit/dom"
	te "github.com/rkosegi/yaml-toolkit/pipeline/template_engine"
	"github.com/samber/lo"
)

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
