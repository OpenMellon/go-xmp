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
	"fmt"
	"reflect"
	"sort"
)

func SetModelPath(v Model, path Path, value string, flags SyncFlags) error {
	if flags == 0 {
		flags = DEFAULT
	}
	if !path.IsXmpPath() {
		return fmt.Errorf("xmp: invalid path '%s'", path.String())
	}

	val := derefIndirect(v)

	l := path.Len()
	for n, walker := path.PopFront(); n != ""; n, walker = walker.PopFront() {
		name, idx, lang := parsePathSegment(n)
		if idx < -1 {
			return fmt.Errorf("path field %d (%s): invalid index", l-walker.Len(), n)
		}

		// using the short-form of name here to find attribute names across namespaces
		// (e.g. some models use different namespace structs internally)
		finfo, err := findField(val, name, "xmp")
		if err != nil {
			// special error is catched by caller to retry setting as raw node
			return errNotFound
		}

		// allocate memory for pointer values in structs
		fv := finfo.value(val)
		if !fv.IsValid() {
			return nil
		}
		if fv.Type().Kind() == reflect.Ptr && fv.IsNil() && fv.CanSet() {
			fv.Set(reflect.New(fv.Type().Elem()))
		}
		fv = derefValue(fv)

		// continue loop when field is a struct and we're not at the end
		if fv.Kind() == reflect.Struct && walker.Len() > 0 {
			val = fv
			continue
		}

		// handle maps
		if fv.Kind() == reflect.Map {
			// use proper name depending on flattening
			if finfo.flags&fFlat == 0 {
				n, walker = walker.PopFront()
				name, _, _ = parsePathSegment(n)
			}

			t := fv.Type()
			if fv.IsNil() {
				fv.Set(reflect.MakeMap(t))
			}
			switch t.Key().Kind() {
			case reflect.String:
			default:
				return fmt.Errorf("map key type must be string")
			}

			switch {
			case flags&DELETE > 0 && value == "":
				// remove map value
				fv.SetMapIndex(reflect.ValueOf(name), reflect.Zero(t.Elem()))
				return nil
			case flags&(REPLACE|CREATE) > 0 && value != "":
				// set map value
				switch t.Elem().Kind() {
				case reflect.String:
					fv.SetMapIndex(reflect.ValueOf(name), reflect.ValueOf(value).Convert(t.Elem()))
				case reflect.Struct:
					val = reflect.New(t.Elem()).Elem()
					// FIXME: this recursion may not work as expected because the map
					// value is not updated when set later on
					fv.SetMapIndex(reflect.ValueOf(name), val)
					continue
				default:
					// FIXME: this does not allow for struct pointers as map values
					mval := reflect.New(t.Elem()).Elem()
					if mval.Type().Kind() == reflect.Ptr && mval.IsNil() && mval.CanSet() {
						mval.Set(reflect.New(mval.Type().Elem()))
					}
					if mval.CanInterface() && mval.Type().Implements(textUnmarshalerType) {
						if err := mval.Interface().(encoding.TextUnmarshaler).UnmarshalText([]byte(value)); err != nil {
							return err
						}
					} else if mval.CanAddr() {
						pv := mval.Addr()
						if pv.CanInterface() && pv.Type().Implements(textUnmarshalerType) {
							if err := pv.Interface().(encoding.TextUnmarshaler).UnmarshalText([]byte(value)); err != nil {
								return err
							}
						}
					} else {
						if err := setValue(mval, value); err != nil {
							return err
						}
					}
					fv.SetMapIndex(reflect.ValueOf(name), mval)
					return nil
				}
			}
			continue
		}

		// handle Slice/Array type fields as pointers to make them settable
		av := fv
		if fv.CanAddr() {
			av = fv.Addr()
		}
		if av.CanInterface() && (finfo != nil && finfo.flags&fArray > 0 || av.Type().Implements(arrayType)) {
			switch arr := av.Interface().(type) {
			case *ExtensionArray:
				// unwrap extension namespace and keep path remainder
				name = walker.PeekNamespacePrefix()

				// sanitize idx (it is -1 when no array index was specified)
				if idx < 0 {
					idx = 0
				}

				// add empty nodes up to idx if necessary
				if l := len(*arr); l <= idx {
					if flags&(CREATE|APPEND) == 0 && value != "" {
						return fmt.Errorf("CREATE flag required to add extension %s on path %s", name, path)
					}
					// grow slice and fill with initialized nodes
					for ; l <= idx; l++ {
						*arr = append(*arr, (*Extension)(NewNode(EmptyName)))
					}
				}

				// use the node at index
				ext := (*arr)[idx]
				node := (*Node)(ext)

				// there is two types of extensions
				if node.Name() != "" || node.FullName() == "rdf:Description" {
					// Type 1: model and nodes on top-level (name not empty)
					node = ext.FindNodeByName(name)
					if node == nil {
						if ns, err := GetNamespace(name); err == nil {
							node.Model = ns.NewModel()
						}
					}
					if node.Model != nil {
						// FIXME: when the model does not contain the path we might
						// want to store it as child node, however, the for{} loop
						// allows no back-tracking on error
						val = reflect.Indirect(reflect.ValueOf(node.Model))
						continue
					}
					// store as child node
					return node.SetPath(walker, value, flags)
				} else {
					// Type 2: models on child level (empty name); used for xmpMM:Pantry
					child := node.Nodes.FindNodeByName(name)
					if child == nil {
						child = node.AddNode(NewNode(NewName(name)))
					}
					if child.Model == nil {
						if ns, err := GetNamespace(name); err == nil {
							child.Model = ns.NewModel()
						}
					}
					if child.Model != nil {
						// FIXME: when the model does not contain the path we might
						// want to store it as extension model child node. however,
						// the for{} loop allows no back-tracking on error
						val = reflect.Indirect(reflect.ValueOf(child.Model))
						continue
					}
					// store as child node
					return child.SetPath(NewPath(name, walker.Fields()...), value, flags)
				}

			case *NamedExtensionArray:
				// Named extensions do not contain models
				// unwrap extension namespace and keep path remainder
				name, walker = walker.PopFront()
				node := arr.FindNodeByName(name)
				if node == nil {
					// create new extension node without model
					if flags&(CREATE|APPEND) == 0 && value != "" {
						return fmt.Errorf("CREATE flag required to add extension %s on path %s", name, path)
					}
					node = NewNode(NewName(name))
					ext := (*Extension)(node)
					fv.Set(reflect.Append(fv, reflect.ValueOf(ext)))
				}
				// store as child node
				return node.SetPath(walker, value, flags)

			case *AltString:
				switch {
				case flags&UNIQUE > 0 && value != "":
					// append source when not exist
					if !arr.AddUnique(lang, value) {
						return fmt.Errorf("equal value exists for %s on path %s", name, path)
					}
				case flags&APPEND > 0 && value != "":
					// append source value
					arr.Add(lang, value)
				case flags&(REPLACE|CREATE) > 0 && value != "":
					// replace entire AltString with a new version
					if lang != "" {
						arr.Set(lang, value)
					} else {
						*arr = NewAltString(value)
					}
				case flags&DELETE > 0 && value == "":
					// delete the entire AltString or just a specific language
					if lang != "" {
						arr.RemoveLang(lang)
					} else {
						fv.Set(reflect.Zero(fv.Type()))
					}
				}
				return nil

			default:
				// deref the pointer
				av = reflect.Indirect(av)
				// are we at the end of the path?
				if walker.Len() > 0 {
					// sanitize idx (it is -1 when no array index was specified)
					if idx < 0 {
						idx = 0
					}
					// handle arrays along the path
					if av.Len() <= idx {
						if flags&DELETE > 0 && value == "" {
							// would be deleting smth inside a non-existent element
							return nil
						}
						if flags&(CREATE|APPEND) == 0 {
							return fmt.Errorf("CREATE flag required to grow slice %s to index %d", name, idx)
						}
						if av.Kind() == reflect.Array {
							return fmt.Errorf("array %s index %d out of bounds", name, idx)
						}

						// add empty items up to idx
						// FIXME: when an error occurs later on, we cannot remove slice entry
						growSlice(fv, idx)
						val = reflect.New(fv.Type().Elem())
						fv.Index(idx).Set(val.Elem())
						val = derefValue(fv.Index(idx))

						// unmarshal text
						if val.CanInterface() && val.Type().Implements(textUnmarshalerType) {
							return val.Interface().(encoding.TextUnmarshaler).UnmarshalText([]byte(value))
						}

						// recurse into struct
						if reflect.Indirect(val).Kind() == reflect.Struct {
							continue
						}

						// when reaching this, path is too long
						return errNotFound
					}

					// recurse to next path segment using the value at slice index
					val = derefValue(fv.Index(idx))
					continue
				}

				// arrays are fixed size
				if av.Kind() == reflect.Array {
					if flags&REPLACE == 0 && value != "" {
						return fmt.Errorf("REPLACE flag required for setting array value in %s", name)
					}
					if flags&DELETE == 0 && value == "" {
						return fmt.Errorf("DELETE flag required to clear array value in %s", name)
					}

					// sanitize idx (it is -1 when no array index was specified)
					if idx < 0 {
						idx = 0
					}
					if av.Len() <= idx {
						return fmt.Errorf("array %s index %d out of bounds", name, idx)
					}

					v := reflect.New(fv.Type().Elem())
					if v.CanInterface() && v.Type().Implements(textUnmarshalerType) {
						if err := v.Interface().(encoding.TextUnmarshaler).UnmarshalText([]byte(value)); err != nil {
							return err
						}
					} else if err := setValue(v, value); err != nil {
						return err
					}
					av.Index(idx).Set(v.Elem())
					return nil
				}

				switch {
				// case flags&UNIQUE > 0 && value != "" && idx > -1:
				// ignore the unique flag when setting with absolute slice index
				case flags&UNIQUE > 0 && value != "" && idx == -1:
					// append if unique
					for i, l := 0, fv.Len(); i < l; i++ {
						v := derefValue(fv.Index(i))

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
								return err
							}
							if value == string(b) {
								return fmt.Errorf("equal value exists for %s on path %s", name, path)
							}
						} else {
							if s, b, err := marshalSimple(v.Type(), v); err != nil {
								return err
							} else {
								if b != nil {
									s = string(b)
								}
								if value == s {
									return fmt.Errorf("equal value exists for %s on path %s", name, path)
								}
							}
						}
					}

					v := reflect.New(fv.Type().Elem())
					if v.CanInterface() && v.Type().Implements(textUnmarshalerType) {
						if err := v.Interface().(encoding.TextUnmarshaler).UnmarshalText([]byte(value)); err != nil {
							return err
						}
					} else if err := setValue(v, value); err != nil {
						return err
					}

					fv.Set(reflect.Append(fv, v.Elem()))

				case flags&APPEND > 0 && value != "" && idx == -1:
					// always append
					v := reflect.New(fv.Type().Elem())
					if v.CanInterface() && v.Type().Implements(textUnmarshalerType) {
						if err := v.Interface().(encoding.TextUnmarshaler).UnmarshalText([]byte(value)); err != nil {
							return err
						}
					} else if err := setValue(v, value); err != nil {
						return err
					}

					fv.Set(reflect.Append(fv, v.Elem()))

				case flags&(REPLACE|CREATE) > 0 && value != "" && idx > -1:
					// replace or create slice element at index
					v := reflect.New(fv.Type().Elem())
					if v.CanInterface() && v.Type().Implements(textUnmarshalerType) {
						if err := v.Interface().(encoding.TextUnmarshaler).UnmarshalText([]byte(value)); err != nil {
							return err
						}
					} else if err := setValue(v, value); err != nil {
						return err
					}
					// optionally growing the slice
					growSlice(fv, idx)
					// overwrite current entry,
					fv.Index(idx).Set(v.Elem())

				case flags&(REPLACE|CREATE) > 0 && value != "" && idx == -1:
					// replace entire slice with a single element
					v := reflect.New(fv.Type().Elem())
					if v.CanInterface() && v.Type().Implements(textUnmarshalerType) {
						if err := v.Interface().(encoding.TextUnmarshaler).UnmarshalText([]byte(value)); err != nil {
							return err
						}
					} else if err := setValue(v, value); err != nil {
						return err
					}
					fv.Set(reflect.Zero(fv.Type()))
					fv.Set(reflect.Append(fv, v.Elem()))

				case flags&DELETE > 0 && value == "" && idx > -1:
					// delete slice element at index
					l := fv.Len()
					if l <= idx {
						return fmt.Errorf("slice %s index %d out of bounds", name, idx)
					}
					switch {
					case idx == 0:
						fv.Set(fv.Slice(1, l))
					case idx == l-1:
						fv.Set(fv.Slice(0, idx))
					default:
						fv.Set(reflect.AppendSlice(fv.Slice3(0, idx, l-1), fv.Slice(idx+1, l)))
					}

				case flags&DELETE > 0 && value == "" && idx == -1:
					// delete the entire slice
					fv.Set(reflect.Zero(fv.Type()))
				default:
					return fmt.Errorf("unsupported flag combination %v", flags)
				}
				return nil
			}
		}

		// for simple values, check current value before replacing
		if flags&REPLACE == 0 && value != "" && !isEmptyValue(fv) {
			return fmt.Errorf("REPLACE flag required for overwriting value at %s on path %s", name, path)
		}

		// Text
		if fv.CanAddr() {
			pv := fv.Addr()
			if pv.CanInterface() && (finfo != nil && finfo.flags&fTextUnmarshal > 0 || pv.Type().Implements(textUnmarshalerType)) {
				return pv.Interface().(encoding.TextUnmarshaler).UnmarshalText([]byte(value))
			}
		}

		// otherwise set simple value directly
		return setValue(fv, value)
	}
	return nil
}

func (d *Document) ListPaths() (PathValueList, error) {
	// sync individual models to establish correct XMP entries
	if err := d.syncToXMP(); err != nil {
		return nil, err
	}
	l := make(PathValueList, 0)
	for _, v := range d.nodes {
		if v.Model != nil {
			if pvl, err := ListModelPaths(v.Model); err != nil {
				return nil, err
			} else {
				l = append(l, pvl...)
			}
		}
		r, err := v.ListPaths(NewPath(v.Name()))
		if err != nil {
			return nil, err
		}
		l = append(l, r...)
	}
	sort.Sort(byPath(l))
	return l.Unique(), nil
}

func ListModelPaths(v Model) (PathValueList, error) {
	return listPaths(reflect.ValueOf(v), NewPath(v.Namespaces()[0].GetName()))
}
