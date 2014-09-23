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
The unmarshal code generates a Unmarshal method for each message.
The `Unmarshal([]byte) error` method results in the fact that the message
implements the Unmarshaler interface.
The allows proto.Unmarshal to be faster by calling the generated Unmarshal method rather than using reflect.

The generation of unmarshalling tests are enabled using one of the following extensions:

  - testgen
  - testgen_all

And benchmarks given it is enabled using one of the following extensions:

  - benchgen
  - benchgen_all

Let us look at:

  code.google.com/p/gogoprotobuf/test/example/example.proto

Btw all the output can be seen at:

  code.google.com/p/gogoprotobuf/test/example/*

The following message:

  option (gogoproto.unmarshaler_all) = true;

  message B {
	option (gogoproto.description) = true;
	optional string A = 1 [(gogoproto.embed) = true];
	repeated int64 G = 2 [(gogoproto.customtype) = "github.com/dropbox/goprotoc/test.Id"];
  }

the unmarshal will generate the following code:

	func (m *B) Unmarshal(data []byte) error {
		l := len(data)
		index := 0
		for index < l {
			var wire uint64
			for shift := uint(0); ; shift += 7 {
				if index >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[index]
				index++
				wire |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			fieldNum := int32(wire >> 3)
			wireType := int(wire & 0x7)
			switch fieldNum {
			case 1:
				if wireType != 2 {
					return fmt.Errorf("proto: wrong wireType = %d for field a", wireType)
				}
				m.xxx_IsASet = true
				var stringLen uint64
				for shift := uint(0); ; shift += 7 {
					if index >= l {
						return io.ErrUnexpectedEOF
					}
					b := data[index]
					index++
					stringLen |= (uint64(b) & 0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				postIndex := index + int(stringLen)
				if postIndex > l {
					return io.ErrUnexpectedEOF
				}
				m.a = string(data[index:postIndex])
				index = postIndex
			case 2:
				if wireType != 0 {
					return fmt.Errorf("proto: wrong wireType = %d for field g", wireType)
				}
				m.xxx_LenG += 1
				var v int64
				for shift := uint(0); ; shift += 7 {
					if index >= l {
						return io.ErrUnexpectedEOF
					}
					b := data[index]
					index++
					v |= (int64(b) & 0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				m.g = append(m.g, github_com_dropbox_goprotoc_test.Id(v))
			default:
				var sizeOfWire int
				for {
					sizeOfWire++
					wire >>= 7
					if wire == 0 {
						break
					}
				}
				index -= sizeOfWire
				skippy, err := proto.Skip(data[index:])
				if err != nil {
					return err
				}
				if (index + skippy) > l {
					return io.ErrUnexpectedEOF
				}
				m.XXX_unrecognized = append(m.XXX_unrecognized, data[index:index+skippy]...)
				index += skippy
			}
		}
		return nil
	}

Remember when using this code to call proto.Unmarshal.
This will call m.Reset and invoke the generated Unmarshal method for you.
If you call m.Unmarshal without m.Reset you could be merging protocol buffers.

*/
package generator

import (
	"fmt"
	"github.com/dropbox/goprotoc/gogoproto"
	"github.com/dropbox/goprotoc/proto"
	descriptor "github.com/dropbox/goprotoc/protoc-gen-dgo/descriptor"
	"strconv"
	"strings"
)

func (g *Generator) decodeVarint(varName string, typName string) {
	g.P(`for shift := uint(0); ; shift += 7 {`)
	g.In()
	g.P(`if index >= l {`)
	g.In()
	g.P(`return `, g.Pkg["io"], `.ErrUnexpectedEOF`)
	g.Out()
	g.P(`}`)
	g.P(`b := data[index]`)
	g.P(`index++`)
	g.P(varName, ` |= (`, typName, `(b) & 0x7F) << shift`)
	g.P(`if b < 0x80 {`)
	g.In()
	g.P(`break`)
	g.Out()
	g.P(`}`)
	g.Out()
	g.P(`}`)
}

func (g *Generator) decodeFixed32(varName string, typeName string) {
	g.P(`i := index + 4`)
	g.P(`if i > l {`)
	g.In()
	g.P(`return `, g.Pkg["io"], `.ErrUnexpectedEOF`)
	g.Out()
	g.P(`}`)
	g.P(`index = i`)
	g.P(varName, ` = `, typeName, `(data[i-4])`)
	g.P(varName, ` |= `, typeName, `(data[i-3]) << 8`)
	g.P(varName, ` |= `, typeName, `(data[i-2]) << 16`)
	g.P(varName, ` |= `, typeName, `(data[i-1]) << 24`)
}

