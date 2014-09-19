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
The marshalto generates a Marshal and MarshalTo method for each message.
The `Marshal() ([]byte, error)` method results in the fact that the message
implements the Marshaler interface.
This allows proto.Marshal to be faster by calling the generated Marshal method rather than using reflect to Marshal the struct.

The generation of marshalling tests are enabled using one of the following extensions:

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

  message B {
	option (gogoproto.description) = true;
	optional string A = 1 [(gogoproto.embed) = true];
	repeated int64 G = 2 [(gogoproto.customtype) = "github.com/dropbox/goprotoc/test.Id"];
  }

the marshalto code will generate the following code:

	func (m *B) Marshal() (data []byte, err error) {
		size := m.Size()
		data = make([]byte, size)
		n, err := m.MarshalToUsingCachedSize(data)
		if err != nil {
			return nil, err
		}
		return data[:n], nil
	}

	func (m *B) MarshalTo(data []byte) (n int, err error) {
		m.Size()
		return m.MarshalToUsingCachedSize(data)
	}

	func (m *B) MarshalToUsingCachedSize(data []byte) (n int, err error) {
		var i int
		_ = i
		var l int
		_ = l
		if m.xxx_IsASet {
			data[i] = 0xa
			i++
			i = encodeVarintCustom(data, i, uint64(len(m.a)))
			i += copy(data[i:], m.a)
		}
		if m.xxx_LenG > 0 {
			for idx := 0; idx < m.xxx_LenG; idx++ {
				num := m.g[idx]
				data[i] = 0x10
				i++
				i = encodeVarintCustom(data, i, uint64(num))
			}
		}
		if m.XXX_unrecognized != nil {
			i += copy(data[i:], m.XXX_unrecognized)
		}
		return i, nil
	}

As shown above Marshal calculates the size of the not yet marshalled message
and allocates the appropriate buffer.
This is followed by calling the MarshalTo method which requires a preallocated buffer.
The MarshalTo method allows a user to rather preallocated a reusable buffer.

The Size method is generated using the size plugin and the gogoproto.sizer, gogoproto.sizer_all extensions.
The user can also using the generated Size method to check that his reusable buffer is still big enough.

The generated tests and benchmarks will keep you safe and show that this is really a significant speed improvement.

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

type NumGen interface {
	Next() string
	Current() string
}

type numGen struct {
	index int
}

func NewNumGen() NumGen {
	return &numGen{0}
}

func (this *numGen) Next() string {
	this.index++
	return this.Current()
}

func (this *numGen) Current() string {
	return strconv.Itoa(this.index)
}

func (g *Generator) callFixed64(varName ...string) {
	g.P(`i = encodeFixed64`, g.localName, `(data, i, uint64(`, strings.Join(varName, ""), `))`)
}

func (g *Generator) callFixed32(varName ...string) {
	g.P(`i = encodeFixed32`, g.localName, `(data, i, uint32(`, strings.Join(varName, ""), `))`)
}

func (g *Generator) callVarint(varName ...string) {
	g.P(`i = encodeVarint`, g.localName, `(data, i, uint64(`, strings.Join(varName, ""), `))`)
}

func (g *Generator) callInt32Varint(varName ...string) {
	g.P(`i = encodeVarint`, g.localName, `(data, i, uint64(uint32(`, strings.Join(varName, ""), `)))`)
}

func (g *Generator) encodeVarint(varName string) {
	g.P(`for `, varName, ` >= 1<<7 {`)
	g.In()
	g.P(`data[i] = uint8(uint64(`, varName, `)&0x7f|0x80)`)
	g.P(varName, ` >>= 7`)
	g.P(`i++`)
	g.Out()
	g.P(`}`)
	g.P(`data[i] = uint8(`, varName, `)`)
	g.P(`i++`)
}

func (g *Generator) encodeFixed64(varName string) {
	g.P(`data[i] = uint8(`, varName, `)`)
	g.P(`i++`)
	g.P(`data[i] = uint8(`, varName, ` >> 8)`)
	g.P(`i++`)
	g.P(`data[i] = uint8(`, varName, ` >> 16)`)
	g.P(`i++`)
	g.P(`data[i] = uint8(`, varName, ` >> 24)`)
	g.P(`i++`)
	g.P(`data[i] = uint8(`, varName, ` >> 32)`)
	g.P(`i++`)
	g.P(`data[i] = uint8(`, varName, ` >> 40)`)
	g.P(`i++`)
	g.P(`data[i] = uint8(`, varName, ` >> 48)`)
	g.P(`i++`)
	g.P(`data[i] = uint8(`, varName, ` >> 56)`)
	g.P(`i++`)
}

