/*
Copyright 2025 Richard Kosegi

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

package template_engine

type TemplateEngine interface {
	Render(template string, data map[string]interface{}) (string, error)
	// RenderLenient attempts to render given template using provided data, while swallowing any error.
	// Value of template is first checked by simple means if it is actually template to avoid unnecessary errors.
	// Use with caution.
	RenderLenient(template string, data map[string]interface{}) string
	// RenderSliceLenient attempts to render slices of template strings using provided data, without reporting any error.
	RenderSliceLenient(templates []string, data map[string]interface{}) []string
	// RenderMapLenient attempts to render every leaf value in provided map
	RenderMapLenient(input map[string]interface{}, data map[string]interface{}) map[string]interface{}
	// EvalBool attempts to evaluate template as boolean value
	EvalBool(template string, data map[string]interface{}) (bool, error)
}
