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

package analytics

import (
	"testing"

	"github.com/rkosegi/yaml-toolkit/common"
	"github.com/rkosegi/yaml-toolkit/dom"
	"github.com/rkosegi/yaml-toolkit/fluent"
	"github.com/stretchr/testify/assert"
)

func TestDocumentSetAdd(t *testing.T) {
	ds := NewDocumentSet()
	assert.NoError(t, ds.AddPropertiesFromManifest("../testdata/secret3.yaml", WithTags("prop1")))
	assert.Equal(t, 2, len(ds.TaggedSubset("prop1").Merged().Flatten()))
	assert.Equal(t, 0, len(ds.TaggedSubset("unknown").Merged().Flatten()))
	assert.Error(t, ds.AddPropertiesFromManifest("../testdata/non-existent-manifest-file.yaml"))
	assert.NoError(t, ds.AddDocumentFromFile("../testdata/doc2.yaml", dom.DefaultYamlDecoder))
	assert.NoError(t, ds.AddDocumentsFromManifest("../testdata/cm2.yaml", fluent.DefaultFileDecoderProvider))
	assert.Error(t, ds.AddDocumentsFromManifest("non-existent/file.unknown", fluent.DefaultFileDecoderProvider))
	assert.Error(t, ds.AddDocumentFromFile("non-existent/file.unknown", nil))
	assert.Error(t, ds.AddDocumentFromReader("none", common.FailingReader(), dom.DefaultJsonDecoder))
	assert.Error(t, ds.AddDocumentsFromDirectory("[]]", fluent.DefaultFileDecoderProvider))
	assert.NoError(t, ds.AddDocumentsFromDirectory("../testdata/cm2.yaml", fluent.DefaultFileDecoderProvider))
	assert.Error(t, ds.AddDocumentsFromDirectory("../testdata/invalid.yaml", fluent.DefaultFileDecoderProvider))
	assert.Nil(t, ds.NamedDocument("invalid"))
	assert.NotNil(t, ds.NamedDocument("../testdata/cm2.yaml"))
}

func TestDocumentSetMergeTags(t *testing.T) {
	ds := NewDocumentSet()
	assert.NoError(t, ds.AddPropertiesFromManifest("../testdata/secret3.yaml", WithTags("prop1")))
	assert.NoError(t, ds.AddPropertiesFromManifest("../testdata/secret3.yaml", WithTags("prop2"), MergeTags()))
	assert.Equal(t, 3, len(ds.(*documentSet).ctxMap["../testdata/secret3.yaml"].tags))
}

func TestDocumentSetMustCreate(t *testing.T) {
	ds := NewDocumentSet()
	assert.NoError(t, ds.AddPropertiesFromManifest("../testdata/secret3.yaml", WithTags("prop1")))
	assert.Error(t, ds.AddPropertiesFromManifest("../testdata/secret3.yaml", WithTags("prop2"), MustCreate()))
}

func TestDocumentSetOrder(t *testing.T) {
	ds := NewDocumentSet()
	assert.NoError(t, ds.AddDocument("name1", dom.ContainerNode(), WithTags("tag3")))
	assert.NoError(t, ds.AddDocument("name2", dom.ContainerNode(), WithTags("tag1")))
	assert.NoError(t, ds.AddDocument("name3", dom.ContainerNode(), WithTags("tag1")))
	assert.NoError(t, ds.AddDocument("name4", dom.ContainerNode(), WithTags("tag2")))
	assert.NoError(t, ds.AddDocument("name5", dom.ContainerNode(), WithTags("tag2")))

	var (
		od     dom.OverlayDocument
		layers []string
	)
	layers = ds.AsOne().LayerNames()
	assert.Equal(t, "name1", layers[0])

	od = ds.TaggedSubset("tag1")
	layers = od.LayerNames()
	assert.Equal(t, "name2", layers[0])
	assert.Equal(t, "name3", layers[1])

	od = ds.TaggedSubset("tag2")
	layers = od.LayerNames()
	assert.Equal(t, "name4", layers[0])
	assert.Equal(t, "name5", layers[1])

	od = ds.TaggedSubset("tag1", "tag3")
	layers = od.LayerNames()
	assert.Equal(t, "name1", layers[0])
	assert.Equal(t, "name2", layers[1])
	assert.Equal(t, "name3", layers[2])
}
