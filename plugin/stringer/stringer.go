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
The stringer plugin generates a String method for each message.

It is enabled by the following extensions:

  - stringer
  - stringer_all

The stringer plugin also generates a test given it is enabled using one of the following extensions:

  - testgen
  - testgen_all

Let us look at:

  code.google.com/p/gogoprotobuf/test/example/example.proto

Btw all the output can be seen at:

  code.google.com/p/gogoprotobuf/test/example/*

The following message:

  option (gogoproto.goproto_stringer_all) = false;
  option (gogoproto.stringer_all) =  true;

  message A {
	optional string Description = 1;
	optional int64 Number = 2;
	optional bytes Id = 3 [(gogoproto.customtype) = "github.com/dropbox/goprotoc/test/custom.Uuid"];
  }

given to the stringer stringer, will generate the following code:

  func (this *A) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&A{`,
		`Description:` + fmt.Sprintf("%v", this.Description) + `,`,
		`Number:` + fmt.Sprintf("%v", this.Number) + `,`,
		`Id:` + fmt.Sprintf("%v", this.Id) + `,`,
		`XXX_unrecognized:` + fmt.Sprintf("%v", this.XXX_unrecognized) + `,`,
		`}`,
	}, "")
	return s
  }

and the following test code:

	func TestAStringer(t *testing4.T) {
		popr := math_rand4.New(math_rand4.NewSource(time4.Now().UnixNano()))
		p := NewPopulatedA(popr, false)
		s1 := p.String()
		s2 := fmt1.Sprintf("%v", p)
		if s1 != s2 {
			t.Fatalf("String want %v got %v", s1, s2)
		}
	}

Typically fmt.Printf("%v") will stop to print when it reaches a pointer and
not print their values, while the generated String method will always print all values, recursively.

*/
package stringer

import (
    "github.com/dropbox/goprotoc/gogoproto"
    "github.com/dropbox/goprotoc/protoc-gen-dgo/generator"
    "strings"
)

type stringer struct {
    *generator.Generator
    generator.PluginImports
    atleastOne bool
    localName  string
}

func NewStringer() *stringer {
    return &stringer{}
}

func (p *stringer) Name() string {
    return "stringer"
}

func (p *stringer) Init(g *generator.Generator) {
    p.Generator = g
}

func (p *stringer) Generate(file *generator.FileDescriptor) {
    p.PluginImports = generator.NewPluginImports(p.Generator)
    p.atleastOne = false

    p.localName = generator.FileName(file)

    fmtPkg := p.NewImport("fmt")
    stringsPkg := p.NewImport("strings")
    for _, message := range file.Messages() {
        if !gogoproto.IsStringer(file.FileDescriptorProto, message.DescriptorProto) {
            continue
        }
        p.atleastOne = true
        ccTypeName := generator.CamelCaseSlice(message.TypeName())
        p.P(`func (this *`, ccTypeName, `) String() string {`)
        p.In()
        p.P(`if this == nil {`)
        p.In()
        p.P(`return "nil"`)
        p.Out()
        p.P(`}`)
        p.P("s := ", stringsPkg.Use(), ".Join([]string{`&", ccTypeName, "{`,")
        for _, field := range message.Field {
            fieldname := p.GetFieldName(message, field)
            if field.IsMessage() || p.IsGroup(field) {
                desc := p.ObjectNamed(field.GetTypeName())
                msgname := p.TypeName(desc)
                msgnames := strings.Split(msgname, ".")
                typeName := msgnames[len(msgnames)-1]
                if field.IsRepeated() {
                    p.P("`", fieldname, ":`", ` + `, stringsPkg.Use(), `.Replace(`, fmtPkg.Use(),
                        `.Sprintf("%v", this.`, fieldname, `[:this.`, generator.SizerName(fieldname),
                        `]), "`, typeName, `","`, msgname, `"`, ", 1) + `,", "`,")
                } else {
                    fieldValue := "this.Get" + generator.CamelCase(fieldname) + "()"
                    if gogoproto.IsCustomType(field) || gogoproto.IsEmbed(field) {
                        fieldValue = "this." + fieldname
                    }
                    p.P("`", fieldname, ":`", ` + `, stringsPkg.Use(), `.Replace(`, fmtPkg.Use(),
                        `.Sprintf("%v", `, fieldValue, `), "`, typeName, `","`, msgname, `"`,
                        ", 1) + `,", "`,")
                }
            } else if field.IsRepeated() {
                p.P("`", fieldname, ":`", ` + `, fmtPkg.Use(), `.Sprintf("%v", this.`, fieldname,
                    "[:this.", generator.SizerName(fieldname), "]) + `,", "`,")
            } else {
                fieldValue := "this.Get" + generator.CamelCase(fieldname) + "()"
                if gogoproto.IsCustomType(field) || gogoproto.IsEmbed(field) {
                    fieldValue = "this." + fieldname
                }
                p.P("`", fieldname, ":`", ` + `, fmtPkg.Use(), `.Sprintf("%v", `, fieldValue, ") + `,", "`,")
            }
        }
        if message.DescriptorProto.HasExtension() {
            if gogoproto.HasExtensionsMap(file.FileDescriptorProto, message.DescriptorProto) {
                p.P("`XXX_extensions:` + proto.StringFromExtensionsMap(this.XXX_extensions) + `,`,")
            } else {
                p.P("`XXX_extensions:` + proto.StringFromExtensionsBytes(this.XXX_extensions) + `,`,")
            }
        }
        p.P("`XXX_unrecognized:` + ", fmtPkg.Use(), `.Sprintf("%v", this.XXX_unrecognized) + `, "`,`,")
        p.P("`}`,")
        p.P(`}`, `,""`, ")")
        p.P(`return s`)
        p.Out()
        p.P(`}`)
    }

    if !p.atleastOne {
        return
    }

}

func init() {
    generator.RegisterPlugin(NewStringer())
}
