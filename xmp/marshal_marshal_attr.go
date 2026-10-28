// Copyright (c) 2017-2018 Alexander Eichhorn
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package xmp

import (
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"reflect"
	"strconv"
)

func (e *Encoder) marshalAttr(node *Node, name xml.Name, val reflect.Value) error {
	attr, ok, err := e.attrValue(node, name, val)
	if err != nil || !ok {
		return err
	}
	node.AddAttr(attr)
	return nil
}

func marshalSimple(typ reflect.Type, val reflect.Value) (string, []byte, error) {
	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(val.Int(), 10), nil, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return strconv.FormatUint(val.Uint(), 10), nil, nil
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(val.Float(), 'g', -1, val.Type().Bits()), nil, nil
	case reflect.String:
		return val.String(), nil, nil
	case reflect.Bool:
		return strconv.FormatBool(val.Bool()), nil, nil
	case reflect.Array:
		if typ.Elem().Kind() != reflect.Uint8 {
			break
		}
		// [...]byte
		var bytes []byte
		if val.CanAddr() {
			bytes = val.Slice(0, val.Len()).Bytes()
		} else {
			bytes = make([]byte, val.Len())
			reflect.Copy(reflect.ValueOf(bytes), val)
		}
		return "", bytes, nil
	case reflect.Slice:
		if typ.Elem().Kind() != reflect.Uint8 {
			break
		}
		// []byte
		return "", val.Bytes(), nil
	}
	return "", nil, fmt.Errorf("xmp: no method for marshalling type %s (%v)\n", typ.String(), val.Kind())
}

func (e Encoder) _findNsByURI(uri string) *Namespace {
	if v, ok := e.intNsMap[uri]; ok {
		return v
	}
	if v, ok := e.extNsMap[uri]; ok {
		return v
	}
	return nil
}

func (e Encoder) _findNsByPrefix(pre string) *Namespace {
	for _, v := range e.intNsMap {
		if v.GetName() == pre {
			return v
		}
	}
	for _, v := range e.extNsMap {
		if v.GetName() == pre {
			return v
		}
	}
	if ns, err := NsRegistry.GetNamespace(pre); err == nil {
		return ns
	}
	return nil
}

func (e Encoder) findNs(n xml.Name) *Namespace {
	var ns *Namespace
	if n.Space != "" {
		ns = e._findNsByURI(n.Space)
	}
	if ns == nil {
		ns = e._findNsByPrefix(getPrefix(n.Local))
	}
	return ns
}

func ToString(t any) string {
	val := reflect.Indirect(reflect.ValueOf(t))
	if !val.IsValid() {
		return ""
	}
	if val.Type().Implements(stringerType) {
		return t.(fmt.Stringer).String()
	}
	if s, err := ToRawString(val.Interface()); err == nil {
		return s
	}
	return fmt.Sprintf("%v", val.Interface())
}

func isBase64(s string) bool {
	_, err := base64.StdEncoding.DecodeString(s)
	return err == nil
}

func ToRawString(t any) (string, error) {
	val := reflect.Indirect(reflect.ValueOf(t))
	typ := val.Type()
	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(val.Int(), 10), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return strconv.FormatUint(val.Uint(), 10), nil
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(val.Float(), 'g', -1, val.Type().Bits()), nil
	case reflect.String:
		return val.String(), nil
	case reflect.Bool:
		return strconv.FormatBool(val.Bool()), nil
	case reflect.Array:
		if typ.Elem().Kind() != reflect.Uint8 {
			break
		}
		// [...]byte
		var b []byte
		if val.CanAddr() {
			b = val.Slice(0, val.Len()).Bytes()
		} else {
			b = make([]byte, val.Len())
			reflect.Copy(reflect.ValueOf(b), val)
		}
		if !isBase64(string(b)) {
			return base64.StdEncoding.EncodeToString(b), nil
		}
		return string(b), nil
	case reflect.Slice:
		if typ.Elem().Kind() != reflect.Uint8 {
			break
		}
		// []byte
		b := val.Bytes()
		if !isBase64(string(b)) {
			return base64.StdEncoding.EncodeToString(b), nil
		}
		return string(b), nil
	}
	return "", fmt.Errorf("xmp: no method for converting type %s (%v) to string", typ.String(), val.Kind())
}