func (g *Generator) decodeFixed64(varName string, typeName string) {
	g.P(`i := index + 8`)
	g.P(`if i > l {`)
	g.In()
	g.P(`return `, g.Pkg["io"], `.ErrUnexpectedEOF`)
	g.Out()
	g.P(`}`)
	g.P(`index = i`)
	g.P(varName, ` = `, typeName, `(data[i-8])`)
	g.P(varName, ` |= `, typeName, `(data[i-7]) << 8`)
	g.P(varName, ` |= `, typeName, `(data[i-6]) << 16`)
	g.P(varName, ` |= `, typeName, `(data[i-5]) << 24`)
	g.P(varName, ` |= `, typeName, `(data[i-4]) << 32`)
	g.P(varName, ` |= `, typeName, `(data[i-3]) << 40`)
	g.P(varName, ` |= `, typeName, `(data[i-2]) << 48`)
	g.P(varName, ` |= `, typeName, `(data[i-1]) << 56`)
}

func (g *Generator) field(field *descriptor.FieldDescriptorProto, fieldname string) {
	repeated := field.IsRepeated()
	gotype, _ := g.GoType(nil, field)
	fieldtype := GoTypeToName(gotype)
	if repeated {
		g.P(`m.`, SizerName(fieldname), ` += 1`)
	} else {
		g.P(`m.`, SetterName(fieldname), ` = true`)
	}
	if gogoproto.IsCustomType(field) {
		_, typ, err := GetCustomType(field)
		if err != nil {
			panic(err)
		}
		fieldtype = typ
	}
	switch *field.Type {
	case descriptor.FieldDescriptorProto_TYPE_DOUBLE:
		if repeated {
			g.P(`var v uint64`)
			g.decodeFixed64("v", "uint64")
			g.P(`v2 := `, g.Pkg["math"], `.Float64frombits(v)`)
			g.P(`m.`, fieldname, ` = append(m.`, fieldname, `, `, fieldtype, `(v2))`)
		} else {
			g.P(`var v uint64`)
			g.decodeFixed64("v", "uint64")
			g.P(`m.`, fieldname, ` = `, fieldtype, `(`, g.Pkg["math"], `.Float64frombits(v))`)
		}
	case descriptor.FieldDescriptorProto_TYPE_FLOAT:
		if repeated {
			g.P(`var v uint32`)
			g.decodeFixed32("v", "uint32")
			g.P(`v2 := `, g.Pkg["math"], `.Float32frombits(v)`)
			g.P(`m.`, fieldname, ` = append(m.`, fieldname, `, `, fieldtype, `(v2))`)
		} else {
			g.P(`var v uint32`)
			g.decodeFixed32("v", "uint32")
			g.P(`m.`, fieldname, ` = `, fieldtype, `(`, g.Pkg["math"], `.Float32frombits(v))`)
		}
	case descriptor.FieldDescriptorProto_TYPE_INT64:
		if repeated {
			g.P(`var v int64`)
			g.decodeVarint("v", "int64")
			g.P(`m.`, fieldname, ` = append(m.`, fieldname, `, `, fieldtype, `(v))`)
		} else {
			g.decodeVarint("m."+fieldname, fieldtype)
		}
	case descriptor.FieldDescriptorProto_TYPE_UINT64:
		if repeated {
			g.P(`var v uint64`)
			g.decodeVarint("v", "uint64")
			g.P(`m.`, fieldname, ` = append(m.`, fieldname, `, `, fieldtype, `(v))`)
		} else {
			g.decodeVarint("m."+fieldname, fieldtype)
		}
	case descriptor.FieldDescriptorProto_TYPE_INT32:
		if repeated {
			g.P(`var v int32`)
			g.decodeVarint("v", "int32")
			g.P(`m.`, fieldname, ` = append(m.`, fieldname, `, `, fieldtype, `(v))`)
		} else {
			g.decodeVarint("m."+fieldname, fieldtype)
		}
	case descriptor.FieldDescriptorProto_TYPE_FIXED64:
		if repeated {
			g.P(`var v uint64`)
			g.decodeFixed64("v", "uint64")
			g.P(`m.`, fieldname, ` = append(m.`, fieldname, `, `, fieldtype, `(v))`)
		} else {
			g.decodeFixed64("m."+fieldname, fieldtype)
		}
	case descriptor.FieldDescriptorProto_TYPE_FIXED32:
		if repeated {
			g.P(`var v uint32`)
			g.decodeFixed32("v", "uint32")
			g.P(`m.`, fieldname, ` = append(m.`, fieldname, `, `, fieldtype, `(v))`)
		} else {
			g.decodeFixed32("m."+fieldname, fieldtype)
		}
	case descriptor.FieldDescriptorProto_TYPE_BOOL:
		if repeated {
			g.P(`var v int`)
			g.decodeVarint("v", "int")
			g.P(`m.`, fieldname, ` = append(m.`, fieldname, `, `, fieldtype, `(bool(v != 0)))`)
		} else {
			g.P(`var v int`)
			g.decodeVarint("v", "int")
			g.P(`m.`, fieldname, ` = `, fieldtype, `(bool(v != 0))`)
		}
	case descriptor.FieldDescriptorProto_TYPE_STRING:
		g.P(`var stringLen uint64`)
		g.decodeVarint("stringLen", "uint64")
		g.P(`postIndex := index + int(stringLen)`)
		g.P(`if postIndex > l {`)
		g.In()
		g.P(`return `, g.Pkg["io"], `.ErrUnexpectedEOF`)
		g.Out()
		g.P(`}`)
		if repeated {
			g.P(`m.`, fieldname, ` = append(m.`, fieldname, `, `, fieldtype, `(data[index:postIndex]))`)
		} else {
			g.P(`m.`, fieldname, ` = `, fieldtype, `(data[index:postIndex])`)
		}
		g.P(`index = postIndex`)
	case descriptor.FieldDescriptorProto_TYPE_GROUP:
		panic(fmt.Errorf("unmarshaler does not support group %v", fieldname))
	case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
		desc := g.ObjectNamed(field.GetTypeName())
		msgname := g.TypeName(desc)
		g.P(`var msglen int`)
		g.decodeVarint("msglen", "int")
		g.P(`postIndex := index + msglen`)
		g.P(`if postIndex > l {`)
		g.In()
		g.P(`return `, g.Pkg["io"], `.ErrUnexpectedEOF`)
		g.Out()
		g.P(`}`)
		if repeated {
			g.P(`m.`, fieldname, ` = append(m.`, fieldname, `, &`, msgname, `{})`)
			g.P(`m.`, fieldname, `[len(m.`, fieldname, `)-1].Unmarshal(data[index:postIndex])`)
		} else {
			g.P(`m.`, fieldname, ` = &`, msgname, `{}`)
			g.P(`if err := m.`, fieldname, `.Unmarshal(data[index:postIndex]); err != nil {`)
			g.In()
			g.P(`return err`)
			g.Out()
			g.P(`}`)
		}
		g.P(`index = postIndex`)
	case descriptor.FieldDescriptorProto_TYPE_BYTES:
		g.P(`var byteLen int`)
		g.decodeVarint("byteLen", "int")
		g.P(`postIndex := index + byteLen`)
		g.P(`if postIndex > l {`)
		g.In()
		g.P(`return `, g.Pkg["io"], `.ErrUnexpectedEOF`)
		g.Out()
		g.P(`}`)
		if !gogoproto.IsCustomType(field) {
			if repeated {
				g.P(`m.`, fieldname, ` = append(m.`, fieldname, `, make([]byte, postIndex-index))`)
				g.P(`copy(m.`, fieldname, `[len(m.`, fieldname, `)-1], data[index:postIndex])`)
			} else {
				g.P(`m.`, fieldname, ` = append(m.`, fieldname, `, data[index:postIndex]...)`)
			}
		} else {
			if repeated {
				g.P(`m.`, fieldname, ` = append(m.`, fieldname, `, make(([]byte), postIndex-index))`)
				g.P(`copy(m.`, fieldname, `[len(m.`, fieldname, `)-1], data[index:postIndex])`)
			} else {
				g.P(`m.`, fieldname, ` = append(m.`, fieldname, `, data[index:postIndex]...)`)
			}
		}
		g.P(`index = postIndex`)
	case descriptor.FieldDescriptorProto_TYPE_UINT32:
		if repeated {
			g.P(`var v uint32`)
			g.decodeVarint("v", "uint32")
			g.P(`m.`, fieldname, ` = append(m.`, fieldname, `, `, fieldtype, `(v))`)
		} else {
			g.decodeVarint("m."+fieldname, fieldtype)
		}
	case descriptor.FieldDescriptorProto_TYPE_ENUM:
		if !gogoproto.IsCustomType(field) {
			typName := g.TypeName(g.ObjectNamed(field.GetTypeName()))
			if repeated {
				g.P(`var v `, typName)
				g.decodeVarint("v", typName)
				g.P(`m.`, fieldname, ` = append(m.`, fieldname, `, v)`)
			} else {
				g.decodeVarint("m."+fieldname, typName)
			}
		} else {
			panic("Enum custom types is not supported!")
		}
	case descriptor.FieldDescriptorProto_TYPE_SFIXED32:
		if repeated {
			g.P(`var v int32`)
			g.decodeFixed32("v", "int32")
			g.P(`m.`, fieldname, ` = append(m.`, fieldname, `, `, fieldtype, `(v))`)
		} else {
			g.decodeFixed32("m."+fieldname, fieldtype)
		}
	case descriptor.FieldDescriptorProto_TYPE_SFIXED64:
		if repeated {
			g.P(`var v int64`)
			g.decodeFixed64("v", "int64")
			g.P(`m.`, fieldname, ` = append(m.`, fieldname, `, `, fieldtype, `(v))`)
		} else {
			g.decodeFixed64("m."+fieldname, fieldtype)
		}
	case descriptor.FieldDescriptorProto_TYPE_SINT32:
		g.P(`var v int32`)
		g.decodeVarint("v", "int32")
		g.P(`v = int32((uint32(v) >> 1) ^ uint32(((v&1)<<31)>>31))`)
		if repeated {
			g.P(`m.`, fieldname, ` = append(m.`, fieldname, `, `, fieldtype, `(v))`)
		} else {
			g.P(`m.`, fieldname, ` = `, fieldtype, `(v)`)
		}
	case descriptor.FieldDescriptorProto_TYPE_SINT64:
		g.P(`var v uint64`)
		g.decodeVarint("v", "uint64")
		g.P(`v = (v >> 1) ^ uint64((int64(v&1)<<63)>>63)`)
		if repeated {
			g.P(`m.`, fieldname, ` = append(m.`, fieldname, `, `, fieldtype, `(int64(v)))`)
		} else {
			g.P(`m.`, fieldname, ` = `, fieldtype, `(int64(v))`)
		}
	default:
		panic("not implemented")
	}
}

