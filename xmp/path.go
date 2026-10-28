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
	"strings"
)

type Path string

type PathValue struct {
	Path      Path      `json:"path"`
	Namespace string    `json:"namespace,omitempty"`
	Value     string    `json:"value"`
	Flags     SyncFlags `json:"flags,omitempty"`
}

type PathValueList []PathValue

func (x *PathValueList) Add(p Path, v string) {
	if v == "" {
		return
	}
	*x = append(*x, PathValue{Path: p, Value: v})
}

func (x *PathValueList) AddFlags(p Path, v string, f SyncFlags) {
	if v == "" {
		return
	}
	*x = append(*x, PathValue{Path: p, Value: v, Flags: f})
}

func (x PathValueList) Find(p Path) *PathValue {
	for _, v := range x {
		if v.Path == p {
			return &v
		}
	}
	return nil
}

// assumes a sorted list
func (x PathValueList) Unique() PathValueList {
	l := make(PathValueList, 0, len(x))
	var last Path
	for _, v := range x {
		if last != v.Path {
			l = append(l, v)
			last = v.Path
		}
	}
	return l
}

func (x PathValueList) Diff(y PathValueList) PathValueList {
	if len(x) == 0 {
		return y
	}
	if len(y) == 0 {
		return x
	}
	diff := make(PathValueList, 0, Max(len(x), len(y)))
	for _, xv := range x {
		yv := y.Find(xv.Path)
		if yv == nil {
			diff.AddFlags(xv.Path, xv.Value, DELETE)
			continue
		}
		if yv.Value != xv.Value {
			diff.AddFlags(xv.Path, xv.Value, REPLACE)
		}
	}
	for _, yv := range y {
		xv := x.Find(yv.Path)
		if xv == nil {
			diff.AddFlags(yv.Path, yv.Value, CREATE)
		}
	}
	return diff
}

type byPath PathValueList

func (l byPath) Len() int           { return len(l) }
func (l byPath) Swap(i, j int)      { l[i], l[j] = l[j], l[i] }
func (l byPath) Less(i, j int) bool { return l[i].Path.String() < l[j].Path.String() }

type byValue PathValueList

func (l byValue) Len() int           { return len(l) }
func (l byValue) Swap(i, j int)      { l[i], l[j] = l[j], l[i] }
func (l byValue) Less(i, j int) bool { return l[i].Value < l[j].Value }

func (x Path) String() string {
	return string(x)
}

func NewPath(prefix string, segments ...string) Path {
	return Path(prefix + ":" + strings.Join(segments, "/"))
}

// ns:tagname/subtagname
func (x Path) IsXmpPath() bool {
	return strings.Index(string(x), ":") > -1
}

func (x Path) NamespacePrefix() string {
	if i := strings.Index(string(x), ":"); i > -1 {
		return string(x[:i])
	}
	return string(x)
}

func (x Path) Namespace(d *Document) (*Namespace, error) {
	if i := strings.Index(string(x), ":"); i > -1 {
		if ns := d.findNsByPrefix(string(x[:i])); ns != nil {
			return ns, nil
		}
	}
	return nil, fmt.Errorf("xmp: invalid path '%s'", x.String())
}

func (x Path) MatchNamespace(ns *Namespace) bool {
	if ns == nil {
		return false
	}
	if i := strings.Index(string(x), ":"); i > -1 {
		return ns.GetName() == string(x[:i])
	}
	return false
}

func (x Path) Len() int {
	s := string(x)
	if i := strings.Index(s, ":"); i > -1 {
		s = strings.TrimPrefix(s[i+1:], "/")
		if len(s) > 0 {
			return strings.Count(s, "/") + 1
		}
	}
	return 0
}

func (x Path) Fields() []string {
	s := string(x)
	if _, after, ok := strings.Cut(s, ":"); ok {
		y := strings.TrimPrefix(after, "/")
		if len(y) > 0 {
			return strings.Split(y, "/")
		}
	}
	return nil
}

func (x Path) AppendIndex(i int) Path {
	f := "%s[%d]"
	if strings.HasSuffix(x.String(), "]") {
		f = "%s/[%d]"
	}
	return Path(fmt.Sprintf(f, x.String(), i))
}

func (x Path) AppendIndexString(s string) Path {
	f := "%s[%s]"
	if strings.HasSuffix(x.String(), "]") {
		f = "%s/[%s]"
	}
	return Path(fmt.Sprintf(f, x.String(), s))
}

func (x Path) Push(segment ...string) Path {
	f := x.Fields()
	f = append(f, segment...)
	return NewPath(x.NamespacePrefix(), f...)
}

