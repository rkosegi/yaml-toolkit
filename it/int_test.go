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

package it

import (
	"github.com/rkosegi/yaml-toolkit/dom"
	"github.com/rkosegi/yaml-toolkit/k8s"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestLoadListOverlay(t *testing.T) {
	f, err := os.CreateTemp("", "yt.*.yaml")
	defer func() {
		_ = os.Remove(f.Name())
	}()
	assert.NoError(t, err)
	_, err = f.WriteString(`
kind: Secret
metadata:
  name: secret2
apiVersion: v1
stringData:
  root.sub.list[0].prop: 123
  root.sub.list[1].list2[2]: "abc"
  root.sub.list[2].prop: "def"
  root.sub.list[3]: "xyz"
`)
	assert.NoError(t, f.Close())
	assert.NoError(t, err)

	doc, err := k8s.Properties(f.Name())
	assert.NoError(t, err)

	od := dom.NewOverlayDocument()
	od.Put("main", "", doc.Document())
	od.Put("other", "root.sub.list[0].prop2", dom.LeafNode("000"))

	assert.Equal(t, "abc", od.LookupAny("root.sub.list[1].list2[2]").(dom.Leaf).Value())

	m := od.Merged().Flatten()
	assert.Equal(t, 7, len(m))
}
