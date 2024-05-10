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
	"errors"
	"fmt"
	"github.com/rkosegi/yaml-toolkit/dom"
	"github.com/rkosegi/yaml-toolkit/k8s"
	"github.com/rkosegi/yaml-toolkit/props"
	"github.com/rkosegi/yaml-toolkit/utils"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

const (
	wildcardTag = "*"
)

var (
	ErrLayerAlreadyExists = errors.New("layer already exists")
)

var (
	b           = dom.Builder()
	defaultOpts = []AddLayerOpt{
		// give every document '*' tag by default, so it can be matched with user-supplied value
		WithTags(wildcardTag),
	}
)

type documentSet struct {
	ctxMap         map[string]*docContext
	unnamedLayerId int
}

func (ds *documentSet) NamedDocument(name string) dom.ContainerBuilder {
	ctx := ds.ctxMap[name]
	if ctx == nil {
		return nil
	}
	return ctx.doc
}

// data structure to associate arbitrary data to document within the set
type docContext struct {
	// actual document
	doc dom.ContainerBuilder
	// list of tags
	tags []string
	// function used to merge 2 contexts
	mergeFn func(ctx *docContext, doc dom.ContainerBuilder) error
}

// DefaultFileDecoderProvider is FileDecoderProvider that uses file suffix to choose dom.DecoderFunc
func DefaultFileDecoderProvider(file string) dom.DecoderFunc {
	switch filepath.Ext(file) {
	case ".yaml", ".yml":
		return dom.DefaultYamlDecoder
	case ".json":
		return dom.DefaultJsonDecoder
	case ".properties":
		return props.DecoderFn
	default:
		return nil
	}
}

func WithTags(tag ...string) AddLayerOpt {
	return func(ds *documentSet, docName string, ctx *docContext) {
		ctx.tags = append(ctx.tags, tag...)
	}
}

// MergeTags does not consider newly added dom.ContainerBuilder, it just merges tags into existing context
func MergeTags() AddLayerOpt {
	return func(_ *documentSet, _ string, context *docContext) {
		context.mergeFn = func(newCtx *docContext, doc dom.ContainerBuilder) error {
			context.tags = utils.Unique(append(context.tags, newCtx.tags...))
			if context.doc == nil {
				context.doc = newCtx.doc
			}
			return nil
		}
	}
}

// MustCreate ensures that layer does not already exist
func MustCreate() AddLayerOpt {
	return func(_ *documentSet, _ string, context *docContext) {
		context.mergeFn = func(newCtx *docContext, doc dom.ContainerBuilder) error {
			return ErrLayerAlreadyExists
		}
	}
}

func (ds *documentSet) applyOpts(name string, ctx *docContext, opts ...AddLayerOpt) *docContext {
	for _, opt := range defaultOpts {
		opt(ds, name, ctx)
	}
	for _, opt := range opts {
		opt(ds, name, ctx)
	}
	return ctx
}

func (ds *documentSet) addContext(name string, doc dom.ContainerBuilder, newCtx *docContext) error {
	existingCtx, exists := ds.ctxMap[name]
	if exists {
		if newCtx.mergeFn != nil {
			err := newCtx.mergeFn(existingCtx, doc)
			if err != nil {
				return err
			}
		}
		ds.ctxMap[name] = newCtx
		return nil
	} else {
		newCtx.doc = doc
		ds.ctxMap[name] = newCtx
		return nil
	}
}

func (ds *documentSet) newContext(name string, opts ...AddLayerOpt) *docContext {
	return ds.applyOpts(name, &docContext{}, opts...)
}

func (ds *documentSet) AddDocument(name string, doc dom.ContainerBuilder, opts ...AddLayerOpt) error {
	return ds.addContext(name, doc, ds.newContext(name, opts...))
}

func (ds *documentSet) AddUnnamedDocument(doc dom.ContainerBuilder, opts ...AddLayerOpt) error {
	ds.unnamedLayerId++
	return ds.AddDocument(fmt.Sprintf("default__%d", ds.unnamedLayerId), doc, opts...)
}

func (ds *documentSet) AddDocumentFromReader(name string, r io.Reader, dec dom.DecoderFunc, opts ...AddLayerOpt) error {
	cb, err := b.FromReader(r, dec)
	if err != nil {
		return err
	}
	return ds.AddDocument(name, cb, opts...)
}

func (ds *documentSet) AddDocumentFromFile(file string, dec dom.DecoderFunc, opts ...AddLayerOpt) error {
	f, err := utils.FileOpener(file)
	if err != nil {
		return err
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)
	return ds.AddDocumentFromReader(file, f, dec, opts...)
}

func (ds *documentSet) AddDocumentsFromDirectory(pattern string, decProv FileDecoderProvider, opts ...AddLayerOpt) error {
	files, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}
	for _, file := range files {
		err = ds.AddDocumentFromFile(file, decProv(file), opts...)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ds *documentSet) AddDocumentsFromManifest(manifest string, decProv FileDecoderProvider, opts ...AddLayerOpt) error {
	m, err := k8s.ManifestFromFile(manifest)
	if err != nil {
		return err
	}
	d := m.StringData()
	for _, item := range d.List() {
		// only error that can be returned by Reader (from strings.NewReader) is EOF under some circumstances, so
		// ignoring err seems to be safe here. I'm happy to be proven wrong, however.
		_ = ds.AddDocumentFromReader(fmt.Sprintf("%s/%s", manifest, item),
			strings.NewReader(*d.Get(item)), decProv(item), opts...)
	}
	return nil
}

func (ds *documentSet) AddPropertiesFromManifest(manifest string, opts ...AddLayerOpt) error {
	doc, err := k8s.Properties(manifest)
	if err != nil {
		return err
	} else {
		return ds.AddDocument(manifest, doc.Document(), opts...)
	}
}

func containsAnyOf(col []string, contains []string) bool {
	for _, i := range col {
		if slices.Contains(contains, i) {
			return true
		}
	}
	return false
}

func (ds *documentSet) TaggedSubset(tag ...string) dom.OverlayDocument {
	return ds.filtered(func(_ string, ctx *docContext) bool {
		return containsAnyOf(ctx.tags, tag)
	})
}

func (ds *documentSet) AsOne() dom.OverlayDocument {
	return ds.filtered(func(string, *docContext) bool {
		return true
	})
}

func (ds *documentSet) filtered(filterFn func(name string, ctx *docContext) bool) dom.OverlayDocument {
	o := dom.NewOverlayDocument()
	for n, c := range ds.ctxMap {
		if filterFn(n, c) {
			o.Add(n, c.doc)
		}
	}
	return o
}

// NewDocumentSet creates new instance of documentSet with all fields initialized to default values.
func NewDocumentSet() DocumentSet {
	return &documentSet{
		ctxMap: make(map[string]*docContext),
	}
}
