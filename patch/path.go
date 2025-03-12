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
	"fmt"
	"strconv"
	"strings"

	"github.com/rkosegi/yaml-toolkit/dom"
	"github.com/rkosegi/yaml-toolkit/path"
)

type PathSegment string

// TODO add support for "-" - the "last" item designation
func (ps PathSegment) IsNumeric() (int, bool) {
	i, err := strconv.Atoi(string(ps))
	return i, err == nil
}

type Path []PathSegment

func (p Path) Parent() Path {
	if len(p) <= 1 {
		return nil
	}
	return p[0 : len(p)-1]
}

func (p Path) LastSegment() PathSegment {
	if len(p) == 0 {
		return ""
	}
	return p[len(p)-1]
}

// String returns lexical representation of this path according to RFC
func (p Path) String() string {
	if len(p) == 0 {
		return ""
	}
	var sb strings.Builder
	for _, ps := range p {
		sb.WriteRune('/')
		rps := []rune(ps)
		for i := 0; i < len(rps); i++ {
			if rps[i] == '~' {
				sb.WriteString("~0")
			} else if rps[i] == '/' {
				sb.WriteString("~1")
			} else {
				sb.WriteRune(rps[i])
			}
		}
	}
	return sb.String()
}

var emptyPath = Path{}

func MustParsePath(pathSpec string) Path {
	p, err := ParsePath(pathSpec)
	if err != nil {
		panic(err)
	}
	return p
}

func ParsePath(pathSpec string) (Path, error) {
	if pathSpec == "" {
		// special case that match anything
		return emptyPath, nil
	}
	if !strings.HasPrefix(pathSpec, "/") {
		return nil, fmt.Errorf("path must start with '/':%s", pathSpec)
	}
	// strip leading slash
	pathSpec = pathSpec[1:]
	rps := []rune(pathSpec)
	ps := make([]PathSegment, 0)
	cs := strings.Builder{}
	for i := 0; i < len(rps); i++ {
		c := rps[i]
		switch c {
		case '~':
			// check if there is room to look ahead
			if i < len(rps)-1 {
				// https://datatracker.ietf.org/doc/html/rfc6901#section-4
				// first transforming any occurrence of the sequence '~1' to '/'
				if rps[i+1] == '1' {
					cs.WriteRune('/')
					i++
				} else
				// then transforming any occurrence of the sequence '~0' to '~'
				if rps[i+1] == '0' {
					cs.WriteRune('~')
					i++
				} else {
					// not a special case, just regular character
					cs.WriteRune(c)
				}
			} else {
				// tilda is very last character
				cs.WriteRune(c)
			}
		case '/':
			ps = append(ps, PathSegment(cs.String()))
			cs.Reset()
		default:
			cs.WriteRune(c)
		}
	}
	ps = append(ps, PathSegment(cs.String()))
	cs.Reset()
	return ps, nil
}

// Eval evaluates path within provided target dom.Container. Call to this function returns pair,
// where one part is series of Nodes encountered while walking from root of target dom.Container down the path, second part is final Node.
// If path does not fully resolve, second part will be nil, while first part will contain all Nodes up to point where resolution stopped.
func (p Path) Eval(target dom.Container) (dom.NodeList, dom.Node) {
	var (
		curr dom.Node
		res  dom.NodeList
	)
	// special case, empty path resolves to "whole document" so return target as-is.
	if len(p) == 0 {
		return []dom.Node{target}, target
	}
	curr = target
	res = make([]dom.Node, 0)

	for _, ps := range p {
		// first try list item
		if curr.IsList() {
			if idx, isIndex := ps.IsNumeric(); isIndex {
				l := curr.(dom.List)
				if len(l.Items()) > idx {
					curr = l.Items()[idx]
					res = append(res, curr)
				} else {
					// list index out of bounds
					return res, nil
				}
			}
		} else
		// regular child within container
		{
			if !curr.IsContainer() {
				return res, nil
			} else {
				curr = curr.(dom.Container).Child(string(ps))
				if curr == nil {
					return res, nil
				}
				res = append(res, curr)
			}
		}
	}
	return res, curr
}

type pathSupport struct{}

func (ps pathSupport) Serialize(in path.Path) string {
	if in.IsEmpty() {
		return ""
	}
	var (
		sb strings.Builder
		pc path.Component
	)
	for _, pc = range in.Components() {
		sb.WriteRune('/')
		rps := []rune(pc.Value())
		for i := 0; i < len(rps); i++ {
			if rps[i] == '~' {
				sb.WriteString("~0")
			} else if rps[i] == '/' {
				sb.WriteString("~1")
			} else {
				sb.WriteRune(rps[i])
			}
		}
	}
	return sb.String()
}

func (ps pathSupport) MustParse(s string) path.Path {
	if x, err := ps.Parse(s); err != nil {
		panic(err)
	} else {
		return x
	}
}

func (ps pathSupport) Parse(in string) (path.Path, error) {
	p, err := ParsePath(in)
	if err != nil {
		return nil, err
	}
	b := path.NewBuilder()
	for _, pc := range p {
		b.Append(path.Simple(string(pc)))
	}
	return b.Build(), nil
}

func NewPathParser() path.Parser {
	return pathSupport{}
}

func NewPathSerializer() path.Serializer {
	return pathSupport{}
}
