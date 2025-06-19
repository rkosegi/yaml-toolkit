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

// Package analytics provides visibility into documentSet.
// Currently following reports are implemented:
//
//	1, placeholder resolution - attempts to resolve every property placeholder within documentSet.
//	2, dependency resolution - counts references to every property in documentSet
//  3, deduplication - find properties that are common in given documentSet
//
// Examples.
//
// given files in common_test.go
//
// Now image scenarios, where application loads and merge configs in following order:
// a) 1, application.yaml  2, defaults.yaml 3, env-dev.yaml
// b) 1, application.yaml  2, defaults.yaml 3, env-prod.yaml
// c) 1, application.yaml  2, defaults.yaml 3, env-invalid.yaml
//
// In both a) and b) cases document set can be resolved,
// however c) case will be left with 2 unresolved keys: "unresolved.prop.name1" and "unresolved.prop.name2"
//

package analytics
