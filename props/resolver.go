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

package props

import (
	"slices"
	"strings"
)

const notFound = -1

type propImpl struct {
	b   *builder
	pl  int
	sl  int
	vsl int
}

func removeFromSlice(slice []string, toRemove string) []string {
	if idx := slices.Index(slice, toRemove); idx != notFound {
		ret := make([]string, 0)
		ret = append(ret, slice[:idx]...)
		return append(ret, slice[idx+1:]...)
	}
	return slice
}

func indexAfter(str, sub string, offset int) int {
	if offset > len(str) {
		return notFound
	}
	if ret := strings.Index(str[offset:], sub); ret != notFound {
		return ret + offset
	}
	return notFound
}

func replaceAt(in string, start, end int, replacement string) string {
	ret := in[0:start]
	ret += replacement
	if end < len(in) {
		ret += in[end:]
	}
	return ret
}

func matchAt(str string, index int, substring string) bool {
	if index+len(substring) > len(str) {
		return false
	}
	for i := 0; i < len(substring); i++ {
		if str[index+i] != substring[i] {
			return false
		}
	}
	return true
}

func (p *propImpl) findEndIndex(buf string, startIndex int) int {
	index := startIndex + p.pl
	nested := 0
	for index < len(buf) {
		if matchAt(buf, index, p.b.suffix) {
			if nested > 0 {
				nested--
				index += p.sl
			} else {
				return index
			}
		} else if matchAt(buf, index, p.b.prefix) {
			nested++
			index += p.pl
		} else {
			index++
		}
	}
	return notFound
}

func (p *propImpl) resolve(value string, lookupFn LookupFn, seen []string) *string {
	si := strings.Index(value, p.b.prefix)
	if si == notFound {
		return &value
	}
	result := value
	for si != notFound {
		ei := p.findEndIndex(result, si)
		if ei != notFound {
			ph := result[si+p.pl : ei]
			orig := ph
			if slices.Contains(seen, orig) {
				panic("Circular placeholder reference '" + orig + "' in property definitions")
			}
			seen = append(seen, orig)
			ph = *p.resolve(ph, lookupFn, seen)
			pv := p.resolvePlaceholder(lookupFn, ph)
			if pv != nil {
				pv = p.resolve(*pv, lookupFn, seen)
				result = replaceAt(result, si, ei+p.sl, *pv)
				si = indexAfter(result, p.b.prefix, si+len(*pv))
			} else {
				si = indexAfter(result, p.b.prefix, ei+p.sl)
			}
			removeFromSlice(seen, orig)
		} else {
			si = notFound
		}
	}
	return &result
}

func (p *propImpl) resolvePlaceholder(lookupFn LookupFn, ph string) *string {
	pv := lookupFn(ph)
	if pv == nil {
		sep := strings.Index(ph, p.b.vs)
		if sep != notFound {
			actualPlaceholder := ph[0:sep]
			def := ph[sep+p.vsl:]
			pv = lookupFn(actualPlaceholder)
			if pv == nil {
				pv = &def
			}
		}
	}
	return pv
}

func (p *propImpl) Resolve(placeholder string) string {
	return *p.resolve(placeholder, p.b.lookupFunc, []string{})
}

type builder struct {
	prefix     string
	suffix     string
	vs         string
	lookupFunc LookupFn
}

func (b *builder) Prefix(prefix string) ResolverBuilder {
	b.prefix = prefix
	return b
}

func (b *builder) Suffix(suffix string) ResolverBuilder {
	b.suffix = suffix
	return b
}

func (b *builder) ValueSeparator(vs string) ResolverBuilder {
	b.vs = vs
	return b
}

func (b *builder) LookupFunc(fn LookupFn) ResolverBuilder {
	b.lookupFunc = fn
	return b
}

func (b *builder) MustBuild() Resolver {
	if b.lookupFunc == nil {
		panic("lookup function not configured")
	}
	return &propImpl{
		b:   b,
		pl:  len(b.prefix),
		sl:  len(b.suffix),
		vsl: len(b.vs),
	}
}

// Builder returns new ResolverBuilder with defaults set.
func Builder() ResolverBuilder {
	return &builder{
		prefix: "${",
		suffix: "}",
		vs:     ":",
	}
}

// MapLookup returns simple LookupFn that looks into provided map
func MapLookup(data map[string]string) LookupFn {
	return func(key string) *string {
		if r, ok := data[key]; ok {
			return &r
		}
		return nil
	}
}

var (
	_ ResolverBuilder = &builder{}
	_ Resolver        = &propImpl{}
)
