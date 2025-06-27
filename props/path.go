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
	"io"
	"strings"

	"github.com/rkosegi/yaml-toolkit/path"
	"github.com/rkosegi/yaml-toolkit/utils"
)

type pathSupport struct {
	// delimiter, usually '.'
	d rune
}

func (p *pathSupport) MustParse(s string) path.Path {
	return p.mustParse(strings.NewReader(s))
}

func (p *pathSupport) mustParse(r io.RuneReader) path.Path {
	if x, err := p.parse(r); err != nil {
		panic(err)
	} else {
		return x
	}
}

func (p *pathSupport) parse(r io.RuneReader) (path.Path, error) {
	prev := rune(0)
	sb := strings.Builder{}
	segments := make([]string, 0)
	for {
		if c, _, err := r.ReadRune(); err != nil {
			if err == io.EOF {
				// don't append empty component
				t := strings.ReplaceAll(sb.String(), "\\.", ".")
				if len(t) > 0 {
					segments = append(segments, t)
					// b.Append(path.Simple(t))
				}
				break
			} else {
				return nil, err
			}
		} else {
			if c == p.d {
				if prev == '\\' {
					// escaped delimiter is just a part of path component
					sb.WriteRune(p.d)
					prev = c
				} else {
					// flush current component to builder and reset state
					segments = append(segments, sb.String())
					sb.Reset()
					prev = rune(0)
				}
			} else {
				sb.WriteRune(c)
				prev = c
			}
		}
	}
	b := path.NewBuilder()
	for _, seg := range segments {
		if n, idxes, isListItem := utils.ParseListPathComponent(seg); isListItem {
			b = b.Append(path.Simple(n))
			for _, idx := range idxes {
				b = b.Append(path.Numeric(idx))
			}
		} else {
			b = b.Append(path.Simple(seg))
		}
	}
	return b.Build(), nil
}
func (p *pathSupport) Parse(s string) (path.Path, error) {
	return p.parse(strings.NewReader(s))
}

func (p *pathSupport) Serialize(in path.Path) string {
	var sb strings.Builder
	cp := in.Components()
	lastIdx := len(cp) - 1

	for idx, c := range cp {
		nextIsNum := false
		if idx+1 < len(cp) && cp[idx+1].IsNumeric() {
			nextIsNum = true
		}
		if c.IsNumeric() {
			sb.WriteRune('[')
			sb.WriteString(c.Value())
			sb.WriteRune(']')
		} else {
			sb.WriteString(c.Value())
		}
		if idx < lastIdx && !nextIsNum {
			sb.WriteRune(p.d)
		}
	}
	return sb.String()
}

func newPathSupport(delim rune) *pathSupport {
	return &pathSupport{
		d: delim,
	}
}

func NewPathParser() path.Parser {
	return newPathSupport('.')
}

func NewPathSerializer() path.Serializer {
	return newPathSupport('.')
}
