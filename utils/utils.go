/*
Copyright 2023 Richard Kosegi

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

package utils

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

var (
	FileOpener = os.Open
)

// ToPath creates path from path and key name
func ToPath(path, key string) string {
	if len(path) == 0 {
		return key
	} else {
		return fmt.Sprintf("%s.%s", path, key)
	}
}

// ToListPath like ToPath, but for lists
func ToListPath(path string, index int) string {
	sub := fmt.Sprintf("[%d]", index)
	if len(path) == 0 {
		return sub
	} else {
		return path + sub
	}
}

var listPropRe = regexp.MustCompile(".*(\\[\\d+])+")

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

// NewYamlEncoder creates and uniformly configures yaml.Encoder across project
func NewYamlEncoder(w io.Writer) *yaml.Encoder {
	enc := yaml.NewEncoder(w)
	enc.SetIndent(2)
	return enc
}

type failingReader struct{}

func (fr failingReader) Read([]byte) (int, error) {
	return 0, errors.New("read just failed")
}

func FailingReader() io.Reader {
	return &failingReader{}
}

type failingWriter struct{}

func (fw failingWriter) Write([]byte) (int, error) {
	return 0, errors.New("write just failed")
}

func FailingWriter() io.Writer {
	return &failingWriter{}
}

func Unique(in []string) []string {
	ret := make([]string, 0)
	for _, s := range in {
		if !slices.Contains(ret, s) {
			ret = append(ret, s)
		}
	}
	return ret
}
