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

package analytics

import (
	"github.com/rkosegi/yaml-toolkit/dom"
	"io"
)

// FileDecoderProvider resolves dom.DecoderFunc for given file.
// If file is not recognized, nil is returned.
type FileDecoderProvider func(file string) dom.DecoderFunc

// AddLayerOpt returns function that can be used to customize document being added to set.
type AddLayerOpt func(*documentSet, string, *docContext)

// DocumentSet is interface that allows to interact with multiple documents in simple way
type DocumentSet interface {
	// TaggedSubset creates overlay document from all documents matching at least one of given tags
	TaggedSubset(tag ...string) dom.OverlayDocument

	// AsOne creates overlay document from all documents in this DocumentSet.
	AsOne() dom.OverlayDocument

	// AddDocument adds given document into documentSet.
	AddDocument(name string, doc dom.ContainerBuilder, opts ...AddLayerOpt)

	// AddUnnamedDocument adds document that has no particular name designation, one will be generated internally
	AddUnnamedDocument(doc dom.ContainerBuilder, opts ...AddLayerOpt)

	// AddDocumentFromFile calls AddDocumentFromFileWithDecoder with second argument set to DefaultFileDecoderProvider.
	AddDocumentFromFile(file string, dec dom.DecoderFunc, opts ...AddLayerOpt) error

	// AddDocumentFromReader reads given io.Reader using dom.DecoderFunc into document and adds it into this documentSet
	AddDocumentFromReader(string, io.Reader, dom.DecoderFunc, ...AddLayerOpt) error

	// AddPropertiesFromManifest loads string data from provided manifest as properties and adds them into this documentSet.
	AddPropertiesFromManifest(file string, opts ...AddLayerOpt) error

	// AddDocumentsFromDirectory takes provided glob pattern and loads all matching files into this documentSet.
	AddDocumentsFromDirectory(pattern string, decProvFn FileDecoderProvider, opts ...AddLayerOpt) error

	// AddDocumentsFromManifest parses K8s manifest data entries into dom.ContainerBuilder and adds them into documentSet.
	// See k8s package for more details about manifest support details.
	// Warning: invocation of this function is not atomic if error occurs mid-execution;
	// some manifest items might be added to DocumentSet before error occurred, while rest of them not.
	AddDocumentsFromManifest(manifest string, decProvFn FileDecoderProvider, opts ...AddLayerOpt) error
}

// OnPlaceholderEncounteredFn is invoked when property value that contains placeholder
// is encountered during resolution process.
type OnPlaceholderEncounteredFn func(key, ph string)

// StringPredicateFn is predicate to match string value.
type StringPredicateFn func(string) bool

// OnResolutionFailureFn is callback function invoked when placeholder can't be resolved to actual value
type OnResolutionFailureFn func(key, value string, coordinates dom.Coordinates)

// PlaceholderResolutionReport encompasses result of placeholder resolution process.
type PlaceholderResolutionReport struct {
	// FailedKeys is collection of all keys that failed placeholder resolution
	FailedKeys []string
	// ActualValues is mapping between keys (those that failed resolution) and their actual values
	// as observed during resolution process.
	ActualValues map[string]interface{}
	// Coordinates is mapping between failed keys and dom.Coordinates where placeholders are found
	Coordinates map[string]dom.Coordinates
}

type PlaceholderResolver interface {
	Resolve(doc dom.OverlayDocument) *PlaceholderResolutionReport
}

// PlaceholderResolverBuilder is used to create instance of PlaceholderResolver
type PlaceholderResolverBuilder interface {
	// WithPlaceholderMatcher allows to override default predicate to match presence of placeholder in property value.
	// Normally you don't need to override it.
	WithPlaceholderMatcher(StringPredicateFn) PlaceholderResolverBuilder
	// WithKeyFilter allows to set filter to narrow down resolution only to keys matching provided predicate
	WithKeyFilter(StringPredicateFn) PlaceholderResolverBuilder

	OnPlaceholderEncountered(OnPlaceholderEncounteredFn) PlaceholderResolverBuilder

	// OnResolutionFailure sets OnResolutionFailureFn callback
	OnResolutionFailure(OnResolutionFailureFn) PlaceholderResolverBuilder

	// Build builds new PlaceholderResolver with all properties set from this builder.
	// It's safe to mutate state of builder and calling Build() again, it won't affect existing resolver instances.
	Build() PlaceholderResolver
}

// DependencyResolver can be used to find dependency errors in document set.
type DependencyResolver interface {
	// Resolve takes each key in srcDoc and attempt to resolve its usage within srcDoc and optionally in zero or more refDoc.
	// On output, report will contain information about every location of inbound references
	// and every keys that has not been referenced at all (aka orpan keys)
	Resolve(srcDoc dom.OverlayDocument, refDoc ...dom.OverlayDocument) *DependencyResolutionReport
}

// DependencyResolverBuilder is fluent builder interface to create DependencyResolver instance
type DependencyResolverBuilder interface {
	// OnPlaceholderEncountered sets callback function that is invoked whenever placeholder is
	// encountered in property value during resolution process
	OnPlaceholderEncountered(func(string, dom.Coordinates)) DependencyResolverBuilder
	// PlaceholderMatcher overrides function that provides dom.SearchValueFunc to check for presence of placeholder in property value
	PlaceholderMatcher(func(string) dom.SearchValueFunc) DependencyResolverBuilder
	// Build creates new instance of DependencyResolver using current state of builder.
	// It's safe to call this method multiple times and/or call other method on this builder that mutate state of builder;
	// every invocation creates new, immutable instance.
	Build() DependencyResolver
}

// DependencyResolutionReport contains result of property dependency resolution
type DependencyResolutionReport struct {
	// Keys of properties that has not been reached from source document (aka orphan properties)
	Orphan []string
	// All keys that where scanned during resolution
	AllKeys []string
	// Keys of properties that were not resolved
	OrphanKeys []string
	// Mapping between any property key and coordinates
	Map map[string]dom.Coordinates
}
