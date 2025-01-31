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
	"io"

	"github.com/google/go-cmp/cmp"
	"github.com/rkosegi/yaml-toolkit/utils"
	"gopkg.in/yaml.v3"
)

// SearchValueFunc is used to search for value within document
type SearchValueFunc func(val interface{}) bool

// SearchEqual is SearchValueFunc that search for equivalent value
func SearchEqual(in interface{}) SearchValueFunc {
	return func(val interface{}) bool {
		return cmp.Equal(val, in)
	}
}

// NodeMappingFunc maps internal Container value into external data representation.
// Deprecated. Use NodeEncoderFunc
type NodeMappingFunc func(Container) interface{}

// NodeEncoderFunc maps internal Container value into external data representation
type NodeEncoderFunc func(Container) interface{}

// NodeDecoderFunc takes external data representation and decode it to Container
type NodeDecoderFunc func(map[string]interface{}) Container

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

func DefaultNodeEncoderFn(n Container) interface{} {
	return DefaultNodeMappingFn(n)
}

// Deprecated. Use DefaultNodeEncoderFn
func DefaultNodeMappingFn(n Container) interface{} {
	return encodeContainerFn(n)
}

func DefaultNodeDecoderFn(m map[string]interface{}) Container {
	cb := containerBuilderImpl{}
	decodeContainerFn(&m, &cb)
	return &cb
}

// Serializable interface allows to persist data into provided io.Writer
type Serializable interface {
	// Serialize writes content into given writer, while encoding using provided EncoderFunc
	Serialize(writer io.Writer, mappingFunc NodeEncoderFunc, encFn EncoderFunc) error
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
	// Equals return true is value of this node is equal to value of provided Node.
	Equals(Node) bool
	// Clone returns deep copy of this Node.
	Clone() Node
}

// Leaf represent Node of scalar value
type Leaf interface {
	Node
	Value() interface{}
}

// List is collection of Nodes
type List interface {
	Node
	// Size returns current count of items in this list
	Size() int
	// Items returns copy of slice of all nodes in this list
	Items() []Node

	// AsSlice converts recursively content of this List into []interface{}.
	// Result consists from Go vanilla constructs only.
	AsSlice() []interface{}
}

// NodeList is sequence of zero or more Nodes
type NodeList []Node

// Container is element that has zero or more child Nodes
type Container interface {
	Node
	Serializable
	// Children returns mapping between child name and its corresponding Node
	Children() map[string]Node
	// Child returns single child Node by its name
	Child(name string) Node
	// Lookup attempts to find child Node at given path
	Lookup(path string) Node
	// Flatten flattens this Container into list of leaves
	Flatten() map[string]Leaf

	// AsMap converts recursively content of this container into map[string]interface{}
	// Result consists from Go vanilla constructs only and thus could be directly used in Go templates.
	AsMap() map[string]interface{}
	// Search finds all paths where Node's value is equal to given value according to provided SearchValueFunc.
	// If no match is found, nil is returned.
	Search(fn SearchValueFunc) []string
}

// ContainerBuilder is mutable extension of Container
type ContainerBuilder interface {
	Container
	// AddValue adds Node value into this Container
	AddValue(name string, value Node) ContainerBuilder
	// AddValueAt adds Leaf value into this Container at given path.
	// Child nodes are creates as needed.
	AddValueAt(path string, value Node) ContainerBuilder
	// AddContainer adds child Container into this Container
	AddContainer(name string) ContainerBuilder
	// AddList adds child List into this Container
	AddList(name string) ListBuilder
	// Remove removes direct child Node.
	Remove(name string) ContainerBuilder
	// RemoveAt removes child Node at given path.
	RemoveAt(path string) ContainerBuilder
	// Walk walks whole document tree, visiting every node
	Walk(fn WalkFn)
	// Merge creates new Container instance and merge current Container with other into it.
	Merge(other Container, opts ...MergeOption) ContainerBuilder
}

type WalkFn func(path string, parent ContainerBuilder, node Node) bool

// OverlayVisitorFn visits every node in OverlayDocument.
// Returning false from this function will cause termination of process.
type OverlayVisitorFn func(layer, path string, parent Node, node Node) bool

var (
	// CompactFn is WalkFn that you can use to compact document tree by removing empty containers.
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
	// FromProperties is similar to FromMap except that keys are parsed into path before inserting into ContainerBuilder
	FromProperties(in map[string]interface{}) ContainerBuilder
}

// Coordinate is address of Node within OverlayDocument
type Coordinate interface {
	// Layer returns name of layer within OverlayDocument
	Layer() string
	// Path returns path to Node within layer
	Path() string
}

// Coordinates is collection of Coordinate
type Coordinates []Coordinate

type ListBuilder interface {
	List
	// Clear sets items to empty slice
	Clear() ListBuilder
	// Set sets item at given index. Items are allocated and set to nil Leaf as necessary.
	Set(uint, Node) ListBuilder
	// MustSet sets item at given index. Panics if index is out of bounds.
	MustSet(uint, Node) ListBuilder
	// Append adds new item at the end of slice
	Append(Node) ListBuilder
}

// MergeOption is function used to customize merger behavior
type MergeOption func(*merger)

// OverlayDocument represents multiple documents layered over each other.
// It allows lookup across all layers while respecting precedence
type OverlayDocument interface {
	Serializable
	// Lookup lookups data in given overlay and path
	// if no node is present at any level, nil is returned
	Lookup(overlay, path string) Node
	// LookupAny lookups data in all overlays (in creation order) and path.
	// if no node is present at any level, nil is returned
	LookupAny(path string) Node
	// Search finds all occurrences of given value in all layers using custom SearchValueFunc
	Search(fn SearchValueFunc) Coordinates
	// Populate puts dictionary into overlay at given path
	Populate(overlay, path string, data *map[string]interface{})
	// Add adds elements from given Container into root of given layer
	Add(overlay string, value Container)
	// Put puts Node value into overlay at given path
	Put(overlay, path string, value Node)
	// Merged returns read-only, merged view of all overlays
	Merged(option ...MergeOption) Container
	// Layers returns a copy of mapping between layer name and its associated Container.
	// Containers are cloned using Node.Clone()
	Layers() map[string]Container
	// LayerNames returns copy of list of all layer names, in insertion order
	LayerNames() []string
	// Walk walks every layer in this document and visits every node.
	Walk(fn OverlayVisitorFn)
}
