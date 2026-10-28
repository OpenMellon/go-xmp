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
	"encoding"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

func listPaths(val reflect.Value, path Path) (PathValueList, error) {
	if !val.IsValid() {
		return nil, nil
	}

	if isEmptyValue(val) {
		return nil, nil
	}

	for val.Kind() == reflect.Interface || val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return nil, nil
		}
		val = val.Elem()
	}

	typ := val.Type()
	tinfo, err := getTypeInfo(typ, "xmp")
	if err != nil {
		return nil, err
	}

	pvl := make(PathValueList, 0)

	// walk all fields
	for _, finfo := range tinfo.fields {
		fv := finfo.value(val)

		if !fv.IsValid() {
			continue
		}

		if (fv.Kind() == reflect.Interface || fv.Kind() == reflect.Ptr) && fv.IsNil() {
			continue
		}

		if finfo.flags&fOmit > 0 || (finfo.flags&fEmpty == 0 && isEmptyValue(fv)) {
			continue
		}

		// allow changing the namespace on first-level paths to support
		// multi-namespace models like exif, iptc and quicktime
		if path.Len() == 0 {
			ns := path.NamespacePrefix()
			realNs := getPrefix(finfo.name)
			if ns != realNs {
				path = NewPath(realNs)
			}
		}

		fname := stripPrefix(finfo.name)
		if hasPrefix(finfo.name) && getPrefix(finfo.name) != path.NamespacePrefix() {
			fname = finfo.name
		}

		// Drill into interfaces and pointers.
		for fv.Kind() == reflect.Interface || fv.Kind() == reflect.Ptr {
			fv = fv.Elem()
		}
		typ = fv.Type()

		// handle XMP array types
		av := fv
		isArray := false
		if fv.CanInterface() && (finfo.flags&fArray > 0 || typ.Implements(arrayType)) {
			isArray = true
		} else if fv.CanAddr() {
			pv := fv.Addr()
			if pv.CanInterface() && (finfo.flags&fArray > 0 || pv.Type().Implements(arrayType)) {
				av = pv
				isArray = true
			}
		}

		if isArray {
			switch arr := av.Interface().(type) {
			case ExtensionArray:
				for i, v := range arr {
					subpath := path.Push(fname).AppendIndex(i)
					// name := fmt.Sprintf("%s[%d]", fname, i)
					if v.Model != nil {
						if l, err := listPaths(reflect.ValueOf(v.Model), subpath); err != nil {
							return nil, err
						} else {
							pvl = append(pvl, l...)
						}
					}
					if v.XMLName.Local == "" || (*Node)(v).FullName() == "rdf:Description" {
						for _, child := range v.Nodes {
							if child.Model != nil {
								if l, err := listPaths(reflect.ValueOf(child.Model), subpath); err != nil {
									return nil, err
								} else {
									pvl = append(pvl, l...)
								}
							} else {
								if l, err := child.ListPaths(subpath); err != nil {
									return nil, err
								} else {
									pvl = append(pvl, l...)
								}
							}
						}
					} else {
						node := (*Node)(v)
						if l, err := node.ListPaths(subpath); err != nil {
							return nil, err
						} else {
							pvl = append(pvl, l...)
						}
					}
				}
			case NamedExtensionArray:
				for _, v := range arr {
					node := (*Node)(v)
					if l, err := node.ListPaths(path.Push(fname, node.Name())); err != nil {
						return nil, err
					} else {
						pvl = append(pvl, l...)
					}
				}

			case AltString:
				// AltString types are always at the end of a path
				for _, v := range arr {
					pvl.Add(path.Push(fname).AppendIndexString(v.GetLang()), v.Value)
				}
			default:
				for i, l := 0, av.Len(); i < l; i++ {
					v := derefValue(av.Index(i))

					// check for text marshaler
					isText := false
					vv := v
					if v.CanInterface() && v.Type().Implements(textMarshalerType) {
						isText = true
					} else if v.CanAddr() {
						pv := v.Addr()
						if pv.CanInterface() && pv.Type().Implements(textMarshalerType) {
							vv = pv
							isText = true
						}
					}

					if isText {
						b, err := vv.Interface().(encoding.TextMarshaler).MarshalText()
						if err != nil || b == nil {
							return nil, err
						}
						pvl.Add(path.Push(fname).AppendIndex(i), string(b))
						continue
					}

					switch v.Kind() {
					case reflect.Struct, reflect.Slice, reflect.Array:
						l, err := listPaths(v, path.Push(fname).AppendIndex(i))
						if err != nil {
							return nil, err
						}
						pvl = append(pvl, l...)
					default:
						if s, b, err := marshalSimple(v.Type(), v); err != nil {
							return nil, err
						} else {
							if b != nil {
								s = string(b)
							}
							pvl.Add(path.Push(fname).AppendIndex(i), s)
						}
					}
				}
			}
			continue
		}

		// Check for text marshaler and marshal as string
		av = fv
		isText := false
		if fv.CanInterface() && (finfo.flags&fTextMarshal > 0 || typ.Implements(textMarshalerType)) {
			isText = true
		} else if fv.CanAddr() {
			pv := fv.Addr()
			if pv.CanInterface() && (finfo.flags&fTextMarshal > 0 || pv.Type().Implements(textMarshalerType)) {
				av = pv
				isText = true
			}
		}

		if isText {
			b, err := av.Interface().(encoding.TextMarshaler).MarshalText()
			if err != nil || b == nil {
				return nil, err
			}
			pvl.Add(path.Push(fname), string(b))
			continue
		}

		// continue iteration when field is a struct without text marshaler
		if fv.Kind() == reflect.Struct {
			l, err := listPaths(fv, path.Push(fname))
			if err != nil {
				return nil, err
			}
			pvl = append(pvl, l...)
			continue
		}

		// handle maps
		if fv.Kind() == reflect.Map {
			for _, key := range fv.MapKeys() {
				// need a string representation of key and value here
				val := fv.MapIndex(key)
				ks, kb, kerr := marshalSimple(key.Type(), key)
				if kerr != nil {
					return nil, kerr
				}
				vs, vb, verr := marshalSimple(val.Type(), val)
				if verr != nil {
					return nil, verr
				}
				if kb != nil {
					ks = string(kb)
				}
				if vb != nil {
					vs = string(vb)
				}
				if finfo.flags&fFlat == 0 {
					pvl.Add(path.Push(fname, ks), vs)
				} else {
					pvl.Add(path.Push(ks), vs)
				}
			}
			continue
		}

		// otherwise marshal as value
		if s, b, err := marshalSimple(typ, fv); err != nil {
			return nil, err
		} else {
			if b != nil {
				s = string(b)
			}
			pvl.Add(path.Push(fname), s)
		}
	}

	sort.Sort(byPath(pvl))
	return pvl, nil
}

func parsePathSegment(name string) (string, int, string) {
	var lang string
	var idx int = -1
	// split lang or array index from name, be safe with slice indexes
	if k := strings.Index(name, "["); k > 0 && len(name) > k+1 {
		s := strings.TrimSuffix(name[k+1:], "]")
		if len(s) == 0 {
			idx = 0
		} else if j, err := strconv.Atoi(s); err == nil {
			idx = j
		} else {
			lang = s
		}
		name = name[:k]
	}
	return name, idx, lang
}

func growSlice(v reflect.Value, n int) {
	if v.Kind() != reflect.Slice {
		return
	}
	if l := v.Len(); l <= n {
		if v.Cap() <= n {
			ncap := Max(n*2, 4)
			nv := reflect.MakeSlice(v.Type(), l, ncap)
			reflect.Copy(nv, v)
			v.Set(nv)
		}
		v.SetLen(n + 1)
	}
}
