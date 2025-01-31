/*
Copyright 2023 Richard Kosegi

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

package k8s

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/rkosegi/yaml-toolkit/utils"
	"github.com/stretchr/testify/assert"
)

func getTestFileAsReader(file string) (io.Reader, error) {
	if data, err := os.ReadFile(file); err != nil {
		return nil, err
	} else {
		return strings.NewReader(string(data)), nil
	}
}

func TestCRUD(t *testing.T) {
	m, err := ManifestFromFile("../testdata/secret1.yaml")
	assert.Nil(t, err)
	assert.NotNil(t, m)
	assert.NotNil(t, m.StringData().Get("application.yaml"))
	assert.Equal(t, 1, len(m.BinaryData().List()))
	var buff bytes.Buffer
	_, err = m.WriteTo(&buff)
	assert.Nil(t, err)
	m.BinaryData().Update("key2", []byte("abcd"))
	m.BinaryData().Remove("key1")
	assert.Equal(t, 1, len(m.BinaryData().List()))
	buff.Reset()
	_, err = m.WriteTo(&buff)
	assert.Nil(t, err)
}

func TestCRUD2(t *testing.T) {
	r, err := getTestFileAsReader("../testdata/cm1.yaml")
	assert.Nil(t, err)
	m, err := ManifestFromReader(r)
	assert.Nil(t, err)
	assert.NotNil(t, m)
	assert.NotNil(t, m.StringData().Get("application.yaml"))
	assert.Equal(t, 1, len(m.StringData().List()))
	m.StringData().Update("application.yaml", "")
	assert.Equal(t, "", *m.StringData().Get("application.yaml"))
	m.StringData().Remove("application.yaml")
	assert.Nil(t, m.StringData().Get("application.yaml"))
	assert.NotNil(t, m.BinaryData().Get("file.bin"))
	m.BinaryData().Remove("file.bin")
	assert.Nil(t, m.BinaryData().Get("file.bin"))
	var buff bytes.Buffer
	_, err = m.WriteTo(&buff)
	assert.Nil(t, err)
}

func TestInvalidDoc(t *testing.T) {
	_, err := ManifestFromFile("../testdata/invalid.yaml")
	assert.Error(t, err)
}

func TestUnsupportedKind(t *testing.T) {
	_, err := ManifestFromFile("../testdata/unknown_kind.yaml")
	assert.Error(t, err)
}

func TestNoKind(t *testing.T) {
	_, err := ManifestFromFile("../testdata/no_kind.yaml")
	assert.Error(t, err)
}

func TestInvalidBase64(t *testing.T) {
	_, err := ManifestFromFile("../testdata/invalid_base64.yaml")
	assert.Error(t, err)
}

func TestInvalidWriteTo(t *testing.T) {
	m, err := ManifestFromFile("../testdata/cm1.yaml")
	assert.Nil(t, err)
	assert.NotNil(t, m)
	_, err = m.WriteTo(utils.FailingWriter())
	assert.Error(t, err)
}

type failingReader struct{}

func (fr failingReader) Read([]byte) (n int, err error) {
	return 0, anyErr
}

func TestManifestFromReaderFail(t *testing.T) {
	_, err := ManifestFromReader(&failingReader{})
	assert.Error(t, err)
}
