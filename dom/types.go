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
	"github.com/google/go-cmp/cmp"
	"github.com/rkosegi/yaml-toolkit/utils"
	"gopkg.in/yaml.v3"
	"io"
)

// SearchValueFunc is used to search for value within document
type SearchValueFunc func(val interface{}) bool

// SearchEqual is SearchValueFunc that search for equivalent value
func SearchEqual(in interface{}) SearchValueFunc {
	return func(val interface{}) bool {
		return cmp.Equal(val, in)
	}
}

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
	return utils.NewYamlEncoder(w).Encode(v)
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
	// IsContainer returns true if this node is Container
	IsContainer() bool
	// IsList returns true if this node is List
	IsList() bool
	// IsLeaf returns true if this node is Leaf
	IsLeaf() bool
	// SameAs returns true if this node is of same type like other node.
	// The other Node can be nil, in which case return value is false.
	SameAs(Node) bool
}

// Leaf represent Node of scalar value
type Leaf interface {
	Node
	Value() interface{}
}

// List is collection of Nodes
type List interface {
	Node
	// Items returns copy of slice of all nodes in this list
	Items() []Node
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
	// FindValue finds all paths where Node's value is equal to given value.
	// If no match is found, nil is returned.
	// Deprecated: use Search(SearchEqual(x))
	FindValue(val interface{}) []string
	// Search finds all paths where Node's value is equal to given value according to provided SearchValueFunc.
	// If no match is found, nil is returned.
	Search(fn SearchValueFunc) []string
}

// ContainerBuilder is mutable extension of Container
type ContainerBuilder interface {
	Container
	Serializable
	// AddValue adds Node value into this Container
	AddValue(name string, value Node)
	// AddValueAt adds Leaf value into this Container at given path.
	// Child nodes are creates as needed.
	AddValueAt(path string, value Node)
	// AddContainer adds child Container into this Container
	AddContainer(name string) ContainerBuilder
	// AddList adds child List into this Container
	AddList(name string) ListBuilder
	// Remove removes direct child Node.
	Remove(name string)
	// RemoveAt removes child Node at given path.
	RemoveAt(path string)
	// Walk walks whole document tree, visiting every node
	Walk(fn WalkFn)
}

type WalkFn func(path string, parent ContainerBuilder, node Node) bool

var (
	CompactFn = func(path string, parent ContainerBuilder, node Node) bool {
		if node.IsContainer() {
			if len(node.(ContainerBuilder).Children()) == 0 {
				parent.Remove(path)
			}
		}
		return true
	}
)

type ContainerFactory interface {
	// Container creates empty ContainerBuilder
	Container() ContainerBuilder
	// FromReader creates ContainerBuilder pre-populated with data from provided io.Reader and DecoderFunc
	FromReader(r io.Reader, fn DecoderFunc) (ContainerBuilder, error)
	// FromMap creates ContainerBuilder pre-populated with data from provided map
	FromMap(in map[string]interface{}) ContainerBuilder
	// FromAny creates ContainerBuilder from any object. Any error encountered in process will result in panic.
	// This method uses YAML codec internally to perform translation between raw interface and map
	FromAny(v interface{}) ContainerBuilder
}

// Coordinate is address of Node within OverlayDocument
type Coordinate interface {
	// Layer returns name of layer within OverlayDocument
	Layer() string
	// Path returns path to Node within layer
	Path() string
}

type ListBuilder interface {
	List
	// Clear sets items to empty slice
	Clear()
	// Set sets item at given index. Items are allocated and set to nil Leaf as necessary.
	Set(uint, Node)
	// MustSet sets item at given index. Panics if index is out of bounds.
	MustSet(uint, Node)
	// Append adds new item at the end of slice
	Append(Node)
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
	// FindValue find all occurrences of given value in all layers
	// Deprecated: use Search(SearchEqual(x))
	FindValue(val interface{}) []Coordinate
	// Search finds all occurrences of given value in all layers using custom SearchValueFunc
	Search(fn SearchValueFunc) []Coordinate
	// Populate puts dictionary into overlay at given path
	Populate(overlay, path string, data *map[string]interface{})
	// Put puts Node value into overlay at given path
	Put(overlay, path string, value Node)
	// Merged returns read-only, merged view of all overlays
	Merged() Container
	// Layers return copy of list with all layer names
	Layers() []string
}
