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

package main

import (
	"testing"

	_ "github.com/OpenMellon/go-xmp/models"
	"github.com/OpenMellon/go-xmp/models/dc"
	"github.com/OpenMellon/go-xmp/xmp"
)

func TestPathDeleteSlice(T *testing.T) {
	d := xmp.NewDocument()
	c := &dc.DublinCore{
		Type: xmp.NewStringArray("one", "two", "three"),
	}
	d.AddModel(c)
	if err := d.SetPath(xmp.PathValue{
		Path:  xmp.Path("dc:type"),
		Value: "",
		Flags: xmp.DELETE,
	}); err != nil {
		T.Errorf("unexpected error when deleting existing path: %v", err)
	}
	if len(c.Type) != 0 {
		T.Errorf("unexpected non-empty array after delete")
	}
	if v := dc.FindModel(d); v == nil {
		T.Errorf("unexpected empty model after delete")
	}
}

func TestPathDeleteAltString(T *testing.T) {
	d := xmp.NewDocument()
	c := &dc.DublinCore{
		Description: xmp.NewAltString(
			xmp.AltItem{Value: "german", Lang: "de", IsDefault: false},
			xmp.AltItem{Value: "english", Lang: "en", IsDefault: true},
		),
	}
	d.AddModel(c)
	if err := d.SetPath(xmp.PathValue{
		Path:  xmp.Path("dc:description"),
		Value: "",
		Flags: xmp.DELETE,
	}); err != nil {
		T.Errorf("unexpected error when deleting existing path: %v", err)
	}
	if len(c.Description) != 0 {
		T.Errorf("unexpected non-empty array after delete")
	}
	if v := dc.FindModel(d); v == nil {
		T.Errorf("unexpected empty model after delete")
	}
}

func TestPathDeleteFromSlice1(T *testing.T) {
	d := xmp.NewDocument()
	c := &dc.DublinCore{
		Type: xmp.NewStringArray("one", "two", "three"),
	}
	d.AddModel(c)
	if err := d.SetPath(xmp.PathValue{
		Path:  xmp.Path("dc:type[0]"),
		Value: "",
		Flags: xmp.DELETE,
	}); err != nil {
		T.Errorf("unexpected error when deleting existing slice element: %v", err)
	}
	if v := dc.FindModel(d); v == nil {
		T.Errorf("unexpected empty model after delete")
	}
	if l := len(c.Type); l != 2 {
		T.Errorf("unexpected slice length after delete: expected=2 got=%d", l)
		return
	}
	if s := c.Type[0]; s != "two" {
		T.Errorf("unexpected first slice element after delete: expected=two got=%s", s)
	}
	if s := c.Type[1]; s != "three" {
		T.Errorf("unexpected last slice element after delete: expected=three got=%s", s)
	}
}

func TestPathDeleteFromSlice2(T *testing.T) {
	d := xmp.NewDocument()
	c := &dc.DublinCore{
		Type: xmp.NewStringArray("one", "two", "three"),
	}
	d.AddModel(c)
	if err := d.SetPath(xmp.PathValue{
		Path:  xmp.Path("dc:type[1]"),
		Value: "",
		Flags: xmp.DELETE,
	}); err != nil {
		T.Errorf("unexpected error when deleting existing slice element: %v", err)
	}
	if v := dc.FindModel(d); v == nil {
		T.Errorf("unexpected empty model after delete")
	}
	if l := len(c.Type); l != 2 {
		T.Errorf("unexpected slice length after delete: expected=2 got=%d", l)
		return
	}
	if s := c.Type[0]; s != "one" {
		T.Errorf("unexpected first slice element after delete: expected=one got=%s", s)
	}
	if s := c.Type[1]; s != "three" {
		T.Errorf("unexpected last slice element after delete: expected=three got=%s", s)
	}
}

func TestPathDeleteFromSlice3(T *testing.T) {
	d := xmp.NewDocument()
	c := &dc.DublinCore{
		Type: xmp.NewStringArray("one", "two", "three"),
	}
	d.AddModel(c)
	if err := d.SetPath(xmp.PathValue{
		Path:  xmp.Path("dc:type[2]"),
		Value: "",
		Flags: xmp.DELETE,
	}); err != nil {
		T.Errorf("unexpected error when deleting existing slice element: %v", err)
	}
	if v := dc.FindModel(d); v == nil {
		T.Errorf("unexpected empty model after delete")
	}
	if l := len(c.Type); l != 2 {
		T.Errorf("unexpected slice length after delete: expected=2 got=%d", l)
		return
	}
	if s := c.Type[0]; s != "one" {
		T.Errorf("unexpected first slice element after delete: expected=one got=%s", s)
	}
	if s := c.Type[1]; s != "two" {
		T.Errorf("unexpected last slice element after delete: expected=two got=%s", s)
	}
}

