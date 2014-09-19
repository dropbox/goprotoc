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

/*
The testgen plugin generates Test and Benchmark functions for each message.

Tests are enabled using the following extensions:

  - testgen
  - testgen_all

Benchmarks are enabled using the following extensions:

  - benchgen
  - benchgen_all

Let us look at:

  code.google.com/p/gogoprotobuf/test/example/example.proto

Btw all the output can be seen at:

  code.google.com/p/gogoprotobuf/test/example/*

The following message:

  option (gogoproto.testgen_all) = true;
  option (gogoproto.benchgen_all) = true;

  message A {
	optional string Description = 1 [(gogoproto.nullable) = false];
	optional int64 Number = 2 [(gogoproto.nullable) = false];
	optional int64 Id = 3 [(gogoproto.customtype) = "github.com/dropbox/goprotoc/test.Id", (gogoproto.nullable) = false];
  }

given to the testgen plugin, will generate the following test code:

	func TestAProto(t *testing.T) {
		popr := math_rand.New(math_rand.NewSource(time.Now().UnixNano()))
		p := NewPopulatedA(popr, false)
		data, err := dropbox_gogoprotobuf_proto.Marshal(p)
		if err != nil {
			panic(err)
		}
		msg := &A{}
		if err := dropbox_gogoprotobuf_proto.Unmarshal(data, msg); err != nil {
			panic(err)
		}
		for i := range data {
			data[i] = byte(popr.Intn(256))
		}
		if err := p.VerboseEqual(msg); err != nil {
			t.Fatalf("%#v !VerboseProto %#v, since %v", msg, p, err)
		}
		if !p.Equal(msg) {
			t.Fatalf("%#v !Proto %#v", msg, p)
		}
	}

	func BenchmarkAProtoMarshal(b *testing.B) {
		popr := math_rand.New(math_rand.NewSource(616))
		total := 0
		pops := make([]*A, 10000)
		for i := 0; i < 10000; i++ {
			pops[i] = NewPopulatedA(popr, false)
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			data, err := dropbox_gogoprotobuf_proto.Marshal(pops[i%10000])
			if err != nil {
				panic(err)
			}
			total += len(data)
		}
		b.SetBytes(int64(total / b.N))
	}

	func BenchmarkAProtoUnmarshal(b *testing.B) {
		popr := math_rand.New(math_rand.NewSource(616))
		total := 0
		datas := make([][]byte, 10000)
		for i := 0; i < 10000; i++ {
			data, err := dropbox_gogoprotobuf_proto.Marshal(NewPopulatedA(popr, false))
			if err != nil {
				panic(err)
			}
			datas[i] = data
		}
		msg := &A{}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			total += len(datas[i%10000])
			if err := dropbox_gogoprotobuf_proto.Unmarshal(datas[i%10000], msg); err != nil {
				panic(err)
			}
		}
		b.SetBytes(int64(total / b.N))
	}

And the following message:
    message TestSingular {
        optional TestMessage msgs = 1;
        optional double vald = 2;
        optional float valf = 3;
    }

given to the testgen plugin, will generate the following test code:

    func TestTestEnumSingularAPI(t *testing1.T) {
        popr := math_rand1.New(math_rand1.NewSource(time1.Now().UnixNano()))
        p := NewPopulatedTestEnumSingular(popr, false)
        msg := &TestEnumSingular{}
        if !msg.apiEmptyTestEnumSingular() {
            t.Fatalf("TestEnumSingular should be empty")
        }
        apiCopyTestEnumSingular(msg, p, t)
        if err := p.VerboseEqual(msg); err != nil {
            t.Fatalf("%#v !VerboseEqual %#v, since %v", msg, p, err)
        }
        if msg.apiEmptyTestEnumSingular() {
            t.Fatalf("TestEnumSingular should not be empty")
        }
        if !p.Equal(msg) {
            t.Fatalf("%#v !Proto %#v", msg, p)
        }
        msg.Clear()
        if !msg.apiEmptyTestEnumSingular() {
            t.Fatalf("TestEnumSingular should be empty")
        }
    }

    func apiCopyTestEnumSingular(dst *TestEnumSingular, src *TestEnumSingular, t *testing1.T) {
        if dst == nil || src == nil {
            t.Fatalf("Cannot copy to(%v) or from(%v) nil message", dst, src)
        }
        dst.SetField(src.GetField())
        src.XXX_unrecognized = dst.XXX_unrecognized
    }

    func apiEmptyTestEnumSingular(msg *ccTypeName, t *testing1.T) bool {
        if msg == nil {
            return true
        }
        if msg.HasField() {
            return false
        }
        return true
    }

Other registered tests are also generated.
Tests are registered to this test plugin by calling the following function.

  func RegisterTestPlugin(newFunc NewTestPlugin)

where NewTestPlugin is:

  type NewTestPlugin func(g *generator.Generator) TestPlugin

and TestPlugin is an interface:

  type TestPlugin interface {
	Generate(imports generator.PluginImports, file *generator.FileDescriptor) (used bool)
  }

Plugins that use this interface include:

  - populate
  - gostring
  - equal
  - union
  - and more

Please look at these plugins as examples of how to create your own.
A good idea is to let each plugin generate its own tests.

*/
package testgen

