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

package pipeline

import (
	"fmt"
	"github.com/rkosegi/yaml-toolkit/dom"
	"github.com/rkosegi/yaml-toolkit/utils"
	"os"
	"regexp"
	"strings"
)

// TODO : move these predicates to common place, maybe utils package
var (
	MatchAny = func() StringPredicateFn {
		return func(s string) bool {
			return true
		}
	}
	MatchNone = func() StringPredicateFn {
		return func(s string) bool {
			return false
		}
	}
	MatchRe = func(re *regexp.Regexp) StringPredicateFn {
		return func(s string) bool {
			return re.MatchString(s)
		}
	}
)

// TODO: merge with same thing from analytics package and move it to common place
type StringPredicateFn func(string) bool

func (eo *EnvOp) Do(ctx ActionContext) error {
	var (
		inclFn StringPredicateFn
		exclFn StringPredicateFn
		getter func() []string
	)
	getter = os.Environ
	if eo.envGetter != nil {
		getter = eo.envGetter
	}
	inclFn = MatchAny()
	exclFn = MatchNone()
	if eo.Include != nil {
		inclFn = MatchRe(eo.Include)
	}
	if eo.Exclude != nil {
		exclFn = MatchRe(eo.Exclude)
	}
	for _, env := range getter() {
		parts := strings.SplitN(env, "=", 2)
		if inclFn(parts[0]) && !exclFn(parts[0]) {
			k := utils.ToPath(eo.Path, fmt.Sprintf("Env.%s", parts[0]))
			ctx.Data().AddValueAt(k, dom.LeafNode(parts[1]))
		}
	}
	return nil
}

func (eo *EnvOp) String() string {
	return fmt.Sprintf("Env[path=%s,incl=%s,excl=%s]", eo.Path,
		safeRegexpDeref(eo.Include), safeRegexpDeref(eo.Exclude))
}

func (eo *EnvOp) CloneWith(ctx ActionContext) Action {
	return &EnvOp{
		Include: eo.Include,
		Exclude: eo.Exclude,
		Path:    ctx.TemplateEngine().RenderLenient(eo.Path, ctx.Snapshot()),
	}
}
