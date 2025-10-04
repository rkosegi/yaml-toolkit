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

package common

import (
	"errors"
	"io"
	"os"
	"regexp"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

// EmptyChecker interface can be used to check is implementing struct is "empty".
type EmptyChecker interface {
	// IsEmpty returns true if this instance is "empty".
	IsEmpty() bool
}

type StringPredicateFn func(string) bool

var (
	MatchAny = func() StringPredicateFn {
		return func(s string) bool {
			return true
		}
	}
	MatchNone = func() StringPredicateFn {
		return func(s string) bool {
			return false
		}
	}
	MatchRe = func(re *regexp.Regexp) StringPredicateFn {
		return func(s string) bool {
			return re.MatchString(s)
		}
	}
)

var (
	FileOpener = os.Open
)

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

// Unflatten map entries into new map.
func Unflatten(in map[string]interface{}) map[string]interface{} {
	res := make(map[string]interface{})
	for k, v := range in {
		current := res
		pc := strings.Split(k, ".")
		for _, c := range pc[0 : len(pc)-1] {
			if x, exists := current[c].(map[string]interface{}); exists {
				current = x
			} else {
				current[c] = make(map[string]interface{})
				current = current[c].(map[string]interface{})
			}
		}
		current[pc[len(pc)-1]] = v
	}
	return res
}
