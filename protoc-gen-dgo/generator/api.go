// Copyright (c) 2014, Dropbox INC. All rights reserved.
// www.dropbox.com
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
// `AS IS` AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
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
The api function generates accessor methods in respect with the c++ api
documentation: https://developers.google.com/protocol-buffers/docs/reference/cpp-generated

The generation of api tests are enabled using one of the following extensions:

  - testgen
  - testgen_all

Given the following message:

message TestSingular {
    optional TestMessage msgs = 1;
    optional double vald = 2;
    optional float valf = 3;
}

the api code will generate the following code:

func (m *TestSingular) MutateMsgs() (field *TestMessage, err error) {
    if m == nil {
        return nil, errors.New("Cannot mutate a nil message")
    }
    if !m.xxx_IsMsgsSet {
        m.xxx_IsMsgsSet = true
        m.msgs = new(TestMessage)
    }
    return m.msgs, nil
}

func (m *TestSingular) HasMsgs() (isSet bool) {
    if m != nil && m.xxx_IsMsgsSet {
        return true
    }
    return false
}

func (m *TestSingular) ClearMsgs() {
    if m != nil {
        m.msgs.Clear()
        m.xxx_IsMsgsSet = false

    }
}

func (m *TestSingular) SetVald(value float64) (err error) {
    if m == nil {
        return errors.New("Cannot assign to nil message")
    }
    m.xxx_IsValdSet = true
    m.vald = value
    return nil
}

func (m *TestSingular) HasVald() (isSet bool) {
    if m != nil && m.xxx_IsValdSet {
        return true
    }
    return false
}

func (m *TestSingular) ClearVald() {
    if m != nil {
        m.xxx_IsValdSet = false
    }
}

func (m *TestSingular) SetValf(value float32) (err error) {
    if m == nil {
        return errors.New("Cannot assign to nil message")
    }
    m.xxx_IsValfSet = true
    m.valf = value
    return nil
}

func (m *TestSingular) HasValf() (isSet bool) {
    if m != nil && m.xxx_IsValfSet {
        return true
    }
    return false
}

func (m *TestSingular) ClearValf() {
    if m != nil {
        m.xxx_IsValfSet = false
    }
}

func (m *TestSingular) Clear() {
    if m != nil {
        m.msgs.Clear()
        m.xxx_IsMsgsSet = false

        m.ClearVald()
        m.ClearValf()
    }
}
*/

package generator

import (
    "strings"

    descriptor "github.com/dropbox/goprotoc/protoc-gen-dgo/descriptor"
)

const expResizeThreshold string = "1000000"

type fieldNames struct {
    typeName      string
    fieldName     string
    fieldType     string
    fieldTypeBase string
    field         *descriptor.FieldDescriptorProto
    message       *Descriptor
    protoType     descriptor.FieldDescriptorProto_Type
}

func (g *Generator) generateAPI(message *Descriptor) {
    c := new(fieldNames)
    c.typeName = CamelCaseSlice(message.TypeName())
    c.message = message
    g.P(`func (m *`, c.typeName, `) SizeCached() int {`)
    g.In()
    g.P(`return m.xxx_sizeCached`)
    g.Out()
    g.P(`}`)
    g.P(``)
    for _, field := range message.Field {
        isMessageField := IsMessageType(field)
        c.field = field
        c.protoType = *field.Type
        c.fieldName = g.GetFieldName(message, field)
        c.fieldType, _ = g.GoType(message, field)

        switch field.GetLabel() {
        case descriptor.FieldDescriptorProto_LABEL_OPTIONAL,
            descriptor.FieldDescriptorProto_LABEL_REQUIRED:
            c.fieldTypeBase = strings.Replace(c.fieldType, "*", "", -1)
            if isMessageField {
                g.genMutateSingular(c)
            } else {
                g.genSetSingular(c)
            }
            g.genHas(c)
            g.genClear(c)
        case descriptor.FieldDescriptorProto_LABEL_REPEATED:
            c.fieldTypeBase = strings.Replace(strings.Replace(c.fieldType, "*", "", 1),
                "[]", "", 1)
            if isMessageField {
                g.genAddMessage(c)
                g.genMutateMessage(c)
            } else {
                g.genAddScalar(c)
                g.genSetScalar(c)
            }
            g.genSize(c)
            g.genClear(c)
            g.genGetByIndex(c)
        default:
            panic("not implemented")
        }
    }
    g.genClearAll(message)
}

