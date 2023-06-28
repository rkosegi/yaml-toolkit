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
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
)

// ToPath creates path from path and key name
func ToPath(path, key string) string {
	if len(path) == 0 {
		return key
	} else {
		return fmt.Sprintf("%s.%s", path, key)
	}
}

// NewYamlEncoder creates and uniformly configures yaml.Encoder across project
func NewYamlEncoder(w io.Writer) *yaml.Encoder {
	enc := yaml.NewEncoder(w)
	enc.SetIndent(2)
	return enc
}
