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

package pipeline

import (
	"fmt"

	"github.com/rkosegi/yaml-toolkit/dom"
	"github.com/rkosegi/yaml-toolkit/fluent"
	te "github.com/rkosegi/yaml-toolkit/pipeline/template_engine"
)

var (
	b = dom.Builder()
)

type exec struct {
	*runtimeCtx
	*dataCtx

	// settable by options
	l Listener
}

func (p *exec) Runtime() RuntimeServices {
	return p
}

type clientCtx struct {
	*exec
	c Action
	l Listener
}

func (ac *clientCtx) Action() Action     { return ac.c }
func (ac *clientCtx) Executor() Executor { return ac.exec }
func (ac *clientCtx) Logger() Logger     { return ac }
func (ac *clientCtx) Log(v ...interface{}) {
	ac.l.OnLog(ac, v...)
}

func ApplyArgs[T any](ctx ClientContext, opSpec *T, args map[string]interface{}) {
	m := ctx.TemplateEngine().RenderMapLenient(args, ctx.Snapshot())
	x := fluent.NewConfigHelper[T]().Add(opSpec).Add(m).Result()
	*opSpec = *x
}

func (p *exec) newActionCtx(a Action) *clientCtx {
	ctx := p.newServiceCtx()
	ctx.c = a
	return ctx
}

func (p *exec) newServiceCtx() *clientCtx {
	return &clientCtx{
		exec: p,
		l:    p.l,
	}
}

func (p *exec) Execute(act Action) (err error) {
	ctx := p.newActionCtx(act)
	p.l.OnBefore(ctx)
	err = act.Do(ctx)
	p.l.OnAfter(ctx, err)
	return err
}

func (p *exec) initServices(servicesConfig map[string]ConfigurableSpec) error {
	for name, spec := range servicesConfig {
		var impl Service
		// every service specification must have runtime-registered counterpart
		// but still you can have registered service without configuration
		if impl = p.GetService(name); impl == nil {
			return fmt.Errorf("service '%s' is not registered", name)
		}
		// apply configuration to instance and call Init() on it
		sctx := p.newServiceCtx()

		if err := impl.Configure(sctx, spec.Args).Init(); err != nil {
			return err
		}
	}
	return nil
}

func (p *exec) closeServices() {
	for s := range p.EnumServices() {
		_ = s.Close()
	}
}

func (p *exec) Run(po *PipelineOp) error {
	if err := p.initServices(po.Services); err != nil {
		return err
	}
	p.d = p.d.Merge(b.FromMap(po.Vars))
	defer p.closeServices()
	return p.Execute(po)
}

type noopListener struct{}

func (n *noopListener) OnBefore(ActionContext)              {}
func (n *noopListener) OnAfter(ActionContext, error)        {}
func (n *noopListener) OnLog(ActionContext, ...interface{}) {}

type Opt func(*exec)

func WithListener(l Listener) Opt {
	return func(p *exec) {
		p.l = l
	}
}

func WithData(gd dom.ContainerBuilder) Opt {
	return func(p *exec) {
		p.d = gd
	}
}

func WithTemplateEngine(t te.TemplateEngine) Opt {
	return func(p *exec) {
		p.teng = t
	}
}

func WithExtActions(m map[string]ActionFactory) Opt {
	return func(p *exec) {
		for k, v := range m {
			p.RegisterActionFactory(k, v)
		}
	}
}

func WithServices(s map[string]Service) Opt {
	return func(p *exec) {
		for k, v := range s {
			p.RegisterService(k, v)
		}
	}
}

var defOpts = []Opt{
	WithListener(&noopListener{}),
	WithData(b.Container()),
	WithTemplateEngine(te.DefaultTemplateEngine()),
	WithExtActions(make(map[string]ActionFactory)),
	WithServices(make(map[string]Service)),
}

func New(opts ...Opt) Executor {
	p := &exec{
		dataCtx:    &dataCtx{},
		runtimeCtx: newRuntimeCtx(),
	}
	for _, opt := range defOpts {
		opt(p)
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}
