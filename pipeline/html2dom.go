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
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/antchfx/htmlquery"
	"github.com/rkosegi/yaml-toolkit/dom"
	"golang.org/x/net/html"
)

type (
	Html2DomLayout string
	LayoutFn       func(dom.ContainerBuilder, *html.Node)
)

var (
	cf          = dom.Builder()
	layoutFnMap = map[Html2DomLayout]LayoutFn{
		Html2DomLayoutDefault: convertHtmlNode2Dom,
	}
)

const (
	AttributeNode = "Attrs"
	ValueNode     = "Value"

	// Html2DomLayoutDefault will produce "Value" leaf for every text node.
	// Child elements are collected into the list, if their name appears multiple times within the parent,
	// otherwise they are regular child node.
	// Attributes of element are put into container node "Attrs".
	// Namespaces are ignored.
	Html2DomLayoutDefault = Html2DomLayout("default")
)

type Html2DomOp struct {
	// From is path within the global data to the leaf node where XML source is stored as string.
	From string `yaml:"from" clone:"template"`
	// Query is optional xpath expression to use to extract subset from source XML document.
	// When omitted, then whole document is used.
	Query *string `yaml:"query"`
	// To is destination where to put converted document as dom.Container.
	To string `yaml:"to" clone:"template"`
	// Layout defines how HTML data are put into DOM
	Layout *Html2DomLayout `yaml:"layout"`
}

func (x *Html2DomOp) String() string {
	return fmt.Sprintf("Html2DomOp[from=%s,to=%s]", x.From, x.To)
}

func (x *Html2DomOp) Do(ctx ActionContext) error {
	ss := ctx.Snapshot()
	from := ctx.TemplateEngine().RenderLenient(x.From, ss)
	to := ctx.TemplateEngine().RenderLenient(x.To, ss)
	if len(from) == 0 {
		return errors.New("'from' is empty")
	}
	if len(to) == 0 {
		return errors.New("'to' is empty")
	}
	fromNode := ctx.Data().Lookup(from)
	if fromNode == nil || !fromNode.IsLeaf() {
		return fmt.Errorf("cannot find leaf node at '%s'", from)
	}
	htmlData := fromNode.(dom.Leaf).Value().(string)
	var (
		err      error
		buff     bytes.Buffer
		srcNode  *html.Node
		layout   Html2DomLayout
		layoutFn LayoutFn
		ok       bool
	)
	layout = Html2DomLayoutDefault
	if x.Layout != nil {
		layout = *x.Layout
	}
	if layoutFn, ok = layoutFnMap[layout]; !ok {
		return fmt.Errorf("unknown layout %s", layout)
	}
	_, _ = buff.WriteString(htmlData)
	// TODO: how can parse return an error?
	srcNode, _ = htmlquery.Parse(&buff)
	if x.Query != nil {
		srcNode, err = htmlquery.Query(srcNode, *x.Query)
	}
	if err != nil {
		return err
	}
	if srcNode == nil {
		return fmt.Errorf("cannot find node at %s", from)
	}
	cb := ctx.Factory().Container()
	layoutFn(cb, srcNode)
	ctx.Data().AddValueAt(to, cb)
	return nil
}

func convertHtmlNode2Dom(cb dom.ContainerBuilder, node *html.Node) {
	switch node.Type {
	case html.ElementNode:
		c := cf.Container()
		if existing := cb.Child(node.Data); existing != nil {
			if existing.IsList() {
				existing.(dom.ListBuilder).Append(c)
			} else {
				l := dom.ListNode(existing, c)
				cb.AddValue(node.Data, l)
			}
		} else {
			cb.AddValue(node.Data, c)
		}
		if len(node.Attr) > 0 {
			ac := c.AddContainer(AttributeNode)
			for _, attr := range node.Attr {
				ac.AddValue(attr.Key, dom.LeafNode(attr.Val))
			}
		}
		for child := range node.ChildNodes() {
			convertHtmlNode2Dom(c, child)
		}

	case html.TextNode:
		if val := strings.TrimSpace(node.Data); val != "" {
			cb.AddValue(ValueNode, dom.LeafNode(node.Data))
		}
	}
}

func (x *Html2DomOp) CloneWith(ctx ActionContext) Action {
	ss := ctx.Snapshot()
	return &Html2DomOp{
		From:   ctx.TemplateEngine().RenderLenient(x.From, ss),
		Query:  x.Query,
		To:     ctx.TemplateEngine().RenderLenient(x.To, ss),
		Layout: x.Layout,
	}
}
