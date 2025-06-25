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

package props

import (
	"errors"
	"testing"

	"github.com/rkosegi/yaml-toolkit/path"
	"github.com/stretchr/testify/assert"
)

type failingRuneReader struct {
}

func (f *failingRuneReader) ReadRune() (r rune, size int, err error) {
	return 0, 0, errors.New("failing rune reader")
}

func TestPathParser(t *testing.T) {
	var (
		p   path.Path
		err error
	)

	p, err = NewPathParser().Parse("root.part1.list[0].sub2")
	assert.NoError(t, err)
	assert.NotNil(t, p)
	assert.Equal(t, 5, len(p.Components()))
	assert.Equal(t, "sub2", p.Last().Value())

	p, err = newPathSupport('.').Parse("root.files.application\\.yaml")
	assert.NoError(t, err)
	assert.NotNil(t, p)
	assert.Equal(t, 3, len(p.Components()))
	assert.Equal(t, "application.yaml", p.Last().Value())

	_, err = newPathSupport('.').parse(&failingRuneReader{})
	assert.Error(t, err)
}

func TestPathParserMustParseFail(t *testing.T) {
	defer func() {
		recover()
	}()
	newPathSupport('.').mustParse(&failingRuneReader{})
	assert.Fail(t, "failing")
}

func TestPathParserMustParsePass(t *testing.T) {
	p := NewPathParser().MustParse("root.part1.list[0].sub2")
	assert.NotNil(t, p)
	assert.Equal(t, 5, len(p.Components()))
	assert.Equal(t, "sub2", p.Last().Value())
}

func TestPathParseSerialize(t *testing.T) {
	var (
		err error
		x   path.Path
	)
	s := NewPathSerializer()
	p := NewPathParser()

	type testcase struct {
		p    string
		fail bool
		c    int
	}
	for _, tc := range []testcase{
		{
			p:    "root.sub1",
			fail: false,
			c:    2,
		},
		{
			p:    "root.list[1].sublist[3].sub4",
			fail: false,
			c:    6,
		},
		{
			p:    "",
			fail: false,
			c:    0,
		},
	} {
		x, err = p.Parse(tc.p)
		if tc.fail {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, tc.c, len(x.Components()))
		}
		assert.Equal(t, tc.p, s.Serialize(x))
	}
}
