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

package dom

import "github.com/rkosegi/yaml-toolkit/path"

type walkContainerFn func(ctx *walkCtx, pb path.Builder, parent Container)
type walkListFn func(ctx *walkCtx, pb path.Builder, parent List)

type walkCtx struct {
	cFn walkContainerFn
	lFn walkListFn
	vFn NodeVisitorFn
}

// NodeVisitorFn is function that is called for each child Node within the parent Node.
// Returning false from this function will terminate iteration.
type NodeVisitorFn func(p path.Path, parent Node, node Node) bool

// WalkOpt is function that allows customization of walk process
type WalkOpt func(ctx *walkCtx)

// WalkOptDFS will use https://en.wikipedia.org/wiki/Depth-first_search to traverse Container.
func WalkOptDFS() WalkOpt {
	return func(ctx *walkCtx) {
		ctx.cFn = walkContainerDfs
		ctx.lFn = walkListDfs
	}
}

// WalkOptBFS will use https://en.wikipedia.org/wiki/Breadth-first_search to traverse Container.
func WalkOptBFS() WalkOpt {
	return func(ctx *walkCtx) {
		ctx.cFn = walkContainerBfs
		ctx.lFn = walkListBfs
	}
}

func walkStart(fn NodeVisitorFn, c Container, opts ...WalkOpt) {
	ctx := &walkCtx{vFn: fn}
	// BFS is default
	WalkOptBFS()(ctx)
	for _, opt := range opts {
		opt(ctx)
	}
	ctx.cFn(ctx, path.NewBuilder(), c)
}

func walkListBfs(ctx *walkCtx, pb path.Builder, l List) {
	for idx, item := range l.Items() {
		if !ctx.vFn(pb.Append(path.Numeric(idx)).Build(), l, item) {
			return
		}
		walkDown(ctx, item, pb.Append(path.Numeric(idx)))
	}
}

func walkListDfs(ctx *walkCtx, pb path.Builder, l List) {
	for idx, item := range l.Items() {
		x := pb.Append(path.Numeric(idx))
		walkDown(ctx, item, x)
		if !ctx.vFn(x.Build(), l, item) {
			return
		}
	}
}

func walkDown(ctx *walkCtx, n Node, p path.Builder) {
	if n.IsContainer() {
		ctx.cFn(ctx, p, n.AsContainer())
	} else if n.IsList() {
		ctx.lFn(ctx, p, n.AsList())
	}
}

func walkContainerBfs(ctx *walkCtx, pb path.Builder, c Container) {
	for k, v := range c.Children() {
		x := pb.Append(path.Simple(k))
		if !ctx.vFn(x.Build(), c, v) {
			return
		}
		walkDown(ctx, v, x)
	}
}

func walkContainerDfs(ctx *walkCtx, pb path.Builder, c Container) {
	for k, v := range c.Children() {
		x := pb.Append(path.Simple(k))
		walkDown(ctx, v, x)
		if !ctx.vFn(x.Build(), c, v) {
			return
		}
	}
}