import (
	"github.com/dropbox/goprotoc/gogoproto"
	"github.com/dropbox/goprotoc/protoc-gen-dgo/generator"
)

type TestPlugin interface {
	Generate(imports generator.PluginImports, file *generator.FileDescriptor) (used bool)
}

type NewTestPlugin func(g *generator.Generator) TestPlugin

var testplugins = make([]NewTestPlugin, 0)

func RegisterTestPlugin(newFunc NewTestPlugin) {
	testplugins = append(testplugins, newFunc)
}

type plugin struct {
	*generator.Generator
	generator.PluginImports
	tests []TestPlugin
}

func NewPlugin() *plugin {
	return &plugin{}
}

func (p *plugin) Name() string {
	return "testgen"
}

func (p *plugin) Init(g *generator.Generator) {
	p.Generator = g
	p.tests = make([]TestPlugin, 0, len(testplugins))
	for i := range testplugins {
		p.tests = append(p.tests, testplugins[i](g))
	}
}

func (p *plugin) Generate(file *generator.FileDescriptor) {
	p.PluginImports = generator.NewPluginImports(p.Generator)
	atLeastOne := false
	for i := range p.tests {
		used := p.tests[i].Generate(p.PluginImports, file)
		if used {
			atLeastOne = true
		}
	}
	if atLeastOne {
		p.P(`//These tests are generated by github.com/dropbox/goprotoc/plugin/testgen`)
	}
}

type testProto struct {
	*generator.Generator
}

func newProto(g *generator.Generator) TestPlugin {
	return &testProto{g}
}

