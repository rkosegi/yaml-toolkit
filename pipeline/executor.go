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
	te "github.com/rkosegi/yaml-toolkit/pipeline/template_engine"
)

var (
	b = dom.Builder()
)

type exec struct {
	*runtimeCtx

	// settable by options
	d dom.ContainerBuilder
	l Listener
}

func (p *exec) Runtime() RuntimeServices {
	return p
}

type actContext struct {
	*exec
	c Action
	*listenerLoggerAdapter
	ssDirty bool
	ss      *map[string]interface{}
}

func (ac *actContext) Action() Action                { return ac.c }
func (ac *actContext) Executor() Executor            { return ac.exec }
func (ac *actContext) Logger() Logger                { return ac }
func (ac *actContext) Data() dom.ContainerBuilder    { return ac.d }
func (ac *actContext) Factory() dom.ContainerFactory { return b }
func (ac *actContext) Snapshot() map[string]interface{} {
	if ac.ssDirty || ac.ss == nil {
		ac.ss = ptr(dom.DefaultNodeEncoderFn(ac.Data()).(map[string]interface{}))
		ac.ssDirty = false
	}
	return *ac.ss
}

func (ac *actContext) Log(v ...interface{}) {
	ac.listenerLoggerAdapter.Log(v...)
}

func (ac *actContext) InvalidateSnapshot() {
	ac.ssDirty = true
}

func (p *exec) newCtx(a Action) *actContext {
	ctx := &actContext{
		c:    a,
		exec: p,
	}
	ctx.listenerLoggerAdapter = &listenerLoggerAdapter{
		c: ctx,
		l: p.l,
	}
	return ctx
}

func (p *exec) NewActionContext(a Action) ActionContext {
	return p.newCtx(a)
}

func (p *exec) Execute(act Action) (err error) {
	ctx := p.newCtx(act)
	p.l.OnBefore(ctx)
	err = act.Do(ctx)
	p.l.OnAfter(ctx, err)
	return err
}

func initService(spec ConfigurableSpec, impl Service) error {
	return impl.Configure(spec.Args).Init()
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
		if err := initService(spec, impl); err != nil {
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

type listenerLoggerAdapter struct {
	c ActionContext
	l Listener
}

func (n *listenerLoggerAdapter) Log(v ...interface{}) {
	n.l.OnLog(n.c, v...)
}

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
