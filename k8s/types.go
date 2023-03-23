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

package k8s

import (
	"github.com/rkosegi/yaml-toolkit/dom"
	"io"
)

// StringData allows to manipulate content of string data items in k8s manifest
type StringData interface {
	// Get gets data item
	Get(key string) *string
	// List gets all data key
	List() []string
	// Remove removes data
	Remove(key string)
	// Update adds or updates data
	Update(key, value string)
}

// BinaryData allows to manipulate content of binary data items in k8s manifest
type BinaryData interface {
	// Get gets binary data item by its name
	Get(key string) []byte
	// List gets all binary data keys
	List() []string
	// Remove removes given binary data item
	Remove(key string)
	// Update adds or updates binary data item
	Update(key string, value []byte)
}

// Manifest is in-memory representation of k8s Secret/ConfigMap manifest.
// It allows to perform operation over binary and string data inside manifest
type Manifest interface {
	io.WriterTo
	// StringData obtains interface to manipulate content of string data items
	StringData() StringData
	// BinaryData obtains interface to manipulate content of binary data items
	BinaryData() BinaryData
}

// Document interface allows interaction with document embedded inside k8s manifest
type Document interface {
	// Document gets handle to embedded document root ContainerBuilder
	Document() dom.ContainerBuilder
	// Save persists any changes made to embedded document
	Save() error
}

// EncodeInternalFn is responsible for taking ContainerBuilder and encode it into provided Manifest
type EncodeInternalFn func(m Manifest, node dom.ContainerBuilder) error

// DecodeInternalFn is responsible for taking Manifest and producing ContainerBuilder
type DecodeInternalFn func(m Manifest) (dom.ContainerBuilder, error)
