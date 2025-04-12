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
	"os"
	"regexp"
	"strings"

	"github.com/rkosegi/yaml-toolkit/common"
	"github.com/rkosegi/yaml-toolkit/dom"
	"github.com/rkosegi/yaml-toolkit/utils"
)

// for mock purposes only. this could be used to override os.Environ() to arbitrary func
var (
	envGetter = os.Environ
)

// EnvOp is used to import OS environment variables into data
type EnvOp struct {
	// Optional regexp which defines what to include. Only item names matching this regexp are added into data document.
	Include *regexp.Regexp `yaml:"include,omitempty"`

	// Optional regexp which defines what to exclude. Only item names NOT matching this regexp are added into data document.
	// Exclusion is considered after inclusion regexp is processed.
	Exclude *regexp.Regexp `yaml:"exclude,omitempty"`

	// Optional path within data tree under which "Env" container will be put.
	// When omitted, then "Env" goes to root of data.
	Path string `yaml:"path,omitempty" clone:"template"`
}

func (eo *EnvOp) Do(ctx ActionContext) error {
	var (
		inclFn common.StringPredicateFn
		exclFn common.StringPredicateFn
	)
	inclFn = common.MatchAny()
	exclFn = common.MatchNone()
	if eo.Include != nil {
		inclFn = common.MatchRe(eo.Include)
	}
	if eo.Exclude != nil {
		exclFn = common.MatchRe(eo.Exclude)
	}
	for _, env := range envGetter() {
		parts := strings.SplitN(env, "=", 2)
		if inclFn(parts[0]) && !exclFn(parts[0]) {
			k := utils.ToPath(eo.Path, fmt.Sprintf("Env.%s", parts[0]))
			ctx.Data().AddValueAt(k, dom.LeafNode(parts[1]))
		}
	}
	return nil
}

func (eo *EnvOp) String() string {
	return fmt.Sprintf("Env[Path=%s,incl=%s,excl=%s]", eo.Path,
		safeRegexpDeref(eo.Include), safeRegexpDeref(eo.Exclude))
}

func (eo *EnvOp) CloneWith(ctx ActionContext) Action {
	return &EnvOp{
		Include: eo.Include,
		Exclude: eo.Exclude,
		Path:    ctx.TemplateEngine().RenderLenient(eo.Path, ctx.Snapshot()),
	}
}