// Returns the number of elements currently in the field
func (g *Generator) genSize(c *fieldNames) {
    sizerName := SizerName(c.fieldName)
    g.P(`func (m *`, c.typeName, `) `, CamelCase(c.fieldName), `Size() (size int) {`)
    g.In()
    g.P(`if m != nil {`)
    g.In()
    g.P(`return m.`, sizerName)
    g.Out()
    g.P(`}`)
    g.P(`return 0`)
    g.Out()
    g.P(`}`)
    g.P()
}

// Returns true if the field is set. //
func (g *Generator) genHas(c *fieldNames) {
    g.P(`func (m *`, c.typeName, `) Has`, CamelCase(c.fieldName), `() (isSet bool) {`)
    g.In()
    g.P(`if m != nil && m.`, SetterName(c.fieldName), ` {`)
    g.In()
    g.P(`return true`)
    g.Out()
    g.P(`}`)
    g.P(`return false`)
    g.Out()
    g.P(`}`)
    g.P()
}

func (g *Generator) genSmartResize(c *fieldNames, pointer string) {
    sizerName := SizerName(c.fieldName)
    g.P(`if len(m.`, c.fieldName, `) <= m.`, sizerName, ` {`)
    g.In()
    g.P(`newCapacity := 0`)
    g.P(`if len(m.`, c.fieldName, `) == 0 {`)
    g.In()
    g.P(`newCapacity = 8`)
    g.Out()
    g.P(`} else if len(m.`, c.fieldName, `) < `, expResizeThreshold, ` {`)
    g.In()
    g.P(`newCapacity = m.`, sizerName, `*2`)
    g.Out()
    g.P(`} else {`)
    g.In()
    g.P(`newCapacity = m.`, sizerName, ` + `, expResizeThreshold)
    g.Out()
    g.P(`}`)
    g.P(`t := make([]`, pointer, c.fieldTypeBase, `, newCapacity, newCapacity)`)
    g.P(`copy(t, m.`, c.fieldName, `)`)
    g.P(`m.`, c.fieldName, ` = t`)
    g.Out()
    g.P(`}`)
}

// Appends a new element to the field with the given value.
func (g *Generator) genAddScalar(c *fieldNames) {
    sizerName := SizerName(c.fieldName)
    g.P(`func (m *`, c.typeName, `) Add`, CamelCase(c.fieldName),
        `(value `, c.fieldTypeBase, `) (err error) {`)
    g.In()
    g.P(`if m == nil {`)
    g.In()
    g.P(`return `, g.Pkg[`errors`], `.New("Cannot append to nil message")`)
    g.Out()
    g.P(`}`)
    if c.fieldTypeBase == "[]byte" {
        g.P(`if value == nil {`)
        g.In()
        g.P(`return `, g.Pkg[`errors`], `.New("Cannot set with a nil value.")`)
        g.Out()
        g.P(`}`)
    }
    g.genSmartResize(c, "")
    g.P(`m.`, c.fieldName, `[m.`, sizerName, `] = value`)
    g.P(`m.`, sizerName, ` += 1`)
    g.P(`return nil`)
    g.Out()
    g.P(`}`)
    g.P()
}

// Appends a new element to the field with the given value.
func (g *Generator) genAddMessage(c *fieldNames) {
    pointer := getAssignmentPointer(c.fieldType)
    sizerName := SizerName(c.fieldName)
    g.P(`func (m *`, c.typeName, `) Add`, CamelCase(c.fieldName),
        `() (field *`, c.fieldTypeBase, `, err error) {`)
    g.In()
    g.P(`if m != nil {`)
    g.In()
    g.P(`field = new(`, c.fieldTypeBase, `)`)
    g.genSmartResize(c, "*")
    g.P(`m.`, c.fieldName, `[m.`, sizerName, `] = `, pointer, `field`)
    g.P(`m.`, sizerName, ` += 1`)
    g.P(`return field, nil`)
    g.Out()
    g.P(`}`)
    g.P(`return nil, `, g.Pkg[`errors`], `.New("Cannot append to nil message")`)
    g.Out()
    g.P(`}`)
    g.P()
}

