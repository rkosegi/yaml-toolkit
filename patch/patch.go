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

package patch

import (
	"errors"
	"fmt"
	"github.com/rkosegi/yaml-toolkit/dom"
)

type Op string

var (
	ErrNoOo           = errors.New("operations object is nil")
	ErrNoTarget       = errors.New("target is nil")
	ErrOoPathMissing  = errors.New("'path' field is missing in operations object")
	ErrOoValueMissing = errors.New("'value' field is missing in operations object")
	ErrOoFromMissing  = errors.New("'from' field is missing in operations object")
)

const (
	OpAdd     = Op("add")
	OpRemove  = Op("remove")
	OpReplace = Op("replace")
	OpMove    = Op("move")
	OpCopy    = Op("copy")
	OpTest    = Op("test")
)

type opFn func(obj *OpObj, target dom.ContainerBuilder) error

// OpObj is operation object containing all necessary data to perform patch operation
type OpObj struct {
	Op    Op
	From  *Path    // copy,move
	Path  Path     // test,copy,move,replace,remove,add
	Value dom.Node // test,replace,add
}

// Do perform patch operation against designated target.
// Object is modified in place, or error is returned.
func Do(obj *OpObj, target dom.ContainerBuilder) error {
	if obj == nil {
		return ErrNoOo
	}
	if obj.Path == nil {
		return ErrOoPathMissing
	}
	if target == nil {
		return ErrNoTarget
	}
	switch obj.Op {
	// https://datatracker.ietf.org/doc/html/rfc6902#section-4.1
	case OpAdd:
		return doAdd(obj, target)

	// https://datatracker.ietf.org/doc/html/rfc6902#section-4.2
	case OpRemove:
		return doRemove(obj, target)

	// https://datatracker.ietf.org/doc/html/rfc6902#section-4.3
	case OpReplace:
		return doReplace(obj, target)

	// https://datatracker.ietf.org/doc/html/rfc6902#section-4.4
	case OpMove:
		return doMove(obj, target)

	// https://datatracker.ietf.org/doc/html/rfc6902#section-4.5
	case OpCopy:
		return doCopy(obj, target)

	// https://datatracker.ietf.org/doc/html/rfc6902#section-4.6
	case OpTest:
		return doTest(obj, target)
	}
	return fmt.Errorf("invalid operation: %v", obj.Op)
}

func doAdd(obj *OpObj, target dom.ContainerBuilder) error {
	if obj.Value == nil {
		return ErrOoValueMissing
	}
	_, parent := obj.Path.Parent().Eval(target)
	if parent == nil {
		return fmt.Errorf("parent path does not resolve to existing node: %s", obj.Path.Parent().String())
	}
	// If the target location specifies an array index, a new value is
	//      inserted into the array at the specified index.
	if idx, isNum := obj.Path.LastSegment().IsNumeric(); isNum && parent.IsList() {
		insertListItem(parent.(dom.ListBuilder), idx, obj.Value)
		return nil
	} else {
		// If the target location specifies an object member that does not
		//      already exist, a new member is added to the object.
		// If the target location specifies an object member that does exist,
		//      that member's value is replaced.
		parent.(dom.ContainerBuilder).AddValue(string(obj.Path.LastSegment()), obj.Value)
	}
	return nil
}

func doRemove(obj *OpObj, target dom.ContainerBuilder) error {
	_, n := obj.Path.Eval(target)
	// The target location MUST exist for the operation to be successful.
	if n == nil {
		return fmt.Errorf("node at path %s does not exists", obj.Path.String())
	}
	_, parent := obj.Path.Parent().Eval(target)
	if idx, isNum := obj.Path.LastSegment().IsNumeric(); isNum && parent.IsList() {
		// If removing an element from an array, any elements above the
		//		specified index are shifted one position to the left.
		removeListItem(parent.(dom.ListBuilder), idx)
	} else {
		parent.(dom.ContainerBuilder).Remove(string(obj.Path.LastSegment()))
	}
	return nil
}

func doReplace(obj *OpObj, target dom.ContainerBuilder) error {
	if obj.Value == nil {
		return ErrOoValueMissing
	}
	nl, n := obj.Path.Eval(target)
	// The target location MUST exist for the operation to be successful.
	if n == nil {
		return fmt.Errorf("path does not resolve to existing node: %s", obj.Path.String())
	}
	parent := nl[len(nl)-2]
	if idx, isNum := obj.Path.LastSegment().IsNumeric(); isNum && parent.IsList() {
		parent.(dom.ListBuilder).Set(uint(idx), obj.Value)
	} else {
		parent.(dom.ContainerBuilder).AddValue(string(obj.Path.LastSegment()), obj.Value)
	}
	return nil
}

func doMove(obj *OpObj, target dom.ContainerBuilder) error {
	// This operation is functionally identical to a "remove" operation on
	//   the "from" location, followed immediately by an "add" operation at
	//   the target location with the value that was just removed.
	return moveOrCopy(obj, target, true)
}

func doCopy(obj *OpObj, target dom.ContainerBuilder) error {
	// This operation is functionally identical to an "add" operation at the
	//   target location using the value specified in the "from" member.
	return moveOrCopy(obj, target, false)
}

func moveOrCopy(obj *OpObj, target dom.ContainerBuilder, move bool) error {
	if obj.From == nil {
		return ErrOoFromMissing
	}
	// The "from" location MUST exist for the operation to be successful.
	n, err := get(*obj.From, target)
	if err != nil {
		return err
	}
	if move {
		_ = doRemove(&OpObj{
			Path: *obj.From,
		}, target)
	}

	return doAdd(&OpObj{Value: n, Path: obj.Path}, target)
}

func doTest(obj *OpObj, target dom.ContainerBuilder) error {
	if obj.Value == nil {
		return ErrOoValueMissing
	}
	if n, err := get(obj.Path, target); err != nil {
		return err
	} else if !obj.Value.Equals(n) {
		return fmt.Errorf("node at path %s with value %v does not match expected value of %v", obj.Path, n, obj.Value)
	}
	return nil
}

func get(path Path, target dom.ContainerBuilder) (dom.Node, error) {
	_, n := path.Eval(target)
	if n == nil {
		return nil, fmt.Errorf("path does not resolve to existing node: %s", path.String())
	}
	return n, nil
}
