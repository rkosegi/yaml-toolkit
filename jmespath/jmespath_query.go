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

package jmespath

import (
	jm "github.com/jmespath/go-jmespath"
	"github.com/rkosegi/yaml-toolkit/query"
)

type qryImpl struct {
	p *jm.JMESPath
}

func (q *qryImpl) Select(data any) query.Result {
	ret, _ := q.p.Search(data)
	switch ret.(type) {
	case []interface{}:
		return ret.([]interface{})

	default:
		return query.Result{ret}
	}
}

type parserImpl struct{}

var qp = &parserImpl{}

func (p *parserImpl) Parse(in string) (query.Query, error) {
	var err error
	q := qryImpl{}
	if q.p, err = jm.Compile(in); err != nil {
		return nil, err
	}
	return &q, nil

}

func NewParser() query.Parser {
	return qp
}
