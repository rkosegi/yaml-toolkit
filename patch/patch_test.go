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
	"github.com/rkosegi/yaml-toolkit/dom"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMissing(t *testing.T) {
	var err error
	err = Do(nil, b.Container())
	assert.Error(t, err)
	err = Do(&OpObj{}, b.Container())
	assert.Error(t, err)
	err = Do(&OpObj{
		Path: emptyPath,
	}, nil)
	assert.Error(t, err)
	err = Do(&OpObj{
		Path: emptyPath,
	}, makeTestContainer())
	assert.Error(t, err)
}

func TestPatchOpAdd(t *testing.T) {
	var (
		c   dom.ContainerBuilder
		err error
	)
	c = makeTestContainer()
	// value missing
	err = Do(&OpObj{
		Op:   OpAdd,
		Path: MustParsePath("/root/list/2"),
	}, c)
	assert.Error(t, err)
	// parent not resolvable
	err = Do(&OpObj{
		Op:    OpAdd,
		Path:  MustParsePath("/root/not-existent/2"),
		Value: dom.LeafNode(1),
	}, c)
	assert.Error(t, err)

	err = Do(&OpObj{
		Op:    OpAdd,
		Path:  MustParsePath("/root/leaf20"),
		Value: dom.LeafNode(1),
	}, c)
	assert.NoError(t, err)

	c = makeTestContainer()
	assert.Equal(t, 4, c.Lookup("root.list").(dom.List).Size())
	err = Do(&OpObj{
		Op:    OpAdd,
		Path:  MustParsePath("/root/list/2"),
		Value: dom.LeafNode(1),
	}, c)
	assert.NoError(t, err)
	assert.Equal(t, 5, c.Lookup("root.list").(dom.List).Size())
	assert.Equal(t, 1, c.Lookup("root.list[2]").(dom.Leaf).Value())
}

func TestPatchOpRemove(t *testing.T) {
	var (
		c   dom.ContainerBuilder
		err error
	)
	c = makeTestContainer()
	err = Do(&OpObj{
		Op:   OpRemove,
		Path: MustParsePath("/not/exists/10"),
	}, c)
	assert.Error(t, err)
	assert.Equal(t, 4, len(c.Lookup("root.list").(dom.List).Items()))
	err = Do(&OpObj{
		Op:   OpRemove,
		Path: MustParsePath("/root/list/0"),
	}, c)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(c.Lookup("root.list").(dom.List).Items()))

	c = makeTestContainer()
	err = Do(&OpObj{
		Op:   OpRemove,
		Path: MustParsePath("/root/list/2"),
	}, c)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(c.Lookup("root.list").(dom.List).Items()))
	assert.Equal(t, "item4", c.Lookup("root.list[2]").(dom.Leaf).Value())

	err = Do(&OpObj{
		Op:   OpRemove,
		Path: MustParsePath("/root/list"),
	}, c)
	assert.NoError(t, err)
	assert.Nil(t, c.Lookup("root.list"))
}

func TestPatchOpReplace(t *testing.T) {
	var (
		c   dom.ContainerBuilder
		err error
	)
	c = makeTestContainer()
	// value missing
	err = Do(&OpObj{
		Op:   OpReplace,
		Path: emptyPath,
	}, c)
	assert.Error(t, err)
	err = Do(&OpObj{
		Op:    OpReplace,
		Path:  MustParsePath("/root/sub1"),
		Value: dom.LeafNode("abc"),
	}, c)
	assert.NoError(t, err)
	assert.Equal(t, "abc", c.Lookup("root.sub1").(dom.Leaf).Value())
	err = Do(&OpObj{
		Op:    OpReplace,
		Path:  MustParsePath("/root/non-existent/path"),
		Value: dom.LeafNode("abc"),
	}, c)
	assert.Error(t, err)
	assert.Equal(t, 4, len(c.Lookup("root.list").(dom.List).Items()))
	err = Do(&OpObj{
		Op:    OpReplace,
		Path:  MustParsePath("/root/list/1"),
		Value: dom.LeafNode("abc"),
	}, c)
	assert.NoError(t, err)
	assert.Equal(t, 4, len(c.Lookup("root.list").(dom.List).Items()))
	assert.Equal(t, "abc", c.Lookup("root.list[1]").(dom.Leaf).Value())
}

func TestPatchOpMove(t *testing.T) {
	var err error
	// missing "from"
	err = Do(&OpObj{
		Op:    OpMove,
		Path:  MustParsePath("/root/sub1/prop2"),
		Value: dom.LeafNode(1),
	}, makeTestContainer())
	assert.Error(t, err)
	f := MustParsePath("/root/sub1/prop")
	c := makeTestContainer()
	err = Do(&OpObj{
		Op:   OpMove,
		Path: MustParsePath("/root/sub1/prop2"),
		From: &f,
	}, c)
	assert.NoError(t, err)
	assert.Nil(t, c.Lookup("root.sub1.prop"))
	assert.Equal(t, 456, c.Lookup("root.sub1.prop2").(dom.Leaf).Value())
}

func TestPatchOpCopy(t *testing.T) {
	var (
		err error
		c   dom.ContainerBuilder
		f   Path
	)
	// missing "from"
	err = Do(&OpObj{
		Op:    OpCopy,
		Path:  MustParsePath("/root/sub1/prop2"),
		Value: dom.LeafNode(1),
	}, makeTestContainer())
	assert.Error(t, err)
	f = MustParsePath("/root/sub1/prop")
	c = makeTestContainer()
	err = Do(&OpObj{
		Op:   OpCopy,
		Path: MustParsePath("/root/sub1/prop2"),
		From: &f,
	}, c)
	assert.NoError(t, err)
	assert.Equal(t, 456, c.Lookup("root.sub1.prop").(dom.Leaf).Value())
	assert.Equal(t, 456, c.Lookup("root.sub1.prop2").(dom.Leaf).Value())
	f = MustParsePath("/root/sub10/prop10")
	c = makeTestContainer()
	err = Do(&OpObj{
		Op:   OpCopy,
		Path: MustParsePath("/root/sub1/prop2"),
		From: &f,
	}, c)
	assert.Error(t, err)
}

func TestPatchOpTest(t *testing.T) {
	var err error
	// missing value
	err = Do(&OpObj{
		Op:    OpTest,
		Path:  emptyPath,
		Value: nil,
	}, makeTestContainer())
	assert.Error(t, err)
	// positive case
	err = Do(&OpObj{
		Op:    OpTest,
		Path:  MustParsePath("/root/list/0"),
		Value: dom.LeafNode("item1"),
	}, makeTestContainer())
	assert.NoError(t, err)
	// value mismatch
	err = Do(&OpObj{
		Op:    OpTest,
		Path:  MustParsePath("/root/list/0"),
		Value: dom.LeafNode("item2"),
	}, makeTestContainer())
	assert.Error(t, err)
	// unresolvable path
	err = Do(&OpObj{
		Op:    OpTest,
		Path:  MustParsePath("/root/list1/0"),
		Value: dom.LeafNode("item2"),
	}, makeTestContainer())
	assert.Error(t, err)
}
