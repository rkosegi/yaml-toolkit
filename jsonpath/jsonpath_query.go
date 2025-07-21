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

package jsonpath

import (
	"github.com/rkosegi/yaml-toolkit/query"
	jp "github.com/theory/jsonpath"
)

type qryImpl struct {
	p *jp.Path
}

func (q *qryImpl) Select(data any) query.Result {
	return query.Result(q.p.Select(data))
}

type parserImpl struct{}

var qp = &parserImpl{}

func (pi *parserImpl) Parse(in string) (query.Query, error) {
	q := &qryImpl{}
	p, err := jp.Parse(in)
	if err != nil {
		return nil, err
	}
	q.p = p
	return q, nil
}

func NewParser() query.Parser {
	return qp
}