func (p *testProto) Generate(imports generator.PluginImports, file *generator.FileDescriptor) bool {
	used := false
	testingPkg := imports.NewImport("testing")
	randPkg := imports.NewImport("math/rand")
	timePkg := imports.NewImport("time")
	protoPkg := imports.NewImport("github.com/dropbox/goprotoc/proto")
	for _, message := range file.Messages() {
		ccTypeName := generator.CamelCaseSlice(message.TypeName())
		if gogoproto.HasTestGen(file.FileDescriptorProto, message.DescriptorProto) {
			used = true

			p.P(`func Test`, ccTypeName, `Proto(t *`, testingPkg.Use(), `.T) {`)
			p.In()
			p.P(`popr := `, randPkg.Use(), `.New(`, randPkg.Use(), `.NewSource(`, timePkg.Use(), `.Now().UnixNano()))`)
			p.P(`p := NewPopulated`, ccTypeName, `(popr, false)`)
			p.P(`data, err := `, protoPkg.Use(), `.Marshal(p)`)
			p.P(`if err != nil {`)
			p.In()
			p.P(`panic(err)`)
			p.Out()
			p.P(`}`)
			p.P(`msg := &`, ccTypeName, `{}`)
			p.P(`if err := `, protoPkg.Use(), `.Unmarshal(data, msg); err != nil {`)
			p.In()
			p.P(`panic(err)`)
			p.Out()
			p.P(`}`)
			p.P(`for i := range data {`)
			p.In()
			p.P(`data[i] = byte(popr.Intn(256))`)
			p.Out()
			p.P(`}`)
			if gogoproto.HasVerboseEqual(file.FileDescriptorProto, message.DescriptorProto) {
				p.P(`if err := p.VerboseEqual(msg); err != nil {`)
				p.In()
				p.P(`t.Fatalf("%#v !VerboseProto %#v, since %v", msg, p, err)`)
				p.Out()
				p.P(`}`)
			}
			p.P(`if !p.Equal(msg) {`)
			p.In()
			p.P(`t.Fatalf("%#v !Proto %#v", msg, p)`)
			p.Out()
			p.P(`}`)
			p.Out()
			p.P(`}`)
			p.P()
		}

		if gogoproto.HasTestGen(file.FileDescriptorProto, message.DescriptorProto) {
			if gogoproto.IsMarshaler(file.FileDescriptorProto, message.DescriptorProto) {
				p.P(`func Test`, ccTypeName, `MarshalTo(t *`, testingPkg.Use(), `.T) {`)
				p.In()
				p.P(`popr := `, randPkg.Use(), `.New(`, randPkg.Use(), `.NewSource(`, timePkg.Use(), `.Now().UnixNano()))`)
				p.P(`p := NewPopulated`, ccTypeName, `(popr, false)`)
				p.P(`size := p.Size()`)
				p.P(`data := make([]byte, size)`)
				p.P(`for i := range data {`)
				p.In()
				p.P(`data[i] = byte(popr.Intn(256))`)
				p.Out()
				p.P(`}`)
				p.P(`_, err := p.MarshalTo(data)`)
				p.P(`if err != nil {`)
				p.In()
				p.P(`panic(err)`)
				p.Out()
				p.P(`}`)
				p.P(`msg := &`, ccTypeName, `{}`)
				p.P(`if err := `, protoPkg.Use(), `.Unmarshal(data, msg); err != nil {`)
				p.In()
				p.P(`panic(err)`)
				p.Out()
				p.P(`}`)
				p.P(`for i := range data {`)
				p.In()
				p.P(`data[i] = byte(popr.Intn(256))`)
				p.Out()
				p.P(`}`)
				if gogoproto.HasVerboseEqual(file.FileDescriptorProto, message.DescriptorProto) {
					p.P(`if err := p.VerboseEqual(msg); err != nil {`)
					p.In()
					p.P(`t.Fatalf("%#v !VerboseProto %#v, since %v", msg, p, err)`)
					p.Out()
					p.P(`}`)
				}
				p.P(`if !p.Equal(msg) {`)
				p.In()
				p.P(`t.Fatalf("%#v !Proto %#v", msg, p)`)
				p.Out()
				p.P(`}`)
				p.Out()
				p.P(`}`)
				p.P()
			}
		}

		if gogoproto.HasBenchGen(file.FileDescriptorProto, message.DescriptorProto) {
			used = true
			p.P(`func Benchmark`, ccTypeName, `ProtoMarshal(b *`, testingPkg.Use(), `.B) {`)
			p.In()
			p.P(`popr := `, randPkg.Use(), `.New(`, randPkg.Use(), `.NewSource(616))`)
			p.P(`total := 0`)
			p.P(`pops := make([]*`, ccTypeName, `, 10000)`)
			p.P(`for i := 0; i < 10000; i++ {`)
			p.In()
			p.P(`pops[i] = NewPopulated`, ccTypeName, `(popr, false)`)
			p.Out()
			p.P(`}`)
			p.P(`b.ResetTimer()`)
			p.P(`for i := 0; i < b.N; i++ {`)
			p.In()
			p.P(`data, err := `, protoPkg.Use(), `.Marshal(pops[i%10000])`)
			p.P(`if err != nil {`)
			p.In()
			p.P(`panic(err)`)
			p.Out()
			p.P(`}`)
			p.P(`total += len(data)`)
			p.Out()
			p.P(`}`)
			p.P(`b.SetBytes(int64(total / b.N))`)
			p.Out()
			p.P(`}`)
			p.P()

			p.P(`func Benchmark`, ccTypeName, `ProtoUnmarshal(b *`, testingPkg.Use(), `.B) {`)
			p.In()
			p.P(`popr := `, randPkg.Use(), `.New(`, randPkg.Use(), `.NewSource(616))`)
			p.P(`total := 0`)
			p.P(`datas := make([][]byte, 10000)`)
			p.P(`for i := 0; i < 10000; i++ {`)
			p.In()
			p.P(`data, err := `, protoPkg.Use(), `.Marshal(NewPopulated`, ccTypeName, `(popr, false))`)
			p.P(`if err != nil {`)
			p.In()
			p.P(`panic(err)`)
			p.Out()
			p.P(`}`)
			p.P(`datas[i] = data`)
			p.Out()
			p.P(`}`)
			p.P(`msg := &`, ccTypeName, `{}`)
			p.P(`b.ResetTimer()`)
			p.P(`for i := 0; i < b.N; i++ {`)
			p.In()
			p.P(`total += len(datas[i%10000])`)
			p.P(`if err := `, protoPkg.Use(), `.Unmarshal(datas[i%10000], msg); err != nil {`)
			p.In()
			p.P(`panic(err)`)
			p.Out()
			p.P(`}`)
			p.Out()
			p.P(`}`)
			p.P(`b.SetBytes(int64(total / b.N))`)
			p.Out()
			p.P(`}`)
			p.P()
		}
	}
	return used
}