// Sets the value of the element at the given zero-based index.
func (g *Generator) genSetScalar(c *fieldNames) {
    ref := getRefrence(c.fieldType)
    g.P(`func (m *`, c.typeName, `) Set`, CamelCase(c.fieldName),
        `(value `, c.fieldTypeBase, `, index int) (err error) {`)
    g.In()
    g.P(`if m == nil {`)
    g.In()
    g.P(`return `, g.Pkg[`errors`], `.New("Cannot assign to nil message")`)
    g.Out()
    g.P(`}`)
    g.P(`if index < 0 || index >= m.`, SizerName(c.fieldName), ` {`)
    g.In()
    g.P(`return `, g.Pkg[`errors`], `.New("Index is out of bounds")`)
    g.Out()
    g.P(`}`)
    if c.fieldTypeBase == "[]byte" {
        g.P(`if value == nil {`)
        g.In()
        g.P(`return `, g.Pkg[`errors`], `.New("Cannot set with a nil value.")`)
        g.Out()
        g.P(`}`)
    }
    g.P(`m.`, c.fieldName, `[index] = `, ref, `value`)
    g.P(`return nil`)
    g.Out()
    g.P(`}`)
    g.P()
}

// Sets the value of the element at the given zero-based index.
func (g *Generator) genMutateMessage(c *fieldNames) {
    notref := getAssignmentRefrence(c.fieldType)
    g.P(`func (m *`, c.typeName, `) Mutate`, CamelCase(c.fieldName),
        `(index int) (field *`, c.fieldTypeBase, `, err error) {`)
    g.In()
    g.P(`if m == nil {`)
    g.In()
    g.P(`return nil, `, g.Pkg[`errors`], `.New("Cannot mutate a nil message")`)
    g.Out()
    g.P(`}`)
    g.P(`if index < 0 || index >= m.`, SizerName(c.fieldName), ` {`)
    g.In()
    g.P(`return nil, `, g.Pkg[`errors`], `.New("Index is out of bounds")`)
    g.Out()
    g.P(`}`)
    g.P(`if m.`, c.fieldName, `[index] == nil {`)
    g.In()
    g.P(`m.`, c.fieldName, `[index] = new(`, c.fieldTypeBase, `)`)
    g.Out()
    g.P(`}`)
    g.P(`return `, notref, `m.`, c.fieldName, `[index], nil`)
    g.Out()
    g.P(`}`)
    g.P()
}

// Sets the value of the non-repeated element.
func (g *Generator) genSetSingular(c *fieldNames) {
    ref := getRefrence(c.fieldType)
    g.P(`func (m *`, c.typeName, `) Set`, CamelCase(c.fieldName),
        `(value `, c.fieldTypeBase, `) (err error) {`)
    g.In()
    g.P(`if m == nil {`)
    g.In()
    g.P(`return `, g.Pkg[`errors`], `.New("Cannot assign to nil message")`)
    g.Out()
    g.P(`}`)
    if c.fieldType == "[]byte" {
        g.P(`if value == nil {`)
        g.In()
        g.P(`return `, g.Pkg[`errors`], `.New("Cannot set with a nil value.")`)
        g.Out()
        g.P(`}`)
    }
    g.P(`m.`, SetterName(c.fieldName), ` = true`)
    g.P(`m.`, c.fieldName, ` = `, ref, `value`)
    g.P(`return nil`)
    g.Out()
    g.P(`}`)
    g.P()
}

// Mutates the value of the non-repeated element.
func (g *Generator) genMutateSingular(c *fieldNames) {
    notref := getAssignmentRefrence(c.fieldType)
    setterName := SetterName(c.fieldName)
    g.P(`func (m *`, c.typeName, `) Mutate`, CamelCase(c.fieldName),
        `() (field *`, c.fieldTypeBase, `, err error) {`)
    g.In()
    g.P(`if m == nil {`)
    g.In()
    g.P(`return nil, `, g.Pkg[`errors`], `.New("Cannot mutate a nil message")`)
    g.Out()
    g.P(`}`)
    g.P(`if !m.`, setterName, ` {`)
    g.In()
    g.P(`m.`, setterName, ` = true`)
    g.P(`m.`, c.fieldName, ` = new(`, c.fieldTypeBase, `)`)
    g.Out()
    g.P(`}`)
    g.P(`return `, notref, `m.`, c.fieldName, `, nil`)
    g.Out()
    g.P(`}`)
    g.P()
}

