// Copyright (c) 2013, Vastech SA (PTY) LTD. All rights reserved.
// http://code.google.com/p/gogoprotobuf
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are
// met:
//
//     * Redistributions of source code must retain the above copyright
// notice, this list of conditions and the following disclaimer.
//     * Redistributions in binary form must reproduce the above
// copyright notice, this list of conditions and the following disclaimer
// in the documentation and/or other materials provided with the
// distribution.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
// "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
// LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
// A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
// OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
// SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
// LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
// DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
// THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package custom

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/dropbox/goprotoc/test_config"
)

type MixMatch struct {
	Old []string
	New []string
}

func (this MixMatch) Regenerate() {
	data, err := ioutil.ReadFile("custom.proto")
	if err != nil {
		panic(err)
	}
	content := string(data)
	for i, old := range this.Old {
		content = strings.Replace(content, old, this.New[i], -1)
	}
	if err := ioutil.WriteFile("./testdata/custom.proto", []byte(content), 0666); err != nil {
		panic(err)
	}
	data2, err := ioutil.ReadFile("../custom_types.go")
	if err != nil {
		panic(err)
	}
	if err := ioutil.WriteFile("./testdata/custom_types.go", data2, 0666); err != nil {
		panic(err)
	}
	data3, err := ioutil.ReadFile("custom.test.golden")
	if err != nil {
		panic(err)
	}
	if err := ioutil.WriteFile("./testdata/custom_test.go", data3, 0666); err != nil {
		panic(err)
	}
	time.Sleep(10 * time.Millisecond)
	var regenerate = exec.Command("protoc", "--dgo_out=.", "-I="+config.ProtoPath, "./testdata/custom.proto")
	fmt.Printf("regenerating\n")
	out, err := regenerate.CombinedOutput()
	fmt.Printf("regenerate output: %v\n", string(out))
	if err != nil {
		panic(err)
	}
}

func (this MixMatch) test(t *testing.T, shouldPass bool) {
	if _, err := exec.LookPath("protoc"); err != nil {
		t.Skipf("cannot find protoc in PATH")
	}
	if _, err := exec.LookPath("go"); err != nil {
		t.Skipf("cannot find go in PATH")
	}
	if err := os.MkdirAll("./testdata", 0777); err != nil {
		panic(err)
	}
	this.Regenerate()
	var test = exec.Command("go", "test", "-v", "./testdata/")
	fmt.Printf("testing\n")
	out, err := test.CombinedOutput()
	fmt.Printf("test output: %v\n", string(out))
	if !shouldPass && err == nil {
		panic("Expected test failure. Invalid custom type declaration.")
	} else if shouldPass && err != nil {
		panic(err)
	}
	if err := os.RemoveAll("./testdata"); err != nil {
		panic(err)
	}
}

func TestDefault(t *testing.T) {
	MixMatch{}.test(t, true)
}

func TestCustomBytes(t *testing.T) {
	MixMatch{
		Old: []string{
			"optional bytes field5 = 5",
		},
		New: []string{
			"optional string field5 = 5",
		},
	}.test(t, false)

	MixMatch{
		Old: []string{
			"repeated bytes field15 = 15",
		},
		New: []string{
			"repeated string field15 = 15",
		},
	}.test(t, false)
}

func TestCustomTruncate(t *testing.T) {
	MixMatch{
		Old: []string{
			"optional int64 field1 = 1",
		},
		New: []string{
			"optional int32 field1 = 1",
		},
	}.test(t, false)

	MixMatch{
		Old: []string{
			"repeated int64 field11 = 11",
		},
		New: []string{
			"repeated uint64 field11 = 11",
		},
	}.test(t, false)

	MixMatch{
		Old: []string{
			"optional double field2 = 2",
		},
		New: []string{
			"optional float field2 = 2",
		},
	}.test(t, false)
}

func TestInvalidType(t *testing.T) {
	MixMatch{
		Old: []string{
			"optional bool field3 = 3 ",
		},
		New: []string{
			"optional int32 field3 = 3 ",
		},
	}.test(t, false)
}