func TestPathDeleteFromSliceNotExist(T *testing.T) {
	d := xmp.NewDocument()
	c := &dc.DublinCore{
		Type: xmp.NewStringArray("one", "two", "three"),
	}
	d.AddModel(c)
	if err := d.SetPath(xmp.PathValue{
		Path:  xmp.Path("dc:type[3]"),
		Value: "",
		Flags: xmp.DELETE,
	}); err != nil {
		T.Errorf("unexpected error when deleting non-existing slice element: %v", err)
	}
	if v := dc.FindModel(d); v == nil {
		T.Errorf("unexpected empty model after delete")
	}
	if l := len(c.Type); l != 3 {
		T.Errorf("unexpected slice length after delete: expected=3 got=%d", l)
		return
	}
	if s := c.Type[0]; s != "one" {
		T.Errorf("unexpected first slice element after delete: expected=one got=%s", s)
	}
	if s := c.Type[1]; s != "two" {
		T.Errorf("unexpected first slice element after delete: expected=two got=%s", s)
	}
	if s := c.Type[2]; s != "three" {
		T.Errorf("unexpected last slice element after delete: expected=three got=%s", s)
	}
}

func TestPathDeleteFromAltString(T *testing.T) {
	d := xmp.NewDocument()
	c := &dc.DublinCore{
		Description: xmp.NewAltString(
			xmp.AltItem{Value: "german", Lang: "de", IsDefault: false},
			xmp.AltItem{Value: "english", Lang: "en", IsDefault: true},
		),
	}
	d.AddModel(c)
	if err := d.SetPath(xmp.PathValue{
		Path:  xmp.Path("dc:description[de]"),
		Value: "",
		Flags: xmp.DELETE,
	}); err != nil {
		T.Errorf("unexpected error when deleting existing path: %v", err)
	}
	if len(c.Description) != 1 {
		T.Errorf("unexpected non-empty AltString after delete")
	}
	if v := dc.FindModel(d); v == nil {
		T.Errorf("unexpected empty model after delete")
	}
	if s := c.Description.Get("de"); s != "" {
		T.Errorf("invalid result: expected=<empty> got=%s", s)
	}
	if s := c.Description.Get("en"); s != "english" {
		T.Errorf("invalid result: expected=english got=%s", s)
	}
	if s := c.Description.Default(); s != "english" {
		T.Errorf("invalid default: expected=english got=%s", s)
	}
}

func TestPathDeleteFromAltStringDefault(T *testing.T) {
	d := xmp.NewDocument()
	c := &dc.DublinCore{
		Description: xmp.NewAltString(
			xmp.AltItem{Value: "german", Lang: "de", IsDefault: false},
			xmp.AltItem{Value: "english", Lang: "en", IsDefault: true},
		),
	}
	d.AddModel(c)
	if err := d.SetPath(xmp.PathValue{
		Path:  xmp.Path("dc:description[en]"),
		Value: "",
		Flags: xmp.DELETE,
	}); err != nil {
		T.Errorf("unexpected error when deleting existing path: %v", err)
	}
	if len(c.Description) != 1 {
		T.Errorf("unexpected non-empty AltString after delete")
	}
	if v := dc.FindModel(d); v == nil {
		T.Errorf("unexpected empty model after delete")
	}
	if s := c.Description.Get("en"); s != "" {
		T.Errorf("invalid result: expected=<empty> got=%s", s)
	}
	if s := c.Description.Get("de"); s != "german" {
		T.Errorf("invalid result: expected=german got=%s", s)
	}
	if s := c.Description.Default(); s != "german" {
		T.Errorf("invalid default: expected=german got=%s", s)
	}
}

func TestPathDeleteNamespace(T *testing.T) {
	d := xmp.NewDocument()
	c := &dc.DublinCore{}
	d.AddModel(c)
	if err := d.SetPath(xmp.PathValue{
		Path:  xmp.Path("dc:"),
		Value: "",
		Flags: xmp.DELETE,
	}); err != nil {
		T.Errorf("unexpected error on delete existing namespace: %v", err)
	}
	if v := dc.FindModel(d); v != nil {
		T.Errorf("expected no model to exist")
	}
}
