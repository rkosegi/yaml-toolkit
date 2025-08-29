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

func TestMissing(t *testing.T) {
	var err error
	err = Do(nil, dom.ContainerNode())
	assert.Error(t, err)
	err = Do(&OpObj{}, dom.ContainerNode())
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
	assert.Equal(t, 4, c.Child("root").AsContainer().Child("list").AsList().Size())
	err = Do(&OpObj{
		Op:    OpAdd,
		Path:  MustParsePath("/root/list/2"),
		Value: dom.LeafNode(1),
	}, c)
	assert.NoError(t, err)
	assert.Equal(t, 5, c.Child("root").
		AsContainer().Child("list").AsList().Size())
	assert.Equal(t, 1, c.Child("root").
		AsContainer().Child("list").AsList().Get(2).AsLeaf().Value())
}

func TestPatchOpRemove(t *testing.T) {
	var (
		c   dom.ContainerBuilder
		err error
	)
	t.Run("non-existent path", func(t *testing.T) {
		c = makeTestContainer()
		err = Do(&OpObj{
			Op:   OpRemove,
			Path: MustParsePath("/not/exists/10"),
		}, c)
		assert.Error(t, err)
		assert.Equal(t, 4, len(c.Child("root").
			AsContainer().Child("list").AsList().Items()))
	})
	t.Run("valid case - remove list item at 0 index", func(t *testing.T) {
		c = makeTestContainer()
		err = Do(&OpObj{
			Op:   OpRemove,
			Path: MustParsePath("/root/list/0"),
		}, c)
		assert.NoError(t, err)
		assert.Len(t, c.Child("root").AsContainer().Child("list").AsList().Items(), 3)
	})

	t.Run("valid case - remove list item in the middle", func(t *testing.T) {
		c = makeTestContainer()
		err = Do(&OpObj{
			Op:   OpRemove,
			Path: MustParsePath("/root/list/2"),
		}, c)
		assert.NoError(t, err)
		assert.Len(t, c.Child("root").AsContainer().Child("list").AsList().Items(), 3)
		assert.Equal(t, "item4", c.Child("root").
			AsContainer().Child("list").AsList().Get(2).(dom.Leaf).Value())

	})

	t.Run("valid case - remove whole list", func(t *testing.T) {
		err = Do(&OpObj{
			Op:   OpRemove,
			Path: MustParsePath("/root/list"),
		}, c)
		assert.NoError(t, err)
		assert.Nil(t, c.Child("root").AsContainer().Child("list"))
	})
}

func TestPatchOpReplace(t *testing.T) {
	var (
		c   dom.ContainerBuilder
		err error
	)
	c = makeTestContainer()
	t.Run("missing value", func(t *testing.T) {
		assert.Error(t, Do(&OpObj{
			Op:   OpReplace,
			Path: emptyPath,
		}, c))
	})
	t.Run("valid case - leaf", func(t *testing.T) {
		err = Do(&OpObj{
			Op:    OpReplace,
			Path:  MustParsePath("/root/sub1"),
			Value: dom.LeafNode("abc"),
		}, c)
		assert.NoError(t, err)
		assert.Equal(t, "abc", c.Child("root").
			AsContainer().Child("sub1").AsLeaf().Value())
	})
	t.Run("non-existent path", func(t *testing.T) {
		c = makeTestContainer()
		err = Do(&OpObj{
			Op:    OpReplace,
			Path:  MustParsePath("/root/non-existent/path"),
			Value: dom.LeafNode("abc"),
		}, c)
		assert.Error(t, err)
		assert.Equal(t, 4, len(c.Child("root").AsContainer().Child("list").AsList().Items()))
	})
	t.Run("valid case - list item", func(t *testing.T) {
		err = Do(&OpObj{
			Op:    OpReplace,
			Path:  MustParsePath("/root/list/1"),
			Value: dom.LeafNode("abc"),
		}, c)
		assert.NoError(t, err)
		assert.Equal(t, 4, len(c.Child("root").AsContainer().Child("list").AsList().Items()))
		assert.Equal(t, "abc", c.Child("root").AsContainer().Child("list").
			AsList().Get(1).AsLeaf().Value())
	})
}

func TestPatchOpMove(t *testing.T) {
	var err error
	t.Run("missing from", func(t *testing.T) {
		assert.Error(t, Do(&OpObj{
			Op:    OpMove,
			Path:  MustParsePath("/root/sub1/prop2"),
			Value: dom.LeafNode(1),
		}, makeTestContainer()))
	})

	t.Run("valid case", func(t *testing.T) {
		f := MustParsePath("/root/sub1/prop")
		c := makeTestContainer()
		err = Do(&OpObj{
			Op:   OpMove,
			Path: MustParsePath("/root/sub1/prop2"),
			From: &f,
		}, c)
		assert.NoError(t, err)
		assert.Nil(t, c.Child("root").
			AsContainer().Child("sub1").
			AsContainer().Child("prop"))
		assert.Equal(t, 456, c.Child("root").
			AsContainer().Child("sub1").
			AsContainer().Child("prop2").AsLeaf().Value())
	})
}

func TestPatchOpCopy(t *testing.T) {
	var (
		err error
		c   dom.ContainerBuilder
		f   Path
	)
	t.Run("missing from", func(t *testing.T) {
		err = Do(&OpObj{
			Op:    OpCopy,
			Path:  MustParsePath("/root/sub1/prop2"),
			Value: dom.LeafNode(1),
		}, makeTestContainer())
		assert.Error(t, err)
	})
	t.Run("valid case", func(t *testing.T) {
		f = MustParsePath("/root/sub1/prop")
		c = makeTestContainer()
		err = Do(&OpObj{
			Op:   OpCopy,
			Path: MustParsePath("/root/sub1/prop2"),
			From: &f,
		}, c)
		assert.NoError(t, err)
		assert.Equal(t, 456, c.Child("root").
			AsContainer().Child("sub1").
			AsContainer().Child("prop").AsLeaf().Value())
		assert.Equal(t, 456, c.Child("root").
			AsContainer().Child("sub1").
			AsContainer().Child("prop2").AsLeaf().Value())
	})
	t.Run("non-existent path", func(t *testing.T) {
		f = MustParsePath("/root/sub10/prop10")
		c = makeTestContainer()
		err = Do(&OpObj{
			Op:   OpCopy,
			Path: MustParsePath("/root/sub1/prop2"),
			From: &f,
		}, c)
		assert.Error(t, err)
	})
}

func TestPatchOpTest(t *testing.T) {
	t.Run("missing value", func(t *testing.T) {
		assert.Error(t, Do(&OpObj{
			Op:    OpTest,
			Path:  emptyPath,
			Value: nil,
		}, makeTestContainer()))
	})
	t.Run("positive case", func(t *testing.T) {
		assert.NoError(t, Do(&OpObj{
			Op:    OpTest,
			Path:  MustParsePath("/root/list/0"),
			Value: dom.LeafNode("item1"),
		}, makeTestContainer()))
	})
	t.Run("value mismatch", func(t *testing.T) {
		assert.Error(t, Do(&OpObj{
			Op:    OpTest,
			Path:  MustParsePath("/root/list/0"),
			Value: dom.LeafNode("item2"),
		}, makeTestContainer()))
	})
	t.Run("unresolvable path", func(t *testing.T) {
		assert.Error(t, Do(&OpObj{
			Op:    OpTest,
			Path:  MustParsePath("/root/list1/0"),
			Value: dom.LeafNode("item2"),
		}, makeTestContainer()))
	})
}
