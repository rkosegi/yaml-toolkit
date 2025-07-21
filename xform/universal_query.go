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

package xform

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/rkosegi/yaml-toolkit/jmespath"
	"github.com/rkosegi/yaml-toolkit/jsonpath"
	"github.com/rkosegi/yaml-toolkit/query"
	"gopkg.in/yaml.v3"
)

type QuerySyntax string

const (
	// QuerySyntaxJsonpath is RFC9535 (aka jsonpath) query
	QuerySyntaxJsonpath = QuerySyntax("rfc9535")
	// QuerySyntaxJmespath is using syntax described in https://jmespath.org/specification.html
	QuerySyntaxJmespath = QuerySyntax("jmespath")
)

var queryParsersMap = map[QuerySyntax]query.Parser{
	QuerySyntaxJsonpath: jsonpath.NewParser(),
	QuerySyntaxJmespath: jmespath.NewParser(),
}

type UniversalQuery struct {
	Value  query.Query
	Syntax QuerySyntax
}

func (u *UniversalQuery) UnmarshalYAML(node *yaml.Node) error {
	var (
		val string
		syn interface{}
		ok  bool
		x   interface{}
		pf  query.Parser
	)
	u.Syntax = QuerySyntaxJsonpath

	switch node.Kind {
	case yaml.ScalarNode:
		val = node.Value

	case yaml.MappingNode:
		m := make(map[string]interface{})
		// TODO: how to provoke error from this call?
		_ = node.Decode(&m)
		if x, ok = m["value"]; !ok {
			return errors.New("missing 'value' field under query")
		}
		if val, ok = x.(string); !ok {
			return fmt.Errorf("'value' field is not a string (actual type: %v)", reflect.TypeOf(x))
		}
		if syn, ok = m["syntax"]; ok {
			u.Syntax = QuerySyntax(syn.(string))
		}
	default:
		return fmt.Errorf("node kind is not supported: %v", node.Kind)
	}

	if pf, ok = queryParsersMap[u.Syntax]; ok {
		p, err := pf.Parse(val)
		if err != nil {
			return err
		}
		u.Value = p
	} else {
		return fmt.Errorf("unrecognized query syntax: %s", u.Syntax)
	}
	return nil
}
