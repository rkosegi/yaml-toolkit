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
)

func (na ChildActions) Do(ctx ActionContext) error {
	for _, name := range sortActionNames(na) {
		err := ctx.Executor().Execute(na[name])
		if err != nil {
			return err
		}
	}
	return nil
}

func (na ChildActions) CloneWith(ctx ActionContext) Action {
	r := make(ChildActions)
	for k, v := range na {
		r[k] = v.CloneWith(ctx).(ActionSpec)
	}
	return r
}

func (na ChildActions) String() string {
	return fmt.Sprintf("ChildActions[names=%s]", actionNames(na))
}