func (g *Generator) generateUnmarshal(file *FileDescriptor) {
	for _, message := range file.Messages() {
		ccTypeName := CamelCaseSlice(message.TypeName())

		g.P(`func (m *`, ccTypeName, `) Unmarshal(data []byte) error {`)
		g.In()
		g.P(`l := len(data)`)
		g.P(`index := 0`)
		g.P(`for index < l {`)
		g.In()
		g.P(`var wire uint64`)
		g.decodeVarint("wire", "uint64")
		g.P(`fieldNum := int32(wire >> 3)`)
		if len(message.Field) > 0 {
			g.P(`wireType := int(wire & 0x7)`)
		}
		g.P(`switch fieldNum {`)
		g.In()
		for _, field := range message.Field {
			fieldname := g.GetFieldName(message, field)

			packed := field.IsPacked()
			g.P(`case `, strconv.Itoa(int(field.GetNumber())), `:`)
			g.In()
			wireType := field.WireType()
			if packed {
				g.P(`if wireType == `, strconv.Itoa(proto.WireBytes), `{`)
				g.In()
				g.P(`var packedLen int`)
				g.decodeVarint("packedLen", "int")
				g.P(`postIndex := index + packedLen`)
				g.P(`if postIndex > l {`)
				g.In()
				g.P(`return `, g.Pkg["io"], `.ErrUnexpectedEOF`)
				g.Out()
				g.P(`}`)
				g.P(`for index < postIndex {`)
				g.In()
				g.field(field, fieldname)
				g.Out()
				g.P(`}`)
				g.Out()
				g.P(`} else if wireType == `, strconv.Itoa(wireType), `{`)
				g.In()
				g.field(field, fieldname)
				g.Out()
				g.P(`} else {`)
				g.In()
				g.P(`return ` + g.Pkg["fmt"] + `.Errorf("proto: wrong wireType = %d for field ` + fieldname + `", wireType)`)
				g.Out()
				g.P(`}`)
			} else {
				g.P(`if wireType != `, strconv.Itoa(wireType), `{`)
				g.In()
				g.P(`return ` + g.Pkg["fmt"] + `.Errorf("proto: wrong wireType = %d for field ` + fieldname + `", wireType)`)
				g.Out()
				g.P(`}`)
				g.field(field, fieldname)
			}
		}
		g.Out()
		g.P(`default:`)
		g.In()
		if message.DescriptorProto.HasExtension() {
			c := []string{}
			for _, erange := range message.GetExtensionRange() {
				c = append(c, `((fieldNum >= `+strconv.Itoa(int(erange.GetStart()))+") && (fieldNum<"+strconv.Itoa(int(erange.GetEnd()))+`))`)
			}
			g.P(`if `, strings.Join(c, "||"), `{`)
			g.In()
			g.P(`var sizeOfWire int`)
			g.P(`for {`)
			g.In()
			g.P(`sizeOfWire++`)
			g.P(`wire >>= 7`)
			g.P(`if wire == 0 {`)
			g.In()
			g.P(`break`)
			g.Out()
			g.P(`}`)
			g.Out()
			g.P(`}`)
			g.P(`index-=sizeOfWire`)
			g.P(`skippy, err := `, g.Pkg["proto"], `.Skip(data[index:])`)
			g.P(`if err != nil {`)
			g.In()
			g.P(`return err`)
			g.Out()
			g.P(`}`)
			g.P(`if (index + skippy) > l {`)
			g.In()
			g.P(`return `, g.Pkg["io"], `.ErrUnexpectedEOF`)
			g.Out()
			g.P(`}`)
			if gogoproto.HasExtensionsMap(file.FileDescriptorProto, message.DescriptorProto) {
				g.P(`if m.XXX_extensions == nil {`)
				g.In()
				g.P(`m.XXX_extensions = make(map[int32]`, g.Pkg["proto"], `.Extension)`)
				g.Out()
				g.P(`}`)
				g.P(`m.XXX_extensions[int32(fieldNum)] = `, g.Pkg["proto"], `.NewExtension(data[index:index+skippy])`)
			} else {
				g.P(`m.XXX_extensions = append(m.XXX_extensions, data[index:index+skippy]...)`)
			}
			g.P(`index += skippy`)
			g.Out()
			g.P(`} else {`)
			g.In()
		}
		g.P(`var sizeOfWire int`)
		g.P(`for {`)
		g.In()
		g.P(`sizeOfWire++`)
		g.P(`wire >>= 7`)
		g.P(`if wire == 0 {`)
		g.In()
		g.P(`break`)
		g.Out()
		g.P(`}`)
		g.Out()
		g.P(`}`)
		g.P(`index-=sizeOfWire`)
		g.P(`skippy, err := `, g.Pkg["proto"], `.Skip(data[index:])`)
		g.P(`if err != nil {`)
		g.In()
		g.P(`return err`)
		g.Out()
		g.P(`}`)
		g.P(`if (index + skippy) > l {`)
		g.In()
		g.P(`return `, g.Pkg["io"], `.ErrUnexpectedEOF`)
		g.Out()
		g.P(`}`)
		g.P(`m.XXX_unrecognized = append(m.XXX_unrecognized, data[index:index+skippy]...)`)
		g.P(`index += skippy`)
		g.Out()
		if message.DescriptorProto.HasExtension() {
			g.Out()
			g.P(`}`)
		}
		g.Out()
		g.P(`}`)
		g.Out()
		g.P(`}`)
		g.P(`return nil`)
		g.Out()
		g.P(`}`)
	}
}

