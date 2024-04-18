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

type testDocsData struct {
	doc  string
	tags []string
}

var (
	testDocSrc = map[string]testDocsData{
		"application.yaml": {
			doc: `
---
server:
  port: ${env.port:8080}
apiClient1:
  url: https://${env.domain}/v1
  timeout: ${env.timeout}
  retry: ${env.connection.retryCount}
apiClient2:
  url: https://${env.domain}/v2
`,
			tags: []string{"source"},
		},
		"env-dev.yaml": {
			doc: `
---
env:
  connection:
    retryCount: ${defaults.connection.retryCount}
  domain: dev01.int.example.org
  timeout: ${defaults.connection.timeout}`,
			tags: []string{"env/dev"},
		},
		"env-prod.yaml": {
			doc: `
---
env:
  connection:
    retryCount: 10
  domain: prod02.int.example.org
  timeout: 10s`,
			tags: []string{"env/prod"},
		},
		"env-invalid.yaml": {
			doc: `
---
env:
  domain: ${unresolved.prop.name2}
  timeout: ${unresolved.prop.name1}`,

			tags: []string{"env/invalid"},
		},
		"default.yaml": {
			doc: `
---
defaults:
  connection:
    retryCount: 5
    timeout: 30s`,
			tags: []string{"defaults"},
		},
	}
)