func (x Path) Pop() (string, Path) {
	f := x.Fields()
	switch l := len(f); l {
	case 0:
		return "", x
	case 1:
		return f[0], NewPath(x.NamespacePrefix())
	default:
		return f[l-1], NewPath(x.NamespacePrefix(), f[:l-1]...)
	}
}

func (x Path) PopFront() (string, Path) {
	switch f := x.Fields(); len(f) {
	case 0:
		return "", x
	case 1:
		seg := f[0]
		ns := x.NamespacePrefix()
		if hasPrefix(seg) {
			ns = getPrefix(seg)
			seg = stripPrefix(seg)
		}
		return seg, NewPath(ns)
	default:
		seg := f[0]
		ns := x.NamespacePrefix()
		if hasPrefix(seg) {
			ns = getPrefix(seg)
			seg = stripPrefix(seg)
		}
		return seg, NewPath(ns, f[1:]...)
	}
}

func (x Path) PeekNamespacePrefix() string {
	switch f := x.Fields(); len(f) {
	case 0:
		return x.NamespacePrefix()
	default:
		if hasPrefix(f[0]) {
			return getPrefix(f[0])
		}
		return x.NamespacePrefix()
	}
}

func (x *Path) UnmarshalText(data []byte) error {
	p := Path(string(data))
	if !p.IsXmpPath() {
		return fmt.Errorf("xmp: invalid path '%s'", p.String())
	}
	*x = p
	return nil
}

func (d *Document) GetPath(path Path) (string, error) {
	if !path.IsXmpPath() {
		return "", fmt.Errorf("xmp: invalid path '%s'", path.String())
	}
	ns, err := path.Namespace(d)
	if err != nil {
		return "", err
	}
	n := d.FindNode(ns)
	if n == nil {
		return "", fmt.Errorf("xmp: path '%s' not found", path.String())
	}
	if n.Model != nil {
		if s, err := GetModelPath(n.Model, path); err == nil {
			return s, nil
		} else if err != errNotFound {
			return "", fmt.Errorf("xmp: path '%s' error: %v", path.String(), err)
		}
	}
	if v, err := n.GetPath(path); err != nil {
		if err == errNotFound {
			return "", fmt.Errorf("xmp: path '%s' not found", path.String())
		} else {
			return "", fmt.Errorf("xmp: path '%s' error: %v", path.String(), err)
		}
	} else {
		return v, nil
	}
}

func GetModelPath(v Model, path Path) (string, error) {
	val := derefIndirect(v)
	l := path.Len()
	for n, walker := path.PopFront(); n != ""; n, walker = walker.PopFront() {
		// fmt.Printf("%d name=%s walker=%s\n", l-walker.Len(), n, walker.String())
		name, idx, lang := parsePathSegment(n)
		if idx < -1 {
			return "", fmt.Errorf("path field %d (%s): invalid index", l-walker.Len(), n)
		}

		finfo, err := findField(val, name, "xmp")
		if err != nil {
			return "", errNotFound
			// return "", fmt.Errorf("path field %d (%s) not found: %v", i, name, err)
		}

		fv := finfo.value(val)
		typ := fv.Type()

		// ignore empty fields
		if !fv.IsValid() {
			return "", nil
		}
		if (fv.Kind() == reflect.Interface || fv.Kind() == reflect.Ptr) && fv.IsNil() {
			return "", nil
		}
		if finfo.flags&fOmit > 0 || (finfo.flags&fEmpty == 0 && isEmptyValue(fv)) {
			return "", nil
		}

		// Drill into interfaces and pointers.
		for fv.Kind() == reflect.Interface || fv.Kind() == reflect.Ptr {
			fv = fv.Elem()
		}

		// continue loop when field is a struct and we're not at the end yet
		if fv.Kind() == reflect.Struct && walker.Len() > 0 {
			val = fv
			continue
		}

		// handle XMP array types
		av := fv
		isArray := false
		if fv.CanInterface() && (finfo != nil && finfo.flags&fArray > 0 || typ.Implements(arrayType)) {
			isArray = true
		} else if fv.CanAddr() {
			pv := fv.Addr()
			if pv.CanInterface() && (finfo != nil && finfo.flags&fArray > 0 || pv.Type().Implements(arrayType)) {
				av = pv
				isArray = true
			}
		}

		if isArray {
			switch arr := av.Interface().(type) {
			case ExtensionArray:
				if walker.Len() == 0 || fv.Len() <= idx {
					return "", errNotFound
				}
				ext := arr[idx]
				name = walker.PeekNamespacePrefix()
				node := ext.FindNodeByName(name)
				if node == nil {
					return "", errNotFound
				}
				if node.Model != nil {
					val = reflect.Indirect(reflect.ValueOf(node.Model))
					continue
				}
				return node.GetPath(walker)

			case NamedExtensionArray:
				node := arr.FindNodeByName(name)
				if node == nil {
					return "", errNotFound
				}
				return node.GetPath(walker)

			case AltString:
				if lang != "" {
					return arr.Get(lang), nil
				}
				return arr.Default(), nil
			default:
				// sanitize array index
				if idx < 0 {
					idx = 0
				}
				if fv.Len() < idx {
					return "", errNotFound
				}
				if walker.Len() > 0 {
					if av.Len() <= idx {
						return "", errNotFound
					}
					val = derefValue(av.Index(idx))
					continue
				} else {
					if av.Len() <= idx {
						return "", errNotFound
					}
					fv = derefValue(av.Index(idx))
					typ = fv.Type()
					finfo = nil
				}
			}
		}

		// Check for text marshaler and marshal as string
		av = fv
		isText := false
		if fv.CanInterface() && (finfo != nil && finfo.flags&fTextMarshal > 0 || typ.Implements(textMarshalerType)) {
			isText = true
		} else if fv.CanAddr() {
			pv := fv.Addr()
			if pv.CanInterface() && (finfo != nil && finfo.flags&fTextMarshal > 0 || pv.Type().Implements(textMarshalerType)) {
				av = pv
				isText = true
			}
		}

		if isText {
			b, err := av.Interface().(encoding.TextMarshaler).MarshalText()
			if err != nil || b == nil {
				return "", err
			}
			return string(b), nil
		}

		// handle maps
		if fv.Kind() == reflect.Map {
			if finfo.flags&fFlat == 0 {
				n, walker = walker.PopFront()
				name, _, _ = parsePathSegment(n)
			}
			val := fv.MapIndex(reflect.ValueOf(name))
			if !val.IsValid() {
				return "", errNotFound
			}
			// process as simple value below
			fv = val
			typ = val.Type()
		}

		// simple values are just fine, but any other type (slice, array, struct)
		// without textmarshaler will fail here
		if s, b, err := marshalSimple(typ, fv); err != nil {
			return "", err
		} else {
			if b != nil {
				s = string(b)
			}
			return s, nil
		}
	}
	return "", nil
}