func (g *Generator) genMsgClear(message *Descriptor, field *descriptor.FieldDescriptorProto) {
    fieldName := g.GetFieldName(message, field)
    if IsRepeated(field) {
        g.P(`for i := 0; i < m.`, CamelCase(fieldName), `Size(); i ++ {`)
        g.In()
        g.P(`m.`, fieldName, `[i].Clear()`)
        g.Out()
        g.P(`}`)
    } else {
        g.P(`m.`, fieldName, `.Clear()`)
    }
    if IsRepeated(field) {
        g.P(`m.`, SizerName(fieldName), ` = 0`)
    } else {
        g.P(`m.`, SetterName(fieldName), ` = false`)
    }
    g.P()
}

// Removes all elements from the field. After calling this, foo_size() will return zero.
func (g *Generator) genClear(c *fieldNames) {
    g.P(`func (m *`, c.typeName, `) Clear`, CamelCase(c.fieldName), `() {`)
    g.In()
    g.P(`if m != nil {`)
    g.In()
    if IsMessageType(c.field) {
        g.genMsgClear(c.message, c.field)
    } else {
        if IsRepeated(c.field) {
            g.P(`m.`, SizerName(c.fieldName), ` = 0`)
        } else {
            g.P(`m.`, SetterName(c.fieldName), ` = false`)
        }
        if c.fieldType == "string" {
            g.P(`m.`, c.fieldName, ` = ""`)
        } else if c.fieldType == "[]byte" {
            g.P(`m.`, c.fieldName, ` = nil`)
        }
    }
    g.Out()
    g.P(`}`)
    g.Out()
    g.P(`}`)
    g.P()
}

// Removes all fields from the message.
func (g *Generator) genClearAll(message *Descriptor) {
    typeName := CamelCaseSlice(message.TypeName())
    g.P(`func (m *`, typeName, `) Clear() {`)
    g.In()
    g.P(`if m != nil {`)
    g.In()
    for _, field := range message.Field {
        if IsMessageType(field) {
            g.genMsgClear(message, field)
        } else {
            g.P(`m.Clear`, CamelCase(g.GetFieldName(message, field)), `()`)
        }
    }
    g.Out()
    g.P(`}`)
    g.Out()
    g.P(`}`)
    g.P()
}

// Returns the element at the given zero-based index.
func (g *Generator) genGetByIndex(c *fieldNames) {
    defaultValue := GetDefaultValue(c.protoType)
    pointer := ""
    if IsMessageType(c.field) {
        pointer = "*"
    }
    g.P(`func (m *`, c.typeName, `) Get`, CamelCase(c.fieldName),
        `(index int) (field `, pointer, c.fieldTypeBase, `, err error) {`)
    g.In()
    g.P(`if m == nil {`)
    g.In()
    g.P(`return `, defaultValue, `, `, g.Pkg[`errors`], `.New("Cannot get nil message")`)
    g.Out()
    g.P(`}`)
    g.P(`if index < 0 || index >= m.`, SizerName(c.fieldName), ` {`)
    g.In()
    g.P(`return `, defaultValue, `, `, g.Pkg[`errors`], `.New("Index is out of bounds")`)
    g.Out()
    g.P(`}`)
    g.P(`return m.`, c.fieldName, `[index], nil`)
    g.Out()
    g.P(`}`)
    g.P()
}

func getRefrence(fieldType string) string {
    if strings.Contains(fieldType, "*") {
        return "&"
    } else {
        return ""
    }
}

func getAssignmentPointer(fieldType string) string {
    if strings.Contains(fieldType, "*") {
        return ""
    } else {
        return "*"
    }
}

func getAssignmentRefrence(fieldType string) string {
    if strings.Contains(fieldType, "*") {
        return ""
    } else {
        return "&"
    }
}
