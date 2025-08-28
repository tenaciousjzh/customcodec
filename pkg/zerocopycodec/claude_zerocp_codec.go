package zerocopycodec

import (
	"encoding/binary"
	"fmt"
	"unsafe"
)

// Data represents our recursive data structure
type Data []interface{}

// ValueType represents the type of a value in our format
type ValueType byte

const (
	TypeInt32  ValueType = 0x01
	TypeString ValueType = 0x02
	TypeArray  ValueType = 0x03
)

// Header layout for our format:
// [4 bytes: total length][remaining bytes: data]
//
// Each element layout:
// [1 byte: type][4 bytes: length/value][data bytes if applicable]

// Encoder provides efficient serialization with minimal allocations
type Encoder struct {
	buf []byte
}

// NewEncoder creates a new encoder with initial capacity
func NewEncoder(initialCap int) *Encoder {
	return &Encoder{
		buf: make([]byte, 0, initialCap),
	}
}

// Encode serializes Data to a byte slice
// Time Complexity: O(n) where n is total number of elements + string lengths
// Space Complexity: O(m) where m is the total serialized size
func (e *Encoder) Encode(data Data) []byte {
	e.buf = e.buf[:0] // Reset buffer, reuse capacity

	// Reserve 4 bytes for total length header
	e.buf = append(e.buf, 0, 0, 0, 0)

	e.encodeValue(data)

	// Write total length at the beginning
	totalLen := len(e.buf) - 4
	binary.LittleEndian.PutUint32(e.buf[0:4], uint32(totalLen))

	return e.buf
}

func (e *Encoder) encodeValue(value interface{}) {
	switch v := value.(type) {
	case int32:
		e.buf = append(e.buf, byte(TypeInt32))
		e.buf = binary.LittleEndian.AppendUint32(e.buf, uint32(v))

	case string:
		e.buf = append(e.buf, byte(TypeString))
		strBytes := []byte(v)
		e.buf = binary.LittleEndian.AppendUint32(e.buf, uint32(len(strBytes)))
		e.buf = append(e.buf, strBytes...)

	case Data:
		e.buf = append(e.buf, byte(TypeArray))
		e.buf = binary.LittleEndian.AppendUint32(e.buf, uint32(len(v)))
		for _, item := range v {
			e.encodeValue(item)
		}

	case []interface{}:
		// Handle native Go slice
		e.buf = append(e.buf, byte(TypeArray))
		e.buf = binary.LittleEndian.AppendUint32(e.buf, uint32(len(v)))
		for _, item := range v {
			e.encodeValue(item)
		}
	}
}

// Decoder provides zero-copy deserialization by working directly with the source buffer
type Decoder struct {
	data   []byte
	offset int
}

// NewDecoder creates a decoder that works directly with the provided byte slice
// No copying occurs - all string and array references point into the original buffer
func NewDecoder(data []byte) *Decoder {
	return &Decoder{
		data:   data,
		offset: 0,
	}
}

// Decode deserializes the data with zero-copy for strings and minimal allocations
// Time Complexity: O(n) where n is the number of elements
// Space Complexity: O(k) where k is the number of arrays/slices created (strings are zero-copy)
func (d *Decoder) Decode() (Data, error) {
	if len(d.data) < 4 {
		return nil, fmt.Errorf("invalid data: too short")
	}

	// Read total length
	totalLen := binary.LittleEndian.Uint32(d.data[0:4])
	if len(d.data) != int(totalLen)+4 {
		return nil, fmt.Errorf("invalid data: length mismatch")
	}

	d.offset = 4 // Skip the length header

	value, err := d.decodeValue()
	if err != nil {
		return nil, err
	}

	if arr, ok := value.(Data); ok {
		return arr, nil
	}

	return nil, fmt.Errorf("root element must be an array")
}

func (d *Decoder) decodeValue() (interface{}, error) {
	if d.offset >= len(d.data) {
		return nil, fmt.Errorf("unexpected end of data")
	}

	valueType := ValueType(d.data[d.offset])
	d.offset++

	switch valueType {
	case TypeInt32:
		if d.offset+4 > len(d.data) {
			return nil, fmt.Errorf("insufficient data for int32")
		}
		value := int32(binary.LittleEndian.Uint32(d.data[d.offset:]))
		d.offset += 4
		return value, nil

	case TypeString:
		if d.offset+4 > len(d.data) {
			return nil, fmt.Errorf("insufficient data for string length")
		}
		strLen := binary.LittleEndian.Uint32(d.data[d.offset:])
		d.offset += 4

		if d.offset+int(strLen) > len(d.data) {
			return nil, fmt.Errorf("insufficient data for string content")
		}

		// ZERO-COPY: Create string directly from buffer slice
		// This uses unsafe to avoid copying the bytes
		strBytes := d.data[d.offset : d.offset+int(strLen)]
		str := *(*string)(unsafe.Pointer(&strBytes))
		d.offset += int(strLen)

		return str, nil

	case TypeArray:
		if d.offset+4 > len(d.data) {
			return nil, fmt.Errorf("insufficient data for array length")
		}
		arrayLen := binary.LittleEndian.Uint32(d.data[d.offset:])
		d.offset += 4

		// Pre-allocate slice with known capacity
		result := make(Data, 0, arrayLen)

		for i := uint32(0); i < arrayLen; i++ {
			value, err := d.decodeValue()
			if err != nil {
				return nil, err
			}
			result = append(result, value)
		}

		return result, nil

	default:
		return nil, fmt.Errorf("unknown value type: %d", valueType)
	}
}

// Convenience functions that match the required interface
func Encode(data Data) []byte {
	encoder := NewEncoder(1024) // Start with 1KB capacity
	return encoder.Encode(data)
}

func Decode(encoded []byte) (Data, error) {
	decoder := NewDecoder(encoded)
	return decoder.Decode()
}

// Performance-optimized pool for reusing encoders
var encoderPool = make(chan *Encoder, 10)

func GetPooledEncoder() *Encoder {
	select {
	case encoder := <-encoderPool:
		return encoder
	default:
		return NewEncoder(1024)
	}
}

func ReturnPooledEncoder(encoder *Encoder) {
	if cap(encoder.buf) < 64*1024 { // Don't pool overly large buffers
		select {
		case encoderPool <- encoder:
		default:
			// Pool is full, let GC handle it
		}
	}
}

// Helper function for deep equality checking
func DeepEqual(a, b Data) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if !valueEqual(a[i], b[i]) {
			return false
		}
	}

	return true
}

func valueEqual(a, b interface{}) bool {
	switch va := a.(type) {
	case int32:
		if vb, ok := b.(int32); ok {
			return va == vb
		}
		return false

	case string:
		if vb, ok := b.(string); ok {
			return va == vb
		}
		return false

	case Data:
		if vb, ok := b.(Data); ok {
			return DeepEqual(va, vb)
		}
		return false

	default:
		return false
	}
}
