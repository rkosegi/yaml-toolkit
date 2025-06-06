# YAML toolkit

[![codecov](https://codecov.io/gh/rkosegi/yaml-toolkit/graph/badge.svg?token=BX0P2QQPR2)](https://codecov.io/gh/rkosegi/yaml-toolkit)
[![Go Report Card](https://goreportcard.com/badge/github.com/rkosegi/yaml-toolkit)](https://goreportcard.com/report/github.com/rkosegi/yaml-toolkit)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=rkosegi_yaml-toolkit&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=rkosegi_yaml-toolkit)
[![Technical Debt](https://sonarcloud.io/api/project_badges/measure?project=rkosegi_yaml-toolkit&metric=sqale_index)](https://sonarcloud.io/summary/new_code?id=rkosegi_yaml-toolkit)
[![Vulnerabilities](https://sonarcloud.io/api/project_badges/measure?project=rkosegi_yaml-toolkit&metric=vulnerabilities)](https://sonarcloud.io/summary/new_code?id=rkosegi_yaml-toolkit)
[![Maintainability Rating](https://sonarcloud.io/api/project_badges/measure?project=rkosegi_yaml-toolkit&metric=sqale_rating)](https://sonarcloud.io/summary/new_code?id=rkosegi_yaml-toolkit)
[![Security Rating](https://sonarcloud.io/api/project_badges/measure?project=rkosegi_yaml-toolkit&metric=security_rating)](https://sonarcloud.io/summary/new_code?id=rkosegi_yaml-toolkit)
[![Go Reference](https://pkg.go.dev/badge/github.com/rkosegi/yaml-toolkit.svg)](https://pkg.go.dev/github.com/rkosegi/yaml-toolkit)
[![Apache 2.0 License](https://badgen.net/static/license/Apache2.0/blue)](https://github.com/rkosegi/yaml-toolkit/blob/main/LICENSE)
[![CodeQL Status](https://github.com/rkosegi/yaml-toolkit/actions/workflows/codeql.yaml/badge.svg)](https://github.com/rkosegi/yaml-toolkit/security/code-scanning)
[![CI Status](https://github.com/rkosegi/yaml-toolkit/actions/workflows/ci.yaml/badge.svg)](https://github.com/rkosegi/yaml-toolkit/actions/workflows/ci.yaml)
[![OpenSSF Scorecard](https://api.scorecard.dev/projects/github.com/rkosegi/yaml-toolkit/badge)](https://scorecard.dev/viewer/?uri=github.com/rkosegi/yaml-toolkit)

Go library to deal with (not only) YAML documents.


## Usage

### Loading configuration file with defaults

given configuration file

`config.yaml`
```yaml
---
listen_address: 0.0.0.0:8080
read_timeout: 15s
```

```go
package main

import (
	"time"

	"github.com/rkosegi/yaml-toolkit/dom"
	"github.com/rkosegi/yaml-toolkit/fluent"
)

type Config struct {
	Address     string        `yaml:"listen_address"`
	ReadTimeout time.Duration `yaml:"read_timeout"`
}

var defConfig = &Config{
	Address: "0.0.0.0:8081",
}

func main() {
	cfg := fluent.NewConfigHelper[Config]().
		Add(defConfig).         // add defaults
		Load("config.yaml").    // override with values loaded from YAML document
		Result()                // get merged results

	// ... use cfg
}
```

### Opening embedded YAML

`example.yaml`
```yaml
---
kind: ConfigMap
metadata:
  name: cm1
  namespace: default
apiVersion: v1
data:
  application.yaml: |
    xyz: 456
    abc:
      def:
        leaf1: 123
        leaf2: Hello
```

code
```go
package main

import (
    ydom "github.com/rkosegi/yaml-toolkit/dom"
    yk8s "github.com/rkosegi/yaml-toolkit/k8s"
)

func main() {
	d, err := yk8s.YamlDoc("example.yaml", "application.yaml")
	if err != nil {
		panic(err)
	}
	print (d.Document().Child("xyz").(ydom.Leaf).Value()) // 456
	d.Document().AddValue("another-key", ydom.LeafNode("789")) // add new child node
	err = d.Save()
	if err != nil {
		panic(err)
	}
}
```

### Compute difference between 2 documents

<table>
<thead><tr><th>Left</th><th>Right</th></tr></thead>
<tbody>
<tr><td><code>left.yaml</code></td><td><code>right.yaml</code></td></tr>
<tr><td>

```yaml
---
root:
  sub1:
    leaf1: abc
    leaf2: 123
  list:
    - 1
    - 2
```
</td><td>

```yaml
---
root:
  sub1:
    leaf1: def
    leaf3: 789
  list:
    - 3
```
</td>
</tr>
</tbody>
</table>

```go
package main

import (
	"fmt"
	yta "github.com/rkosegi/yaml-toolkit/analytics"
	ytc "github.com/rkosegi/yaml-toolkit/common"
	ydiff "github.com/rkosegi/yaml-toolkit/diff"
)

func main() {
	dp := ytc.DefaultFileDecoderProvider(".yaml")
	ds := yta.NewDocumentSet()
	err := ds.AddDocumentFromFile("left.yaml", dp, yta.WithTags("left"))
	if err != nil {
		panic(err)
	}
	err = ds.AddDocumentFromFile("right.yaml", dp, yta.WithTags("right"))
	if err != nil {
		panic(err)
	}
	changes := ydiff.Diff(
		ds.TaggedSubset("right").Merged(),
		ds.TaggedSubset("left").Merged(),
	)
	fmt.Printf("All changes: %d\n", len(*changes))
	for _, change := range *changes {
		fmt.Printf("%s: %s => %v\n", change.Type, change.Path, change.Value)
	}
}
```

output:
```
All changes: 5
Delete: root.list => <nil>
Add: root.list[0] => 3
Change: root.sub1.leaf1 => abc
Delete: root.sub1.leaf2 => <nil>
Add: root.sub1.leaf3 => 789

```
