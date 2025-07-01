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
	"strings"
	"testing"

	"github.com/rkosegi/yaml-toolkit/dom"
	f "github.com/rkosegi/yaml-toolkit/fluent"
	p "github.com/rkosegi/yaml-toolkit/path"
	"github.com/stretchr/testify/assert"
)

func TestFindCommon(t *testing.T) {
	dd := NewDeduplicator()
	ds := NewDocumentSet()
	assert.NoError(t, ds.AddDocumentFromReader("values-1", strings.NewReader(`
a:
  b:
    c: 1
  m:
    - name: ab
      val: cd
l: x
`), dom.DefaultYamlDecoder))
	assert.NoError(t, ds.AddDocumentFromReader("values-2", strings.NewReader(`
a:
  b:
    c: 1
  m:
    - name: ab
      val: cd
l: y
`), dom.DefaultYamlDecoder))
	res := dd.FindCommon(ds.AsOne())

	assert.Len(t, res.Children(), 1)
	assert.Len(t, res.Child("a").AsContainer().Child("m").AsList().Items(), 1)
}

func TestFindCommonOnlyOne(t *testing.T) {
	dd := NewDeduplicator()
	ds := NewDocumentSet()
	assert.NoError(t, ds.AddDocumentFromReader("values-1", strings.NewReader(`{}`), dom.DefaultJsonDecoder))
	res := dd.FindCommon(ds.AsOne())
	assert.Len(t, res.Children(), 0)
}

func TestDeduplicate(t *testing.T) {
	dd := NewDeduplicator()
	ds := NewDocumentSet()
	d := dom.Builder().FromMap(map[string]interface{}{
		"url":     "http://prod.myapp.tld",
		"timeout": 15000,
		"mounts": []interface{}{
			map[string]interface{}{
				"name": "temp",
				"path": "/tmp",
			},
		},
	})

	assert.NoError(t, ds.AddDocument("prod", d))
	assert.NoError(t, ds.AddDocument("qa", f.NewMorpher().
		Copy(d, f.CopyModeReplace()).
		Mutate(func(b dom.ContainerBuilder) {
			b.AddValue("url", dom.LeafNode("http://qa.myapp.tld"))
		}).Result()))
	assert.NoError(t, ds.AddDocument("test", f.NewMorpher().
		Copy(d, f.CopyModeReplace()).
		Mutate(func(b dom.ContainerBuilder) {
			b.AddValue("url", dom.LeafNode("http://test.myapp.tld"))
		}).Result()))
	assert.NoError(t, ds.AddDocument("dev", f.NewMorpher().
		Copy(d, f.CopyModeReplace()).
		Mutate(func(b dom.ContainerBuilder) {
			b.AddValue("url", dom.LeafNode("http://dev.myapp.tld"))
		}).Result()))

	out, res := dd.Deduplicate(ds.AsOne())
	assert.Equal(t, 2, len(res.Children()))
	assert.Len(t, out.LayerNames(), 4)
}

func TestDeduplicateWithFilter(t *testing.T) {
	doc1 := `
mounts:
  - name: temp
    path: /tmp
probe:
  enabled: true
url: http://prod.myapp.tld
`
	doc2 := `
mounts:
  - name: temp
    path: /tmp
probe:
  enabled: true
url: http://qa.myapp.tld
`

	ds := NewDocumentSet()
	assert.NoError(t, ds.AddDocumentFromReader("doc1", strings.NewReader(doc1), dom.DefaultYamlDecoder))
	assert.NoError(t, ds.AddDocumentFromReader("doc2", strings.NewReader(doc2), dom.DefaultYamlDecoder))
	dd := NewDeduplicator(DeduplicationOptFilterFn(func(p p.Path, parent dom.Node, node dom.Node) bool {
		return !parent.IsList()
	}))
	out, res := dd.Deduplicate(ds.AsOne())
	assert.NotNil(t, out)
	assert.Len(t, res.Children(), 1)
	assert.Len(t, out.Layers()["doc1"].Children(), 2)
}

func TestDeduplicateK8sValues(t *testing.T) {
	ds := NewDocumentSet()
	assert.NoError(t, ds.AddDocumentFromFile("../testdata/k8s_values1.yaml", dom.DefaultYamlDecoder))
	assert.NoError(t, ds.AddDocumentFromFile("../testdata/k8s_values2.yaml", dom.DefaultYamlDecoder))

	dd := NewDeduplicator(DeduplicationOptFilterFn(func(p p.Path, parent dom.Node, node dom.Node) bool {
		// skip lists
		return !parent.IsList()
	}))
	out, res := dd.Deduplicate(ds.AsOne())
	assert.NotNil(t, out)
	assert.Len(t, res.Children(), 1)
	assert.Len(t, out.Layers(), 2)
	assert.Equal(t, `{ "holdApplicationUntilProxyStarts": true }`, res.Get(p.NewBuilder().
		Append(p.Simple("my-app")).
		Append(p.Simple("podAnnotations")).
		Append(p.Simple("proxy.istio.io/config")).
		Build()).AsLeaf().Value())
	assert.Equal(t, "100m", out.Layers()["../testdata/k8s_values1.yaml"].Get(p.NewBuilder().
		Append(p.Simple("my-app")).
		Append(p.Simple("podAnnotations")).
		Append(p.Simple("sidecar.istio.io/proxyCPU")).
		Build()).AsLeaf().Value())
	assert.Equal(t, "50m", out.Layers()["../testdata/k8s_values2.yaml"].Get(p.NewBuilder().
		Append(p.Simple("my-app")).
		Append(p.Simple("podAnnotations")).
		Append(p.Simple("sidecar.istio.io/proxyCPU")).
		Build()).AsLeaf().Value())
}

func TestDeduplicateEmpty(t *testing.T) {
	out, res := NewDeduplicator().Deduplicate(dom.NewOverlayDocument())
	assert.Empty(t, out.LayerNames())
	assert.Empty(t, res.Children())
}

func TestIsContainerEmpty(t *testing.T) {
	assert.True(t, isContainerEmpty(emptyContainer))
	assert.True(t, isContainerEmpty(dom.Builder().Container()))
	assert.False(t, isContainerEmpty(dom.Builder().Container().AddValue("A", dom.LeafNode(1))))
}