func (d *Document) SetPath(desc PathValue) error {
	flags := desc.Flags
	path := desc.Path
	value := desc.Value

	if flags == 0 {
		flags = DEFAULT
	}
	if !path.IsXmpPath() {
		return fmt.Errorf("xmp: invalid path '%s'", path.String())
	}

	ns, err := path.Namespace(d)
	if ns == nil || err != nil {
		if desc.Namespace != "" {
			ns = &Namespace{path.NamespacePrefix(), desc.Namespace, nil}
			Register(ns)
		} else {
			return err
		}
	}

	m := d.FindModel(ns)
	n := d.FindNode(ns)
	if m == nil && n == nil && flags&CREATE == 0 {
		if flags&NOFAIL > 0 {
			return nil
		}
		return fmt.Errorf("xmp: create flag required to make model for '%s'", path.String())
	}
	if m == nil {
		m = ns.NewModel()
		if m != nil {
			n, _ = d.AddModel(m)
		}
	}
	if m == nil && n == nil {
		n = d.nodes.AddNode(NewNode(ns.RootName()))
	}

	// empty source will only be used with delete flag
	if value == "" && flags&DELETE == 0 {
		if flags&NOFAIL > 0 {
			return nil
		}
		return fmt.Errorf("xmp: delete flag required for empty '%s'", path.String())
	}

	// delete model when only namespace is set
	if path.Len() == 0 && flags&DELETE > 0 {
		d.RemoveNamespace(ns)
	}

	// get current version of path value
	var dest string
	if m != nil {
		dest, err = GetModelPath(m, path)
	}
	if m == nil || err == errNotFound {
		dest, err = n.GetPath(path)
	}
	if err != nil {
		return err
	}

	// skip when equal
	if dest == value {
		return nil
	}

	// empty destination values require create flag
	if dest == "" && flags&(CREATE|APPEND|UNIQUE) == 0 {
		if flags&NOFAIL > 0 {
			return nil
		}
		return fmt.Errorf("xmp: create flag required to make new attribute at '%s'", path.String())
	}

	// existing destination values require replace/delete/append/unique flag
	if dest != "" && flags&(REPLACE|DELETE|APPEND|UNIQUE) == 0 {
		if flags&NOFAIL > 0 {
			return nil
		}
		return fmt.Errorf("xmp: update flag required to change existing attribute at '%s'", path.String())
	}

	if m != nil {
		if err = SetModelPath(m, path, value, flags); err != nil && err == errNotFound {
			err = n.SetPath(path, value, flags)
		}
	} else {
		err = n.SetPath(path, value, flags)
	}
	if err == nil {
		d.SetDirty()
	}

	if flags&NOFAIL > 0 {
		return nil
	}
	return err
}
