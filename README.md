# YAML toolkit

[![Go Report Card](https://goreportcard.com/badge/github.com/rkosegi/yaml-toolkit)](https://goreportcard.com/report/github.com/rkosegi/yaml-toolkit)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=rkosegi_yaml-toolkit&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=rkosegi_yaml-toolkit)
[![Technical Debt](https://sonarcloud.io/api/project_badges/measure?project=rkosegi_yaml-toolkit&metric=sqale_index)](https://sonarcloud.io/summary/new_code?id=rkosegi_yaml-toolkit)
[![Vulnerabilities](https://sonarcloud.io/api/project_badges/measure?project=rkosegi_yaml-toolkit&metric=vulnerabilities)](https://sonarcloud.io/summary/new_code?id=rkosegi_yaml-toolkit)
[![Maintainability Rating](https://sonarcloud.io/api/project_badges/measure?project=rkosegi_yaml-toolkit&metric=sqale_rating)](https://sonarcloud.io/summary/new_code?id=rkosegi_yaml-toolkit)
[![Security Rating](https://sonarcloud.io/api/project_badges/measure?project=rkosegi_yaml-toolkit&metric=security_rating)](https://sonarcloud.io/summary/new_code?id=rkosegi_yaml-toolkit)
[![Coverage](https://sonarcloud.io/api/project_badges/measure?project=rkosegi_yaml-toolkit&metric=coverage)](https://sonarcloud.io/summary/new_code?id=rkosegi_yaml-toolkit)

Go library to deal with YAML documents embedded within k8s manifests (like Spring boot's application.yaml).


## Usage

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
