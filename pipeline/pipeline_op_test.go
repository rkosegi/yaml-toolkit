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
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

type noopService struct {
	Department string
}

func (n *noopService) Configure(ctx ServiceContext, args StrKeysAnyValues) Service {
	ApplyArgs[noopService](ctx, n, args)
	return n
}
func (n *noopService) Init() error  { return nil }
func (n *noopService) Close() error { return nil }

type failingInitService struct {
	*noopService
}

func (n *failingInitService) Configure(ServiceContext, StrKeysAnyValues) Service { return n }
func (n *failingInitService) Init() error                                        { return errors.New("failingInitService") }

func TestPipelineFull(t *testing.T) {
	var (
		pp   PipelineOp
		data []byte
		err  error
		ex   Executor
	)
	data, err = os.ReadFile("../testdata/pipeline2.yaml")
	assert.NoError(t, err)
	err = yaml.Unmarshal(data, &pp)
	assert.NoError(t, err)
	t.Log(pp.String())

	// all good
	ex = New(WithServices(map[string]Service{
		"employee": &noopService{},
	}))
	err = ex.Run(&pp)
	assert.NoError(t, err)

	// fail, non-existent service implementation
	ex = New(WithServices(map[string]Service{
		"something else": &noopService{},
	}))
	err = ex.Run(&pp)
	assert.Error(t, err)

	// fail, service init returned error
	ex = New(WithServices(map[string]Service{
		"employee": &failingInitService{},
	}))
	err = ex.Run(&pp)
	assert.Error(t, err)

	// ok, no service configured
	err = New().Run(&PipelineOp{})
	assert.NoError(t, err)
}