type testAPI struct {
	*generator.Generator
}

func newAPI(g *generator.Generator) TestPlugin {
	return &testAPI{g}
}

func (p *testAPI) Generate(imports generator.PluginImports, file *generator.FileDescriptor) bool {
	used := false
	testingPkg := imports.NewImport("testing")
	randPkg := imports.NewImport("math/rand")
	timePkg := imports.NewImport("time")
	for _, message := range file.Messages() {
		ccTypeName := generator.CamelCaseSlice(message.TypeName())
		if !gogoproto.HasTestGen(file.FileDescriptorProto, message.DescriptorProto) {
			continue
		}
		used = true
		p.P(`func Test`, ccTypeName, `API(t *`, testingPkg.Use(), `.T) {`)
		p.In()
		p.P(`popr := `, randPkg.Use(), `.New(`, randPkg.Use(), `.NewSource(`, timePkg.Use(),
			`.Now().UnixNano()))`)
		p.P(`p := NewPopulated`, ccTypeName, `(popr, false)`)
		p.P(`msg := &`, ccTypeName, `{}`)
		p.P(`if !apiEmpty`, ccTypeName, `(msg, t) {`)
		p.In()
		p.P(`t.Fatalf("`, ccTypeName, ` should be empty")`)
		p.Out()
		p.P(`}`)
		p.P(`apiCopy`, ccTypeName, `(msg, p, t)`)
		if gogoproto.HasVerboseEqual(file.FileDescriptorProto, message.DescriptorProto) {
			p.P(`if err := p.VerboseEqual(msg); err != nil {`)
			p.In()
			p.P(`t.Fatalf("%#v !VerboseEqual %#v, since %v", msg, p, err)`)
			p.Out()
			p.P(`}`)
		}
		p.P(`if apiEmpty`, ccTypeName, `(p, t) != apiEmpty`, ccTypeName, `(msg, t) {`)
		p.In()
		p.P(`t.Fatalf("`, ccTypeName, ` should not be empty")`)
		p.Out()
		p.P(`}`)
		p.P(`if !p.Equal(msg) {`)
		p.In()
		p.P(`t.Fatalf("%#v !Proto %#v", msg, p)`)
		p.Out()
		p.P(`}`)
		p.P(`msg.Clear()`)
		p.P(`if !apiEmpty`, ccTypeName, `(msg, t) {`)
		p.In()
		p.P(`t.Fatalf("`, ccTypeName, ` should be empty")`)
		p.Out()
		p.P(`}`)
		p.Out()
		p.P(`}`)
		p.P()

		p.P(`func apiCopy`, ccTypeName, `(dst *`, ccTypeName, `, src *`, ccTypeName, `, t *`,
			testingPkg.Use(), `.T) {`)
		p.In()
		p.P(`if dst == nil || src == nil {`)
		p.In()
		p.P(`t.Fatalf("Cannot copy to(%v) or from(%v) nil message", dst, src)`)
		p.Out()
		p.P(`}`)
		for _, field := range message.Field {
			if gogoproto.IsEmbed(field) {
				p.P(`t.Skip("Cannot copy embed field")`)
				break
			}
			if gogoproto.IsCustomType(field) {
				p.P(`t.Skip("Cannot copy costum field")`)
				break
			}
			fieldName := generator.CamelCase(p.GetFieldName(message, field))
			fieldType, _ := p.GoType(message, field)
			fieldTypeName := generator.GoTypeToName(fieldType)
			if generator.IsRepeated(field) {
				p.P(`for i := 0; i < src.`, fieldName, `Size(); i++ {`)
				p.In()
				if generator.IsMessageType(field) {
					p.P(`src`, fieldName, `, _ := src.Get`, fieldName, `(i)`)
					p.P(`dst`, fieldName, `, _ := dst.Add`, fieldName, `()`)
					p.P(`apiCopy`, fieldTypeName, `(dst`, fieldName, `, src`, fieldName, `, t)`)
				} else {
					p.P(`value, _ := src.Get`, fieldName, `(i)`)
					p.P(`dst.Add`, fieldName, `(value)`)
				}
				p.Out()
				p.P(`}`)
			} else {
				p.P(`if src.Has`, fieldName, `() {`)
				p.In()
				if generator.IsMessageType(field) {
					p.P(`src`, fieldName, ` := src.Get`, fieldName, `()`)
					p.P(`dst`, fieldName, `, _ := dst.Mutate`, fieldName, `()`)
					p.P(`apiCopy`, fieldTypeName, `(dst`, fieldName, `, src`, fieldName, `, t)`)
				} else {
					p.P(`dst.Set`, fieldName, `(src.Get`, fieldName, `())`)
				}
				p.Out()
				p.P(`}`)
			}
		}
		p.P(`src.XXX_unrecognized = dst.XXX_unrecognized`)
		if len(message.ExtensionRange) > 0 {
			p.P(`src.XXX_extensions = dst.XXX_extensions`)
		}
		p.Out()
		p.P(`}`)
		p.P()

		p.P(`func apiEmpty`, ccTypeName, `(msg *`, ccTypeName, `, t *`, testingPkg.Use(), `.T) bool {`)
		p.In()
		p.P(`if msg == nil {`)
		p.In()
		p.P(`return true`)
		p.Out()
		p.P(`}`)
		for _, field := range message.Field {
			if gogoproto.IsEmbed(field) {
				p.P(`t.Skip("Cannot check embed field")`)
				break
			}
			if gogoproto.IsCustomType(field) {
				p.P(`t.Skip("Cannot check costum field")`)
				break
			}
			fieldName := generator.CamelCase(p.GetFieldName(message, field))
			if generator.IsRepeated(field) {
				p.P(`if msg.`, fieldName, `Size() != 0 {`)
			} else {
				p.P(`if msg.Has`, fieldName, `() {`)
			}
			p.In()
			p.P(`return false`)
			p.Out()
			p.P(`}`)
		}
		p.P(`return true`)
		p.Out()
		p.P(`}`)
		p.P()
	}
	return used
}

func init() {
	RegisterTestPlugin(newProto)
	RegisterTestPlugin(newAPI)
}
