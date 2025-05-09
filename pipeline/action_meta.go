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
	"strings"
)

// ActionMeta holds action's metadata used by Executor
type ActionMeta struct {
	// Name of this step, should be unique within current scope
	Name string `yaml:"name,omitempty"`

	// Optional ordinal number that controls order of execution within parent step
	Order int `yaml:"order,omitempty"`

	// Optional expression to make execution of this action conditional.
	// Execution of this step is skipped when this expression is evaluated to false.
	// If value of this field is omitted, then this action is executed.
	When *string `yaml:"when,omitempty"`

	// ErrorPropagation configure behavior of error propagation. By default, error is propagated to caller.
	// When set to ErrorPropagationPolicyIgnore, then error is silently ignored.
	ErrorPropagation *ErrorPropagationPolicy `yaml:"errorPropagation,omitempty"`
}

func (am ActionMeta) String() string {
	var (
		sb    strings.Builder
		parts []string
	)
	sb.WriteByte('[')
	if len(am.Name) > 0 {
		parts = append(parts, fmt.Sprintf("name=%s", am.Name))
	}
	if am.Order != 0 {
		parts = append(parts, fmt.Sprintf("order=%d", am.Order))
	}
	when := strings.TrimSpace(safeStrDeref(am.When))
	if len(when) > 0 {
		parts = append(parts, fmt.Sprintf("when=%s", when))
	}
	sb.WriteString(strings.Join(parts, ","))
	sb.WriteByte(']')
	return sb.String()
}
