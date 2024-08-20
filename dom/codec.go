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

import "reflect"

func encodeLeafFn(n Leaf) interface{} {
	return n.Value()
}

func encoderListFn(n List) []interface{} {
	res := make([]interface{}, n.Size())
	for i, item := range n.Items() {
		if item.IsContainer() {
			res[i] = encodeContainerFn(item.(Container))
		} else if item.IsList() {
			res[i] = encoderListFn(item.(List))
		} else {
			res[i] = encodeLeafFn(item.(Leaf))
		}
	}
	return res
}

func encodeContainerFn(n Container) map[string]interface{} {
	res := map[string]interface{}{}
	for k, v := range n.(Container).Children() {
		if v.IsContainer() {
			res[k] = encodeContainerFn(v.(Container))
		} else if v.IsList() {
			res[k] = encoderListFn(v.(List))
		} else {
			res[k] = encodeLeafFn(v.(Leaf))
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
			list := &listBuilderImpl{}
			decodeListFn(item.([]interface{}), list)
			l.Append(list)
		case reflect.Int, reflect.Float64, reflect.String, reflect.Bool:
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
			case reflect.Int, reflect.Float64, reflect.String, reflect.Bool:
				parent.AddValue(k, decodeLeafFn(v))
			}
		}
	}
}
