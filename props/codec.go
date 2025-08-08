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
	"io"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/magiconair/properties"
	"github.com/rkosegi/yaml-toolkit/common"
	"github.com/rkosegi/yaml-toolkit/dom"
)

func encodeKv(k string, v interface{}, w io.Writer) error {
	_, err := w.Write([]byte(fmt.Sprintf("%s=%v\n", k, v)))
	return err
}

func EncoderFn(w io.Writer, x interface{}) error {
	for k, v := range x.(map[string]interface{}) {
		err := encodeKv(k, v, w)
		if err != nil {
			return err
		}
	}
	return nil
}

func DomEncoderFn(w io.Writer, x interface{}) error {
	for k, v := range x.(dom.Container).Children() {
		err := encodeKv(k, v.(dom.Leaf).Value(), w)
		if err != nil {
			return err
		}
	}
	return nil
}

func DecoderFn(r io.Reader, x interface{}) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	p, _ := properties.Load(data, properties.UTF8)
	m2 := make(map[string]interface{})
	for k, v := range p.Map() {
		m2[k] = v
	}
	out := make(map[string]interface{})
	for k, v := range common.Unflatten(m2) {
		out[k] = v
	}
	// maybe check if xv := reflect.ValueOf(x); xv.Kind() != reflect.Pointer || xv.IsNil() ??
	reflect.ValueOf(x).Elem().Set(reflect.ValueOf(out))
	return nil
}

var listPropRe = regexp.MustCompile(`.*(\[\d+])+`)

func ParseListPathComponent(path string) (string, []int, bool) {
	if !listPropRe.MatchString(path) {
		return "", nil, false
	}
	indexes := make([]int, 0)
	first := strings.Index(path, "[")
	cpath := path
	for {
		start := strings.Index(cpath, "[")
		if start == -1 {
			return path[0:first], indexes, true
		}
		end := strings.Index(cpath, "]")
		index, _ := strconv.Atoi(cpath[start+1 : end])
		indexes = append(indexes, index)
		cpath = cpath[end+1:]
	}
}