func (g *Generator) encodeFixed32(varName string) {
	g.P(`data[i] = uint8(`, varName, `)`)
	g.P(`i++`)
	g.P(`data[i] = uint8(`, varName, ` >> 8)`)
	g.P(`i++`)
	g.P(`data[i] = uint8(`, varName, ` >> 16)`)
	g.P(`i++`)
	g.P(`data[i] = uint8(`, varName, ` >> 24)`)
	g.P(`i++`)
}

func (g *Generator) encodeKey(fieldNumber int32, wireType int) {
	x := uint32(fieldNumber)<<3 | uint32(wireType)
	i := 0
	keybuf := make([]byte, 0)
	for i = 0; x > 127; i++ {
		keybuf = append(keybuf, 0x80|uint8(x&0x7F))
		x >>= 7
	}
	keybuf = append(keybuf, uint8(x))
	for _, b := range keybuf {
		g.P(`data[i] = `, fmt.Sprintf("%#v", b))
		g.P(`i++`)
	}
}

func (g *Generator) generateMarshalto(file *FileDescriptor) {
	numGen := NewNumGen()

	for _, message := range file.Messages() {
		ccTypeName := CamelCaseSlice(message.TypeName())

		g.P(`func (m *`, ccTypeName, `) Marshal() (data []byte, err error) {`)
		g.In()
		g.P(`size := m.Size()`)
		g.P(`data = make([]byte, size)`)
		g.P(`n, err := m.MarshalToUsingCachedSize(data)`)
		g.P(`if err != nil {`)
		g.In()
		g.P(`return nil, err`)
		g.Out()
		g.P(`}`)
		g.P(`return data[:n], nil`)
		g.Out()
		g.P(`}`)
		g.P(``)
		g.P(`func (m *`, ccTypeName, `) MarshalTo(data []byte) (n int, err error) {`)
		g.In()
		g.P(`m.Size()`)
		g.P(`return m.MarshalToUsingCachedSize(data)`)
		g.Out()
		g.P(`}`)
		g.P(``)
		g.P(`func (m *`, ccTypeName, `) MarshalToUsingCachedSize(data []byte) (n int, err error) {`)
		g.In()
		g.P(`var i int`)
		g.P(`_ = i`)
		g.P(`var l int`)
		g.P(`_ = l`)
		for _, field := range message.Field {
			fieldname := g.GetFieldName(message, field)
			repeated := field.IsRepeated()
			sizerName := ""
			if repeated {
				sizerName = SizerName(fieldname)
				g.P(`if m.`, sizerName, ` > 0 {`)
				g.In()
			} else {
				g.P(`if m.`, SetterName(fieldname), ` {`)
				g.In()
			}
			packed := field.IsPacked()
			wireType := field.WireType()
			fieldNumber := field.GetNumber()
			if packed {
				wireType = proto.WireBytes
			}
			switch *field.Type {
			case descriptor.FieldDescriptorProto_TYPE_DOUBLE:
				if packed {
					g.encodeKey(fieldNumber, wireType)
					g.callVarint(`m.`, sizerName, ` * 8`)
					g.P(`for idx := 0; idx < m.`, sizerName, `; idx++ {`)
					g.In()
					g.P(`num := m.`, fieldname, `[idx]`)
					g.P(`f`, numGen.Next(), ` := `, g.Pkg["math"], `.Float64bits(float64(num))`)
					g.encodeFixed64("f" + numGen.Current())
					g.Out()
					g.P(`}`)
				} else if repeated {
					g.P(`for idx := 0; idx < m.`, sizerName, `; idx++ {`)
					g.In()
					g.P(`num := m.`, fieldname, `[idx]`)
					g.encodeKey(fieldNumber, wireType)
					g.P(`f`, numGen.Next(), ` := `, g.Pkg["math"], `.Float64bits(float64(num))`)
					g.encodeFixed64("f" + numGen.Current())
					g.Out()
					g.P(`}`)
				} else {
					g.encodeKey(fieldNumber, wireType)
					g.callFixed64(g.Pkg["math"], `.Float64bits(float64(m.`+fieldname, `))`)
				}
			case descriptor.FieldDescriptorProto_TYPE_FLOAT:
				if packed {
					g.encodeKey(fieldNumber, wireType)
					g.callVarint(`m.`, sizerName, ` * 4`)
					g.P(`for idx := 0; idx < m.`, sizerName, `; idx++ {`)
					g.In()
					g.P(`num := m.`, fieldname, `[idx]`)
					g.P(`f`, numGen.Next(), ` := `, g.Pkg["math"], `.Float32bits(float32(num))`)
					g.encodeFixed32("f" + numGen.Current())
					g.Out()
					g.P(`}`)
				} else if repeated {
					g.P(`for idx := 0; idx < m.`, sizerName, `; idx++ {`)
					g.In()
					g.P(`num := m.`, fieldname, `[idx]`)
					g.encodeKey(fieldNumber, wireType)
					g.P(`f`, numGen.Next(), ` := `, g.Pkg["math"], `.Float32bits(float32(num))`)
					g.encodeFixed32("f" + numGen.Current())
					g.Out()
					g.P(`}`)
				} else {
					g.encodeKey(fieldNumber, wireType)
					g.callFixed32(g.Pkg["math"], `.Float32bits(float32(m.`+fieldname, `))`)
				}
			case descriptor.FieldDescriptorProto_TYPE_INT64,
				descriptor.FieldDescriptorProto_TYPE_UINT64,
				descriptor.FieldDescriptorProto_TYPE_INT32,
				descriptor.FieldDescriptorProto_TYPE_UINT32,
				descriptor.FieldDescriptorProto_TYPE_ENUM:
				if packed {
					jvar := "j" + numGen.Next()
					g.P(`data`, numGen.Next(), ` := make([]byte, m.`, sizerName, `*10)`)
					g.P(`var `, jvar, ` int`)
					if *field.Type == descriptor.FieldDescriptorProto_TYPE_INT32 {
						g.P(`for idx := 0; idx < m.`, sizerName, `; idx++ {`)
						g.In()
						g.P(`num := uint32(m.`, fieldname, `[idx])`)
					} else if *field.Type == descriptor.FieldDescriptorProto_TYPE_INT64 {
						g.P(`for idx := 0; idx < m.`, sizerName, `; idx++ {`)
						g.In()
						g.P(`num := uint64(m.`, fieldname, `[idx])`)
					} else {
						g.P(`for idx := 0; idx < m.`, sizerName, `; idx++ {`)
						g.In()
						g.P(`num := m.`, fieldname, `[idx]`)
					}
					g.P(`for num >= 1<<7 {`)
					g.In()
					g.P(`data`, numGen.Current(), `[`, jvar, `] = uint8(uint64(num)&0x7f|0x80)`)
					g.P(`num >>= 7`)
					g.P(jvar, `++`)
					g.Out()
					g.P(`}`)
					g.P(`data`, numGen.Current(), `[`, jvar, `] = uint8(num)`)
					g.P(jvar, `++`)
					g.Out()
					g.P(`}`)
					g.encodeKey(fieldNumber, wireType)
					g.callVarint(jvar)
					g.P(`i += copy(data[i:], data`, numGen.Current(), `[:`, jvar, `])`)
				} else if repeated {
					g.P(`for idx := 0; idx < m.`, sizerName, `; idx++ {`)
					g.In()
					g.P(`num := m.`, fieldname, `[idx]`)
					g.encodeKey(fieldNumber, wireType)
					if *field.Type == descriptor.FieldDescriptorProto_TYPE_INT32 {
						g.callInt32Varint("num")
					} else {
						g.callVarint("num")
					}
					g.Out()
					g.P(`}`)
				} else {
					g.encodeKey(fieldNumber, wireType)
					if *field.Type == descriptor.FieldDescriptorProto_TYPE_INT32 {
						g.callInt32Varint(`m.`, fieldname)
					} else {
						g.callVarint(`m.`, fieldname)
					}
				}
			case descriptor.FieldDescriptorProto_TYPE_FIXED64,
				descriptor.FieldDescriptorProto_TYPE_SFIXED64:
				if packed {
					g.encodeKey(fieldNumber, wireType)
					g.callVarint(`m.`, sizerName, ` * 8`)
					g.P(`for idx := 0; idx < m.`, sizerName, `; idx++ {`)
					g.In()
					g.P(`num := m.`, fieldname, `[idx]`)
					g.encodeFixed64("num")
					g.Out()
					g.P(`}`)
				} else if repeated {
					g.P(`for idx := 0; idx < m.`, sizerName, `; idx++ {`)
					g.In()
					g.P(`num := m.`, fieldname, `[idx]`)
					g.encodeKey(fieldNumber, wireType)
					g.encodeFixed64("num")
					g.Out()
					g.P(`}`)
				} else {
					g.encodeKey(fieldNumber, wireType)
					g.callFixed64("m." + fieldname)
				}
			case descriptor.FieldDescriptorProto_TYPE_FIXED32,
				descriptor.FieldDescriptorProto_TYPE_SFIXED32:
				if packed {
					g.encodeKey(fieldNumber, wireType)
					g.callVarint(`m.`, sizerName, ` * 4`)
					g.P(`for idx := 0; idx < m.`, sizerName, `; idx++ {`)
					g.In()
					g.P(`num := m.`, fieldname, `[idx]`)
					g.encodeFixed32("num")
					g.Out()
					g.P(`}`)
				} else if repeated {
					g.P(`for idx := 0; idx < m.`, sizerName, `; idx++ {`)
					g.In()
					g.P(`num := m.`, fieldname, `[idx]`)
					g.encodeKey(fieldNumber, wireType)
					g.encodeFixed32("num")
					g.Out()
					g.P(`}`)
				} else {
					g.encodeKey(fieldNumber, wireType)
					g.callFixed32("m." + fieldname)
				}
			case descriptor.FieldDescriptorProto_TYPE_BOOL:
				if packed {
					g.encodeKey(fieldNumber, wireType)
					g.callVarint(`m.`, sizerName, ``)
					g.P(`for idx := 0; idx < m.`, sizerName, `; idx++ {`)
					g.In()
					g.P(`b := m.`, fieldname, `[idx]`)
					g.P(`if b {`)
					g.In()
					g.P(`data[i] = 1`)
					g.Out()
					g.P(`} else {`)
					g.In()
					g.P(`data[i] = 0`)
					g.Out()
					g.P(`}`)
					g.P(`i++`)
					g.Out()
					g.P(`}`)
				} else if repeated {
					g.P(`for idx := 0; idx < m.`, sizerName, `; idx++ {`)
					g.In()
					g.P(`b := m.`, fieldname, `[idx]`)
					g.encodeKey(fieldNumber, wireType)
					g.P(`if b {`)
					g.In()
					g.P(`data[i] = 1`)
					g.Out()
					g.P(`} else {`)
					g.In()
					g.P(`data[i] = 0`)
					g.Out()
					g.P(`}`)
					g.P(`i++`)
					g.Out()
					g.P(`}`)
				} else {
					g.encodeKey(fieldNumber, wireType)
					g.P(`if m.`, fieldname, ` {`)
					g.In()
					g.P(`data[i] = 1`)
					g.Out()
					g.P(`} else {`)
					g.In()
					g.P(`data[i] = 0`)
					g.Out()
					g.P(`}`)
					g.P(`i++`)
				}
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				if repeated {
					g.P(`for idx := 0; idx < m.`, sizerName, `; idx++ {`)
					g.In()
					g.P(`s := m.`, fieldname, `[idx]`)
					g.encodeKey(fieldNumber, wireType)
					g.P(`l = len(s)`)
					g.encodeVarint("l")
					g.P(`i+=copy(data[i:], s)`)
					g.Out()
					g.P(`}`)
				} else {
					g.encodeKey(fieldNumber, wireType)
					g.callVarint(`len(m.`, fieldname, `)`)
					g.P(`i+=copy(data[i:], m.`, fieldname, `)`)
				}
			case descriptor.FieldDescriptorProto_TYPE_GROUP:
				panic(fmt.Errorf("marshaler does not support group %v", fieldname))
			case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
				if repeated {
					g.P(`for idx := 0; idx < m.`, sizerName, `; idx++ {`)
					g.In()
					g.P(`msg := m.`, fieldname, `[idx]`)
					g.encodeKey(fieldNumber, wireType)
					g.callVarint("msg.SizeCached()")
					g.P(`n, err := msg.MarshalToUsingCachedSize(data[i:])`)
					g.P(`if err != nil {`)
					g.In()
					g.P(`return 0, err`)
					g.Out()
					g.P(`}`)
					g.P(`i+=n`)
					g.Out()
					g.P(`}`)
				} else {
					g.encodeKey(fieldNumber, wireType)
					g.callVarint(`m.`, fieldname, `.SizeCached()`)
					g.P(`n`, numGen.Next(), `, err := m.`, fieldname, `.MarshalToUsingCachedSize(data[i:])`)
					g.P(`if err != nil {`)
					g.In()
					g.P(`return 0, err`)
					g.Out()
					g.P(`}`)
					g.P(`i+=n`, numGen.Current())
				}
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				if repeated {
					g.P(`for idx := 0; idx < m.`, sizerName, `; idx++ {`)
					g.In()
					g.P(`b := m.`, fieldname, `[idx]`)
					g.encodeKey(fieldNumber, wireType)
					g.callVarint("len(b)")
					g.P(`i+=copy(data[i:], b)`)
					g.Out()
					g.P(`}`)
				} else {
					g.encodeKey(fieldNumber, wireType)
					g.callVarint(`len(m.`, fieldname, `)`)
					g.P(`i+=copy(data[i:], m.`, fieldname, `)`)
				}
			case descriptor.FieldDescriptorProto_TYPE_SINT32:
				if packed {
					datavar := "data" + numGen.Next()
					jvar := "j" + numGen.Next()
					g.P(datavar, ` := make([]byte, m.`, sizerName, "*5)")
					g.P(`var `, jvar, ` int`)
					g.P(`for idx := 0; idx < m.`, sizerName, `; idx++ {`)
					g.In()
					g.P(`num := m.`, fieldname, `[idx]`)
					xvar := "x" + numGen.Next()
					g.P(xvar, ` := (uint32(num) << 1) ^ uint32((num >> 31))`)
					g.P(`for `, xvar, ` >= 1<<7 {`)
					g.In()
					g.P(datavar, `[`, jvar, `] = uint8(uint64(`, xvar, `)&0x7f|0x80)`)
					g.P(jvar, `++`)
					g.P(xvar, ` >>= 7`)
					g.Out()
					g.P(`}`)
					g.P(datavar, `[`, jvar, `] = uint8(`, xvar, `)`)
					g.P(jvar, `++`)
					g.Out()
					g.P(`}`)
					g.encodeKey(fieldNumber, wireType)
					g.callVarint(jvar)
					g.P(`i+=copy(data[i:], `, datavar, `[:`, jvar, `])`)
				} else if repeated {
					g.P(`for idx := 0; idx < m.`, sizerName, `; idx++ {`)
					g.In()
					g.P(`num := m.`, fieldname, `[idx]`)
					g.encodeKey(fieldNumber, wireType)
					g.P(`x`, numGen.Next(), ` := (uint32(num) << 1) ^ uint32((num >> 31))`)
					g.encodeVarint("x" + numGen.Current())
					g.Out()
					g.P(`}`)
				} else {
					g.encodeKey(fieldNumber, wireType)
					g.callVarint(`(uint32(m.`, fieldname, `) << 1) ^ uint32((m.`, fieldname, ` >> 31))`)
				}
			case descriptor.FieldDescriptorProto_TYPE_SINT64:
				if packed {
					jvar := "j" + numGen.Next()
					xvar := "x" + numGen.Next()
					datavar := "data" + numGen.Next()
					g.P(`var `, jvar, ` int`)
					g.P(datavar, ` := make([]byte, m.`, SizerName(fieldname), `*10)`)
					g.P(`for idx := 0; idx < m.`, sizerName, `; idx++ {`)
					g.In()
					g.P(`num := m.`, fieldname, `[idx]`)
					g.P(xvar, ` := (uint64(num) << 1) ^ uint64((num >> 63))`)
					g.P(`for `, xvar, ` >= 1<<7 {`)
					g.In()
					g.P(datavar, `[`, jvar, `] = uint8(uint64(`, xvar, `)&0x7f|0x80)`)
					g.P(jvar, `++`)
					g.P(xvar, ` >>= 7`)
					g.Out()
					g.P(`}`)
					g.P(datavar, `[`, jvar, `] = uint8(`, xvar, `)`)
					g.P(jvar, `++`)
					g.Out()
					g.P(`}`)
					g.encodeKey(fieldNumber, wireType)
					g.callVarint(jvar)
					g.P(`i+=copy(data[i:], `, datavar, `[:`, jvar, `])`)
				} else if repeated {
					g.P(`for idx := 0; idx < m.`, sizerName, `; idx++ {`)
					g.In()
					g.P(`num := m.`, fieldname, `[idx]`)
					g.encodeKey(fieldNumber, wireType)
					g.P(`x`, numGen.Next(), ` := (uint64(num) << 1) ^ uint64((num >> 63))`)
					g.encodeVarint("x" + numGen.Current())
					g.Out()
					g.P(`}`)
				} else {
					g.encodeKey(fieldNumber, wireType)
					g.callVarint(`(uint64(m.`, fieldname, `) << 1) ^ uint64((m.`, fieldname, ` >> 63))`)
				}
			default:
				panic("not implemented")
			}
			g.Out()
			g.P(`}`)
		}
		if message.DescriptorProto.HasExtension() {
			if gogoproto.HasExtensionsMap(file.FileDescriptorProto, message.DescriptorProto) {
				g.P(`if len(m.XXX_extensions) > 0 {`)
				g.In()
				g.P(`n, err := `, g.Pkg["proto"], `.EncodeExtensionMap(m.XXX_extensions, data[i:])`)
				g.P(`if err != nil {`)
				g.In()
				g.P(`return 0, err`)
				g.Out()
				g.P(`}`)
				g.P(`i+=n`)
				g.Out()
				g.P(`}`)
			} else {
				g.P(`if m.XXX_extensions != nil {`)
				g.In()
				g.P(`i+=copy(data[i:], m.XXX_extensions)`)
				g.Out()
				g.P(`}`)
			}
		}
		g.P(`if m.XXX_unrecognized != nil {`)
		g.In()
		g.P(`i+=copy(data[i:], m.XXX_unrecognized)`)
		g.Out()
		g.P(`}`)
		g.P(`return i, nil`)
		g.Out()
		g.P(`}`)
	}

	g.P(`func encodeFixed64`, g.localName, `(data []byte, offset int, v uint64) int {`)
	g.In()
	g.P(`data[offset] = uint8(v)`)
	g.P(`data[offset+1] = uint8(v >> 8)`)
	g.P(`data[offset+2] = uint8(v >> 16)`)
	g.P(`data[offset+3] = uint8(v >> 24)`)
	g.P(`data[offset+4] = uint8(v >> 32)`)
	g.P(`data[offset+5] = uint8(v >> 40)`)
	g.P(`data[offset+6] = uint8(v >> 48)`)
	g.P(`data[offset+7] = uint8(v >> 56)`)
	g.P(`return offset+8`)
	g.Out()
	g.P(`}`)

	g.P(`func encodeFixed32`, g.localName, `(data []byte, offset int, v uint32) int {`)
	g.In()
	g.P(`data[offset] = uint8(v)`)
	g.P(`data[offset+1] = uint8(v >> 8)`)
	g.P(`data[offset+2] = uint8(v >> 16)`)
	g.P(`data[offset+3] = uint8(v >> 24)`)
	g.P(`return offset+4`)
	g.Out()
	g.P(`}`)

	g.P(`func encodeVarint`, g.localName, `(data []byte, offset int, v uint64) int {`)
	g.In()
	g.P(`for v >= 1<<7 {`)
	g.In()
	g.P(`data[offset] = uint8(v&0x7f|0x80)`)
	g.P(`v >>= 7`)
	g.P(`offset++`)
	g.Out()
	g.P(`}`)
	g.P(`data[offset] = uint8(v)`)
	g.P(`return offset+1`)
	g.Out()
	g.P(`}`)
}
