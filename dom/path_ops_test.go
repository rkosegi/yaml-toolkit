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

package dom

import (
	"testing"

	"github.com/rkosegi/yaml-toolkit/path"
	"github.com/stretchr/testify/assert"
)

func TestApplyTo(t *testing.T) {
	tgt := Builder().Container()

	type tCase struct {
		p   path.Path
		exp interface{}
		n   Node
	}

	p1 := path.NewBuilder().
		Append(path.Simple("top")).
		Append(path.Simple("list")).
		Append(path.Numeric(0)).
		Build()

	for _, tc := range []tCase{
		{
			p:   path.ChildOf(p1, path.Simple("leaf")),
			exp: "abc",
			n:   LeafNode("abc"),
		},
		{
			p:   path.ChildOf(p1, path.Simple("sub")),
			exp: 9876544,
			n:   LeafNode(9876544),
		},
		{
			// already existing node at 0 position
			p:   path.ChildOf(p1, path.Simple("sub")),
			exp: "xyz",
			n:   LeafNode("xyz"),
		},
		{
			p:   path.ChildOf(path.ParentOf(p1), path.Numeric(2), path.Numeric(1)),
			exp: 10,
			n:   LeafNode(10),
		},
	} {
		applyToNode(tgt, tc.p, tc.n)
		assert.Equal(t, tc.exp, tgt.Get(tc.p).AsLeaf().Value())
	}

	assert.Nil(t, tgt.Get(path.NewBuilder().Append(path.Simple("non-existent")).Build()))
}

func TestApplyToWithEmptyPath(t *testing.T) {
	assert.Len(t, applyToNode(Builder().Container(),
		path.NewBuilder().Build(),
		LeafNode(1)).AsContainer().Children(), 0)
}

func TestRemoveChild(t *testing.T) {
	lb := ListNode(LeafNode(1), LeafNode(2), LeafNode(3))
	assert.Equal(t, 2, lb.Items()[1].AsLeaf().Value())
	removeChild(lb, path.NewBuilder().Append(path.Numeric(1)).Build().Last())
	assert.Nil(t, lb.Items()[1])

	assert.Len(t, lb.Items(), 3)
	// this is effectively no-op, but proves that out of bound extends
	removeChild(lb, path.NewBuilder().Append(path.Numeric(5)).Build().Last())
	assert.Len(t, lb.Items(), 6)
}

func TestRemoveFromNode(t *testing.T) {
	cb := getTestDoc(t, "doc1")
	r := removeFromNode(cb, path.NewBuilder().
		Append(path.Simple("level1")).
		Append(path.Simple("level2c")).
		Append(path.Simple("level3c")).Build())
	assert.Equal(t, cb, r)
	p := path.NewBuilder().
		Append(path.Simple("level1")).
		Append(path.Simple("level2a")).
		Append(path.Simple("level3a")).Build()
	assert.NotNil(t, cb.Get(p))
	removeFromNode(cb, p)
	assert.Nil(t, cb.Get(p))
}

func TestUnsealIfNeeded(t *testing.T) {
	cb := ContainerNode()
	assert.IsType(t, &containerBuilderImpl{}, unsealContainerIfNeeded(cb.Seal()))
	assert.IsType(t, &containerBuilderImpl{}, unsealContainerIfNeeded(cb))

	lb := ListNode()
	assert.IsType(t, &listBuilderImpl{}, unsealListIfNeeded(lb.Seal()))
	assert.IsType(t, &listBuilderImpl{}, unsealListIfNeeded(lb))
}
