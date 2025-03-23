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
	"os"
	"testing"

	"github.com/rkosegi/yaml-toolkit/dom"
	"github.com/stretchr/testify/assert"
)

func TestHtml2Dom(t *testing.T) {
	var (
		doc1data []byte
		err      error
	)
	d := b.Container()
	d.AddValue("html1", dom.LeafNode("<root />"))
	doc1data, err = os.ReadFile("../testdata/doc1.html")
	d.AddValue("html2", dom.LeafNode(string(doc1data)))
	assert.NoError(t, err)
	assert.NotNil(t, doc1data)
	ctx := newMockActBuilder().data(d).testLogger(t).build()
	for _, ic := range []*Html2DomOp{
		{},
		{
			From: "from",
		},
		{
			From: "from",
			To:   "to",
		},
		{
			From:   "html1",
			To:     "to",
			Layout: ptr(Html2DomLayout("invalid")),
		},
		{
			From:  "html1",
			To:    "to",
			Query: ptr("////not a valid xpath"),
		},
		{
			From:  "html1",
			To:    "to",
			Query: ptr("//non-existent-node/bla"),
		},
	} {
		assert.Error(t, ctx.Executor().Execute(ic))
	}

	err = ctx.Executor().Execute(&Html2DomOp{
		From:  "html2",
		To:    "Result.Out",
		Query: ptr("//span[@class='panel1']"),
	})
	assert.NoError(t, err)
	assert.Equal(t, "panel1", d.Lookup("Result.Out.span.Attrs.class").(dom.Leaf).Value())
	assert.Equal(t, "Click here", d.Lookup("Result.Out.span.span[2].a.Value").(dom.Leaf).Value())
	assert.Equal(t, "http://localhost:8080/doc1", d.Lookup("Result.Out.span.span[2].a.Attrs.href").(dom.Leaf).Value())
}

func TestHtml2DomCloneWith(t *testing.T) {
	d := b.Container()
	d.AddValueAt("Args.From", dom.LeafNode("from.here"))
	d.AddValueAt("Args.To", dom.LeafNode("to.here"))
	ctx := newMockActBuilder().data(d).testLogger(t).build()
	orig := &Html2DomOp{
		From: "{{ .Args.From }}",
		To:   "{{ .Args.To }}",
	}
	clone := orig.CloneWith(ctx)
	assert.Equal(t, "from.here", clone.(*Html2DomOp).From)
	assert.Equal(t, "to.here", clone.(*Html2DomOp).To)
}
