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

package fluent

import (
	"os"
	"path"
	"strings"
	"testing"

	"github.com/rkosegi/yaml-toolkit/dom"
	"github.com/stretchr/testify/assert"
)

type config struct {
	Path    *string
	Host    string
	Port    int
	Timeout *int
}

func ptr[T any](v T) *T {
	return &v
}

func TestConfigHelper(t *testing.T) {

	defCfg := &config{
		Path: ptr("/tmp/x"),
	}
	tmpDir, err := os.MkdirTemp("", "yt*")
	assert.NoError(t, err)
	t.Cleanup(func() {
		_ = os.RemoveAll(tmpDir)
	})

	t.Run("load normal config", func(t *testing.T) {
		cfg := NewConfigHelper[config]().
			Add(nil).
			Add(defCfg).
			Add(dom.ContainerNode()).
			Load("../testdata/cfg1.yaml").
			Read(strings.NewReader(`port: 9988`)).
			Mutate(func(cb dom.ContainerBuilder) {
				cb.AddValue("host", dom.LeafNode("localhost"))
			}).
			Save(path.Join(tmpDir, "config.yaml")).
			Result()
		assert.Equal(t, "localhost", cfg.Host)
		assert.Equal(t, 9988, cfg.Port)
		assert.Equal(t, 673, *cfg.Timeout)
	})

	t.Run("load empty YAML file", func(t *testing.T) {
		cfg := NewConfigHelper[config]().
			Add(defCfg).
			Load("../testdata/empty_doc.yaml").
			Save(path.Join(tmpDir, "config.yaml")).
			Result()
		assert.Equal(t, "/tmp/x", *cfg.Path)
	})

}

func TestConfigHelperLoadInvalid(t *testing.T) {
	defer func() {
		recover()
	}()
	NewConfigHelper[config]().Load("/tmp/this/should/not/exists")
	t.Fatal("should panic")
}

func TestDefaultFileEncoderProvider(t *testing.T) {
	for _, ext := range []string{"a.yaml", "b.yml", "c.json", "d.properties"} {
		t.Log("file:", ext)
		assert.NotNil(t, DefaultFileEncoderProvider(ext))
	}
	assert.Nil(t, DefaultFileEncoderProvider(".unknown"))
}

func TestDefaultFileDecoderProvider(t *testing.T) {
	for _, ext := range []string{"a.yaml", "b.yml", "c.json", "d.properties"} {
		t.Log("file:", ext)
		assert.NotNil(t, DefaultFileDecoderProvider(ext))
	}
	assert.Nil(t, DefaultFileDecoderProvider(".unknown"))
}
