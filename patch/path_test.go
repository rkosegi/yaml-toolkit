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

package patch

import (
	"testing"

	"github.com/rkosegi/yaml-toolkit/dom"
	"github.com/stretchr/testify/assert"
)

func TestParsePath(t *testing.T) {
	var (
		p   Path
		err error
	)
	_, err = ParsePath("a")
	assert.Error(t, err)
	p, err = ParsePath("/a/b/c")
	assert.NoError(t, err)
	assert.Equal(t, 3, len(p))
	assert.Equal(t, "a", string(p[0]))
	assert.Equal(t, "b", string(p[1]))
	assert.Equal(t, "c", string(p[2]))
	p, err = ParsePath("")
	assert.NoError(t, err)
	assert.Equal(t, emptyPath, p)
	p, err = ParsePath("/a/b/c~1")
	assert.NoError(t, err)
	assert.Equal(t, 3, len(p))
	p, err = ParsePath("/a/xyz~0123/~")
	assert.NoError(t, err)
	assert.Equal(t, 3, len(p))
	assert.Equal(t, "a", string(p[0]))
	assert.Equal(t, "xyz~123", string(p[1]))
	assert.Equal(t, "~", string(p[2]))
	p, err = ParsePath("/a/~3/mnop")
	assert.NoError(t, err)
	assert.Equal(t, 3, len(p))
	assert.Equal(t, "a", string(p[0]))
	assert.Equal(t, "~3", string(p[1]))
	assert.Equal(t, "mnop", string(p[2]))
}

func TestMustParsePathInvalid(t *testing.T) {
	defer func() {
		recover()
	}()
	MustParsePath("invalid")
	assert.Fail(t, "should not be here")
}
func TestMustParsePath(t *testing.T) {
	p := MustParsePath("")
	assert.Equal(t, emptyPath, p)
}

func TestPathString(t *testing.T) {
	assert.Equal(t, "/a/x~1y/c", Path{"a", "x/y", "c"}.String())
	assert.Equal(t, "/a/x~0y/c", Path{"a", "x~y", "c"}.String())
	assert.Equal(t, "", Path{}.String())
}

func TestEvaluate(t *testing.T) {
	var (
		p   Path
		n   dom.Node
		nl  dom.NodeList
		err error
	)
	c := makeTestContainer()
	assert.NoError(t, err)
	p, _ = ParsePath("/root/list/0/c")
	_, n = p.Eval(c)
	assert.Nil(t, n)
	p, _ = ParsePath("/root/list/0")
	_, n = p.Eval(c)
	assert.NotNil(t, n)
	assert.Equal(t, "item1", n.(dom.Leaf).Value())
	p, _ = ParsePath("/root/sub1")
	_, n = p.Eval(c)
	assert.NotNil(t, n)
	assert.True(t, n.IsContainer())
	p, _ = ParsePath("/root/sub1/prop")
	_, n = p.Eval(c)
	assert.NotNil(t, n)
	assert.Equal(t, 456, n.(dom.Leaf).Value())
	p, _ = ParsePath("/root/list/10")
	_, n = p.Eval(c)
	assert.Nil(t, n)
	p, _ = ParsePath("/root/sub1/prop2")
	_, n = p.Eval(c)
	assert.Nil(t, n)
	nl, _ = p.Eval(c)
	assert.NotNil(t, nl)
	assert.Equal(t, 2, len(nl))
	_, n = p.Parent().Eval(c)
	assert.NotNil(t, n)
	nl, n = emptyPath.Eval(c)
	assert.Equal(t, 1, len(nl))
	assert.Equal(t, c, nl[0])
	assert.Equal(t, c, n)
}

func TestPathParent(t *testing.T) {
	var (
		p Path
	)
	p, _ = ParsePath("/x/y/z")
	p = p.Parent()
	assert.Equal(t, 2, len(p))
	assert.Equal(t, "y", string(p[1]))
	assert.Equal(t, "x", string(p[0]))
	p = p.Parent()
	assert.Equal(t, 1, len(p))
	assert.Equal(t, "x", string(p[0]))
	p = p.Parent()
	assert.Equal(t, 0, len(p))
}

func TestPathLastSegment(t *testing.T) {
	var (
		p Path
	)
	p, _ = ParsePath("/x/y/z")
	assert.Equal(t, "z", string(p.LastSegment()))

	p, _ = ParsePath("")
	assert.Equal(t, "", string(p.LastSegment()))
}
