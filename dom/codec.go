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

package dom

import (
	"reflect"

	"github.com/rkosegi/yaml-toolkit/query"
	"gopkg.in/yaml.v3"
)

func encodeLeafFn(n Leaf) interface{} {
	return n.Value()
}

func encodeListFn(n List) []interface{} {
	res := make([]interface{}, n.Size())
	for i, item := range n.Items() {
		if item.IsContainer() {
			res[i] = encodeContainerFn(item.AsContainer())
		} else if item.IsList() {
			res[i] = encodeListFn(item.AsList())
		} else {
			res[i] = encodeLeafFn(item.AsLeaf())
		}
	}
	return res
}

func encodeContainerFn(n Container) map[string]interface{} {
	res := map[string]interface{}{}
	for k, v := range n.Children() {
		if v.IsContainer() {
			res[k] = encodeContainerFn(v.AsContainer())
		} else if v.IsList() {
			res[k] = encodeListFn(v.AsList())
		} else {
			res[k] = encodeLeafFn(v.AsLeaf())
		}
	}
	return res
}

func decodeLeafFn(v interface{}) Leaf {
	return LeafNode(v)
}

func decodeListFn(v []interface{}, l ListBuilder) {
	for _, item := range v {
		t := reflect.ValueOf(item)
		switch t.Kind() {
		case reflect.Map:
			l.Append(DefaultNodeDecoderFn(item.(map[string]interface{})))
		case reflect.Slice, reflect.Array:
			list := initListBuilder()
			decodeListFn(item.([]interface{}), list)
			l.Append(list)
		case reflect.Float32, reflect.Float64, reflect.String, reflect.Bool,
			reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			l.Append(decodeLeafFn(item))
		}
	}
}

func decodeContainerFn(current *map[string]interface{}, parent ContainerBuilder) {
	for k, v := range *current {
		if v == nil {
			parent.AddValue(k, nilLeaf)
		} else {
			t := reflect.ValueOf(v)
			switch t.Kind() {
			case reflect.Map:
				ref := v.(map[string]interface{})
				decodeContainerFn(&ref, parent.AddContainer(k))
			case reflect.Slice, reflect.Array:
				decodeListFn(v.([]interface{}), parent.AddList(k))
			case reflect.Float32, reflect.Float64, reflect.String, reflect.Bool,
				reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
				reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				parent.AddValue(k, decodeLeafFn(v))
			}
		}
	}
}

func decodeYamlNode(n *yaml.Node) Node {
	switch n.Kind {
	case yaml.ScalarNode:
		// TODO: leaf is always string unless more elaborate approach is taken, but good for now
		// see e.g. gopkg.in/yaml.v3@v3.0.1/decode.go:565
		return LeafNode(n.Value)
	case yaml.SequenceNode:
		lb := ListNode()
		for _, x := range n.Content {
			lb.Append(decodeYamlNode(x))
		}
		return lb
	case yaml.DocumentNode:
		if len(n.Content) == 1 {
			return decodeYamlNode(n.Content[0])
		}
	case yaml.MappingNode:
		cb := ContainerNode()
		l := len(n.Content)
		for i := 0; i < l; i += 2 {
			cb.AddValue(n.Content[i].Value, decodeYamlNode(n.Content[i+1]))
		}
		return cb
	}
	return nil
}

func decodeValueToNode(in reflect.Value) Node {
	switch in.Kind() {
	case reflect.Pointer, reflect.Interface:
		return decodeValueToNode(in.Elem())

	case reflect.Struct:
		out := ContainerNode()
		it := in.Type()
		for i := 0; i < it.NumField(); i++ {
			if f := it.Field(i); f.IsExported() {
				out.AddValue(f.Name, decodeValueToNode(in.Field(i)))
			}
		}
		return out

	case reflect.Slice, reflect.Array:
		out := ListNode()
		for i := range in.Len() {
			out.Set(uint(i), decodeValueToNode(in.Index(i)))
		}
		return out

	case reflect.Float32, reflect.Float64, reflect.String, reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return LeafNode(in.Interface())

	case reflect.Map:
		out := ContainerNode()
		for _, k := range in.MapKeys() {
			v := in.MapIndex(k)
			if !v.IsZero() {
				out.AddValue(k.String(), decodeValueToNode(v))
			}
		}
		return out
	}
	return nil
}

func decodeQueryResult(qr query.Result) NodeList {
	qrl := len(qr)
	if qrl == 0 {
		return nil
	}
	res := make([]Node, qrl)
	for i, item := range qr {
		res[i] = DecodeAnyToNode(item)
	}
	return res
}
