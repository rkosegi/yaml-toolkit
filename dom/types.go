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

package dom

import (
	"encoding/json"
	"gopkg.in/yaml.v3"
	"io"
)

// NodeMappingFunc maps internal Container value into external data representation
type NodeMappingFunc func(Container) interface{}

// EncoderFunc encodes raw value into stream using provided io.Writer
type EncoderFunc func(w io.Writer, v interface{}) error

// DecoderFunc decodes byte stream into raw value using provided io.Reader
type DecoderFunc func(r io.Reader, v interface{}) error

func DefaultYamlDecoder(r io.Reader, v interface{}) error {
	return yaml.NewDecoder(r).Decode(v)
}

func DefaultJsonDecoder(r io.Reader, v interface{}) error {
	return json.NewDecoder(r).Decode(v)
}

func DefaultYamlEncoder(w io.Writer, v interface{}) error {
	e := yaml.NewEncoder(w)
	e.SetIndent(2)
	return e.Encode(v)
}

func DefaultJsonEncoder(w io.Writer, v interface{}) error {
	e := json.NewEncoder(w)
	e.SetIndent("", "  ")
	return e.Encode(v)
}

func DefaultNodeMappingFn(n Container) interface{} {
	return containerMappingFn(n.(Container))
}

// Serializable interface allows to persist data into provided io.Writer
type Serializable interface {
	// Serialize writes content into given writer, while encoding using provided EncoderFunc
	Serialize(writer io.Writer, mappingFunc NodeMappingFunc, encFn EncoderFunc) error
}

// Node is elemental unit of document. At runtime, it could be either Leaf or Container.
type Node interface {
	IsContainer() bool
}

// Leaf represent Node of scalar value
type Leaf interface {
	Node
	Value() interface{}
}

// Container is element that has zero or more child Nodes
type Container interface {
	Node
	// Children returns mapping between child name and its corresponding Node
	Children() map[string]Node
	// Child returns single child Node by its name
	Child(name string) Node
	// Lookup attempts to find child Node at given path
	Lookup(path string) Node
	// Flatten flattens this Container into list of leaves
	Flatten() map[string]Leaf
}

type ContainerBuilder interface {
	Container
	Serializable
	// AddValue adds Leaf value into this Container
	AddValue(name string, value Leaf)
	// AddContainer adds child Container into this Container
	AddContainer(name string) ContainerBuilder
	// Remove child
	Remove(name string)
}

type ContainerFactory interface {
	// Container creates empty ContainerBuilder
	Container() ContainerBuilder
	// FromReader creates ContainerBuilder pre-populated with data from provided io.Reader and DecoderFunc
	FromReader(r io.Reader, fn DecoderFunc) (ContainerBuilder, error)
}

// OverlayDocument represents multiple documents layered over each other.
// It allows lookup across all layers while respecting precedence
type OverlayDocument interface {
	Serializable
	// Lookup lookups data in given overlay and path
	// if no node is present at any level, nil is returned
	Lookup(overlay, path string) Node
	// LookupAny lookups data in given all overlays (in creation order) and path.
	// if no node is present at any level, nil is returned
	LookupAny(path string) Node
	// Populate puts dictionary into overlay at given path
	Populate(overlay, path string, data *map[string]interface{})
	// Put puts Node value into overlay at given path
	Put(overlay, path string, value Node)
	// Merged returns read-only, merged view of all overlays
	Merged() Container
}
