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
	"reflect"

	"github.com/google/go-cmp/cmp"
	"github.com/rkosegi/yaml-toolkit/common"
	"github.com/rkosegi/yaml-toolkit/path"
	"github.com/rkosegi/yaml-toolkit/query"
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
	return common.NewYamlEncoder(w).Encode(v)
}

func DefaultJsonEncoder(w io.Writer, v interface{}) error {
	e := json.NewEncoder(w)
	e.SetIndent("", "  ")
	return e.Encode(v)
}

func DefaultNodeEncoderFn(n Container) interface{} {
	return encodeContainerFn(n)
}

func DefaultNodeDecoderFn(m map[string]interface{}) Container {
	cb := *initContainerBuilder()
	decodeContainerFn(&m, &cb)
	return &cb
}

// DecodeAnyToNode decodes any value to Node.
func DecodeAnyToNode(in any) Node {
	return decodeValueToNode(reflect.ValueOf(in))
}

// YamlNodeDecoder returns function that could be used to convert yaml.Node to Node.
// There are few limitations with decoding nodes this way, e.g. leaf values are coerced to string in
// current implementation
func YamlNodeDecoder() func(n *yaml.Node) Node {
	return decodeYamlNode
}

// deprecated
// Serializable interface allows to persist data into provided io.Writer
type Serializable interface {
	// deprecated
	// Serialize writes content into given writer, while encoding using provided EncoderFunc
	Serialize(writer io.Writer, mappingFunc NodeEncoderFunc, encFn EncoderFunc) error
}

// Node is elemental unit of document. At runtime, it could be either Leaf, List or Container.
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

	// AsLeaf casts this Node to a Leaf, panics is this Node is not a Leaf.
	AsLeaf() Leaf

	// AsContainer casts this Node to a Container, panics if this Node is not a Container.
	AsContainer() Container

	// AsList casts this Node to a List, panics if this Node is not a List.
	AsList() List

	// AsAny convert this Node's value to any
	AsAny() any

	// Desc returns human-readable description of runtime type, such as "list".
	Desc() string
}

// Leaf represent Node of scalar value
type Leaf interface {
	Node
	Value() interface{}
}

// List is collection of Nodes
type List interface {
	Node
	// Get gets element at given index. Panics if value is out of bounds.
	Get(int) Node
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

	// Child returns single child Node by its name. If no such child exists, nil is returned.
	Child(name string) Node

	// deprecated, use AsAny
	// AsMap converts recursively content of this container into map[string]interface{}
	// Result consists from Go vanilla constructs only and thus could be directly used in Go templates.
	AsMap() map[string]interface{}

	// Get gets value at given path.
	Get(p path.Path) Node

	// Walk walks all child Nodes in BFS manner, until provided function return false, or all Nodes are processed.
	// TODO: I want to use DFS as well. Maybe use custom Opt?
	Walk(fn NodeVisitorFn)

	// Query queries this container and return matching child nodes.
	Query(qry query.Query) NodeList

	// deprecated, use Walk
	// Search finds all paths where Node's value is equal to given value according to provided SearchValueFunc.
	// If no match is found, nil is returned.
	Search(fn SearchValueFunc) []string
	// deprecated, use Get()
	// Lookup attempts to find child Node at given path
	Lookup(path string) Node
	// deprecated
	// Flatten flattens this Container into list of leaves
	// returned map can break structure since it uses naive property syntax to serialize keys.
	// Just try to add child Node anywhere with name that contains "."
	Flatten() map[string]Leaf
}

// ContainerBuilder is mutable extension of Container
type ContainerBuilder interface {
	Container
	// AddValue adds Node value into this Container
	AddValue(name string, value Node) ContainerBuilder
	// AddContainer adds child Container into this Container and returns it to caller.
	AddContainer(name string) ContainerBuilder
	// AddList adds child List into this Container and returns it to caller.
	AddList(name string) ListBuilder
	// Remove removes direct child Node.
	Remove(name string) ContainerBuilder
	// Merge creates new Container instance and merge current Container with other into it.
	Merge(other Container, opts ...MergeOption) ContainerBuilder
	// Seal seals the builder so that returning object will be immutable
	Seal() Container

	// Set sets node at given path.
	// Child nodes are creates as needed.
	// Upon success, this function return this ContainerBuilder to allow chaining.
	Set(p path.Path, node Node) ContainerBuilder

	// Delete removes node at given path.
	// If no such node exist, this function is no-op.
	// Upon success, this function return this ContainerBuilder to allow chaining.
	Delete(p path.Path) ContainerBuilder

	// deprecated, use Set(path, node)
	// AddValueAt adds Node into this Container at given path.
	// Child nodes are creates as needed.
	AddValueAt(path string, value Node) ContainerBuilder

	// deprecated, use Delete(path)
	// RemoveAt removes child Node at given path.
	RemoveAt(path string) ContainerBuilder
}

var (
	// CompactFn is WalkFn that you can use to compact document tree by removing empty containers.
	CompactFn = func(p path.Path, parent Node, node Node) bool {
		if node.IsContainer() {
			if len(node.(ContainerBuilder).Children()) == 0 {
				parent.(ContainerBuilder).Remove(p.Last().Value())
			}
		}
		return true
	}
)

// deprecated
type ContainerFactory interface {
	// deprecated
	// FromReader creates ContainerBuilder pre-populated with data from provided io.Reader and DecoderFunc
	FromReader(r io.Reader, fn DecoderFunc) (ContainerBuilder, error)
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
	// Seal seals the builder so that returning object will be immutable
	Seal() List
}

// MergeOption is function used to customize merger behavior
type MergeOption func(*merger)

// OverlayVisitorFn visits every Node in each layer within the OverlayDocument.
// Returning false from this function will terminate iteration in current layer.
type OverlayVisitorFn func(layer string, p path.Path, parent Node, node Node) bool

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
	// Walk walks every layer in this document and visits every node in BFS manner.
	// TODO: I want to use DFS as well. Maybe use custom Opt?
	Walk(fn OverlayVisitorFn)
}
