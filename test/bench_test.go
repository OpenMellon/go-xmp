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
	"encoding/json"
	"testing"

	_ "github.com/OpenMellon/go-xmp/models"
	"github.com/OpenMellon/go-xmp/xmp"
)

// Benchmarks
func BenchmarkUnmarshalXMP_5kB(B *testing.B)  { runUnmarshalXmpBench(B, []byte(data_5kb)) }
func BenchmarkMarshalXMP_5kB(B *testing.B)    { runMarshalXmpBench(B, []byte(data_5kb)) }
func BenchmarkMarshalJSON_5kB(B *testing.B)   { runMarshalJsonBench(B, []byte(data_5kb)) }
func BenchmarkUnmarshalJSON_5kB(B *testing.B) { runUnmarshalJsonBench(B, []byte(data_5kb)) }

func BenchmarkUnmarshalXMP_85kB(B *testing.B)  { runUnmarshalXmpBench(B, []byte(data_85kb)) }
func BenchmarkMarshalXMP_85kB(B *testing.B)    { runMarshalXmpBench(B, []byte(data_85kb)) }
func BenchmarkMarshalJSON_85kB(B *testing.B)   { runMarshalJsonBench(B, []byte(data_85kb)) }
func BenchmarkUnmarshalJSON_85kB(B *testing.B) { runUnmarshalJsonBench(B, []byte(data_85kb)) }

func runUnmarshalXmpBench(B *testing.B, data []byte) {
	xmp.SetLogLevel(xmp.LogLevelError)
	for i := 0; i < B.N; i++ {
		d := &xmp.Document{}
		if err := xmp.Unmarshal(data, d); err != nil {
			B.Fatal(err)
		}
		d.Close()
	}
}

func runMarshalXmpBench(B *testing.B, data []byte) {
	xmp.SetLogLevel(xmp.LogLevelError)
	d := &xmp.Document{}
	if err := xmp.Unmarshal(data, d); err != nil {
		B.Fatal(err)
	}
	B.ResetTimer()
	var err error
	for i := 0; i < B.N; i++ {
		buf, err = xmp.Marshal(d)
		if err != nil {
			B.Fatal(err)
		}
	}
	d.Close()
}

func runMarshalJsonBench(B *testing.B, data []byte) {
	xmp.SetLogLevel(xmp.LogLevelError)
	d := &xmp.Document{}
	if err := xmp.Unmarshal(data, d); err != nil {
		B.Fatal(err)
	}
	B.ResetTimer()
	var err error
	for i := 0; i < B.N; i++ {
		buf, err = json.MarshalIndent(d, "", "  ")
		if err != nil {
			B.Fatal(err)
		}
	}
	d.Close()
}

func runUnmarshalJsonBench(B *testing.B, data []byte) {
	xmp.SetLogLevel(xmp.LogLevelError)
	d := &xmp.Document{}
	if err := xmp.Unmarshal(data, d); err != nil {
		B.Fatal(err)
	}
	buf, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		B.Fatal(err)
	}
	B.ResetTimer()
	for i := 0; i < B.N; i++ {
		dx := &xmp.Document{}
		if err = json.Unmarshal(buf, dx); err != nil {
			B.Fatal(err)
		}
	}
	d.Close()
}
