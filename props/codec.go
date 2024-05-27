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
	"github.com/magiconair/properties"
	"io"
)

func EncoderFn(w io.Writer, x interface{}) error {
	for k, v := range x.(map[string]interface{}) {
		_, err := w.Write([]byte(fmt.Sprintf("%s=%v\n", k, v)))
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
	m := x.(*map[string]interface{})
	for k, v := range p.Map() {
		(*m)[k] = v
	}
	return nil
}
