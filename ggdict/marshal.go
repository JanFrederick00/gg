// Copyright 2020 Frederik Zipp. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ggdict

import (
	"sort"
	"strconv"
)

func Marshal(dict map[string]any, shortStringIndices bool) []byte {
	m := newMarshaller(shortStringIndices)
	m.writeRawInt32(formatSignature)
	m.writeRawInt32(1)
	m.writeRawInt32(0)
	m.writeValue(dict)
	m.writeKeys()
	return m.buf
}

type marshaller struct {
	buf                []byte
	offset             int
	keys               []string
	keyIndex           map[string]int
	shortStringIndices bool
}

func newMarshaller(shortStringIndices bool) *marshaller {
	return &marshaller{
		keyIndex:           make(map[string]int),
		shortStringIndices: shortStringIndices,
	}
}

func (m *marshaller) writeValue(value any) {
	switch v := value.(type) {
	case nil:
		m.writeNull()
	case map[string]any:
		m.writeDictionary(v)
	case []any:
		m.writeArray(v)
	case string:
		m.writeString(v)
	case int:
		m.writeInteger(v)
	case int32:
		m.writeInteger(int(v))
	case int64:
		m.writeInteger(int(v))
	case uint32:
		m.writeInteger(int(v))
	case uint64:
		m.writeInteger(int(v))
	case float64:
		m.writeFloat(v)
	case float32:
		m.writeFloat(float64(v))
	}
}

func (m *marshaller) writeTypeMarker(t valueType) {
	m.writeByte(byte(t))
}

func (m *marshaller) writeNull() {
	m.writeTypeMarker(typeNull)
}

func (m *marshaller) writeDictionary(d map[string]any) {
	// sorted keys for reproducible results
	keys := make([]string, 0, len(d))
	for k := range d {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	m.writeTypeMarker(typeDictionary)
	m.writeRawInt32(len(d))
	for _, k := range keys {
		m.writeKeyIndex(k)
		m.writeValue(d[k])
	}
	m.writeTypeMarker(typeDictionary)
}

func (m *marshaller) writeArray(a []any) {
	m.writeTypeMarker(typeArray)
	m.writeRawInt32(len(a))
	for _, v := range a {
		m.writeValue(v)
	}
	m.writeTypeMarker(typeArray)
}

func (m *marshaller) writeString(s string) {
	m.writeTypeMarker(typeString)
	m.writeKeyIndex(s)
}

func (m *marshaller) writeInteger(i int) {
	m.writeTypeMarker(typeInteger)
	m.writeKeyIndex(strconv.Itoa(i))
}

func (m *marshaller) writeFloat(f float64) {
	m.writeTypeMarker(typeFloat)
	m.writeKeyIndex(strconv.FormatFloat(f, 'g', -1, 64))
}

func (m *marshaller) writeKeyIndex(key string) {
	offset, ok := m.keyIndex[key]
	if !ok {
		offset = len(m.keys)
		m.keyIndex[key] = offset
		m.keys = append(m.keys, key)
	}
	if m.shortStringIndices {
		m.writeRawInt16(offset)
	} else {
		m.writeRawInt32(offset)
	}
}

func (m *marshaller) writeKeys() {
	byteOrder.PutUint32(m.buf[8:], uint32(m.offset))

	m.writeTypeMarker(typeOffsets)
	keyOffset := m.offset
	lengths := make([]int, len(m.keys))
	for i, key := range m.keys {
		lengths[i] = len(key) + 1
		keyOffset += 4
	}
	keyOffset += 5
	for _, length := range lengths {
		m.writeRawInt32(keyOffset)
		keyOffset += length
	}
	m.writeRawInt32(0xFFFFFFFF)

	m.writeByte(0x8)
	for _, key := range m.keys {
		m.writeKey(key)
	}
}

func (m *marshaller) writeKey(key string) {
	m.buf = append(m.buf, []byte(key)...)
	m.offset += len(key)
	m.writeByte(0)
}

func (m *marshaller) writeRawInt32(i int) {
	intBytes := make([]byte, 4)
	byteOrder.PutUint32(intBytes, uint32(i))
	m.buf = append(m.buf, intBytes...)
	m.offset += len(intBytes)
}

func (m *marshaller) writeRawInt16(i int) {
	intBytes := make([]byte, 2)
	byteOrder.PutUint16(intBytes, uint16(i))
	m.buf = append(m.buf, intBytes...)
	m.offset += len(intBytes)
}

func (m *marshaller) writeByte(b byte) {
	m.buf = append(m.buf, b)
	m.offset++
}
