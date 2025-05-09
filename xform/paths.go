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

package xform

import (
	"fmt"
	"strings"

	"github.com/rkosegi/yaml-toolkit/patch"
	"github.com/rkosegi/yaml-toolkit/path"
	"github.com/rkosegi/yaml-toolkit/props"
)

func PropPath2Pointer(in path.Path) patch.Path {
	var r strings.Builder
	for _, pc := range in.Components() {
		if pc.IsNumeric() {
			r.WriteString(fmt.Sprintf("/%d", pc.NumericValue()))
		} else {
			r.WriteString(fmt.Sprintf("/%s", pc.Value()))
		}
	}
	return patch.MustParsePath(r.String())
}

func PointerFromPropPathString(raw string) patch.Path {
	return PropPath2Pointer(props.NewPathParser().MustParse(raw))
}