func (g *Generator) skipFixed32() {
	g.P(`m.XXX_unrecognized = append(m.XXX_unrecognized, data[index:index+4]...)`)
	g.P(`index+=4`)
}

func (g *Generator) skipLengthDelimited() {
	g.P(`var length int`)
	g.P(`for shift := uint(0); ; shift += 7 {`)
	g.In()
	g.P(`if index >= l {`)
	g.In()
	g.P(`return `, g.Pkg["io"], `.ErrUnexpectedEOF`)
	g.Out()
	g.P(`}`)
	g.P(`b := data[index]`)
	g.P(`m.XXX_unrecognized = append(m.XXX_unrecognized, b)`)
	g.P(`index++`)
	g.P(`length |= (int(b) & 0x7F) << shift`)
	g.P(`if b < 0x80 {`)
	g.In()
	g.P(`break`)
	g.Out()
	g.P(`}`)
	g.Out()
	g.P(`}`)

	g.P(`m.XXX_unrecognized = append(m.XXX_unrecognized, data[index:index+length]...)`)
	g.P(`index+=length`)
}

func (g *Generator) skipFixed64() {
	g.P(`m.XXX_unrecognized = append(m.XXX_unrecognized, data[index:index+8]...)`)
	g.P(`index+=8`)
}

func (g *Generator) skipVarint() {
	g.P(`for {`)
	g.In()
	g.P(`if index >= l {`)
	g.In()
	g.P(`return `, g.Pkg["io"], `.ErrUnexpectedEOF`)
	g.Out()
	g.P(`}`)
	g.P(`m.XXX_unrecognized = append(m.XXX_unrecognized, data[index])`)
	g.P(`index++`)
	g.P(`if data[index-1] < 0x80 {`)
	g.In()
	g.P(`break`)
	g.Out()
	g.P(`}`)
	g.Out()
	g.P(`}`)
}
