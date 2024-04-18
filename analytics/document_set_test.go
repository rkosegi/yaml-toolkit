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
	"github.com/rkosegi/yaml-toolkit/dom"
	"github.com/rkosegi/yaml-toolkit/utils"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDefaultFileDecoderProvider(t *testing.T) {
	assert.Nil(t, DefaultFileDecoderProvider("file.unknown"))
	assert.NotNil(t, DefaultFileDecoderProvider("file.yaml"))
	assert.NotNil(t, DefaultFileDecoderProvider("file.yml"))
	assert.NotNil(t, DefaultFileDecoderProvider("file.json"))
	assert.NotNil(t, DefaultFileDecoderProvider("file.properties"))
}

func TestDocumentSetAdd(t *testing.T) {
	ds := NewDocumentSet()
	assert.NoError(t, ds.AddPropertiesFromManifest("../testdata/secret3.yaml", WithTags("prop1")))
	assert.Equal(t, 2, len(ds.TaggedSubset("prop1").Merged().Flatten()))
	assert.Equal(t, 0, len(ds.TaggedSubset("unknown").Merged().Flatten()))
	assert.Error(t, ds.AddPropertiesFromManifest("../testdata/non-existent-manifest-file.yaml"))
	assert.NoError(t, ds.AddDocumentFromFile("../testdata/doc2.yaml", dom.DefaultYamlDecoder))
	assert.NoError(t, ds.AddDocumentsFromManifest("../testdata/cm2.yaml", DefaultFileDecoderProvider))
	assert.Error(t, ds.AddDocumentsFromManifest("non-existent/file.unknown", DefaultFileDecoderProvider))
	assert.Error(t, ds.AddDocumentFromFile("non-existent/file.unknown", nil))
	assert.Error(t, ds.AddDocumentFromReader("none", utils.FailingReader(), dom.DefaultJsonDecoder))
	assert.Error(t, ds.AddDocumentsFromDirectory("[]]", DefaultFileDecoderProvider))
	assert.NoError(t, ds.AddDocumentsFromDirectory("../testdata/cm2.yaml", DefaultFileDecoderProvider))
	assert.Error(t, ds.AddDocumentsFromDirectory("../testdata/invalid.yaml", DefaultFileDecoderProvider))
}
