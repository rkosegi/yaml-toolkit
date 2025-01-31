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
	"fmt"
	"strings"

	"github.com/rkosegi/yaml-toolkit/utils"
)

type PathSegment struct {
	Index int
	IsNum bool
	Value string
}

type Path []PathSegment

func (ps PathSegment) String() string {
	if ps.IsNum {
		return fmt.Sprintf("%d", ps.Index)
	} else {
		return ps.Value
	}
}

// ParsePath parses path specification in string form into Path.
func ParsePath(raw string) Path {
	q := Path{}
	raw = strings.TrimFunc(strings.TrimSpace(raw), func(r rune) bool {
		return r == '.'
	})
	parts := strings.Split(raw, ".")

	for _, c := range parts {
		if n, idxes, ok := utils.ParseListPathComponent(c); ok {
			q = append(q, PathSegment{Value: n})
			for _, idx := range idxes {
				q = append(q, PathSegment{
					Index: idx,
					IsNum: true,
				})
			}
		} else {
			q = append(q, PathSegment{Value: c})
		}
	}
	return q
}
