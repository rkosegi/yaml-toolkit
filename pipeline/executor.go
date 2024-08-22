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
	sprig "github.com/go-task/slim-sprig/v3"
	"github.com/rkosegi/yaml-toolkit/dom"
)

var (
	b = dom.Builder()
)

type exec struct {
	gd dom.ContainerBuilder
	l  Listener
	t  TemplateEngine
}

type actContext struct {
	c Action
	d dom.ContainerBuilder
	e Executor
	f dom.ContainerFactory
	t TemplateEngine
	l *listenerLoggerAdapter
}

func (ac actContext) Action() Action                 { return ac.c }
func (ac actContext) Data() dom.ContainerBuilder     { return ac.d }
func (ac actContext) Factory() dom.ContainerFactory  { return ac.f }
func (ac actContext) Executor() Executor             { return ac.e }
func (ac actContext) TemplateEngine() TemplateEngine { return ac.t }
func (ac actContext) Logger() Logger                 { return ac.l }
func (ac actContext) Snapshot() map[string]interface{} {
	return dom.DefaultNodeEncoderFn(ac.Data()).(map[string]interface{})
}

func (p *exec) newCtx(a Action) *actContext {
	ctx := &actContext{
		c: a,
		d: p.gd,
		e: p,
		f: b,
		t: p.t,
		l: &listenerLoggerAdapter{l: p.l},
	}
	ctx.l.c = ctx
	return ctx
}

func (p *exec) Execute(act Action) (err error) {
	ctx := p.newCtx(act)
	p.l.OnBefore(ctx)
	err = act.Do(ctx)
	p.l.OnAfter(ctx, err)
	return err
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
	n.l.OnLog(n.c, v)
}

type Opt func(*exec)

func WithListener(l Listener) Opt {
	return func(p *exec) {
		p.l = l
	}
}

func WithData(gd dom.ContainerBuilder) Opt {
	return func(p *exec) {
		p.gd = gd
	}
}

func WithTemplateEngine(t TemplateEngine) Opt {
	return func(p *exec) {
		p.t = t
	}
}

var defOpts = []Opt{
	WithListener(&noopListener{}),
	WithData(b.Container()),
	WithTemplateEngine(&templateEngine{
		fm: sprig.TxtFuncMap(),
	}),
}

func New(opts ...Opt) Executor {
	p := &exec{}
	for _, opt := range defOpts {
		opt(p)
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}
