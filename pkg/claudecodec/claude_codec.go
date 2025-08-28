package claudecodec

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"unicode/utf8"
)

// Data represents the data structure that can contain strings, int32s, or nested Data arrays
type Data []interface{}

// Type constants for our binary format
const (
	TypeInt32  byte = 0x01
	TypeString byte = 0x02
	TypeList   byte = 0x03
)

// Encode serializes Data into a compact binary format
// Time Complexity: O(n) where n is the total number of elements across all nested structures
// Space Complexity: O(n) for the output buffer, O(d) for recursion stack where d is max depth
func Encode(data Data) ([]byte, error) {
	buf := &bytes.Buffer{}
	err := encodeValue(buf, data)
	return buf.Bytes(), err
}

// encodeValue recursively encodes a single value
func encodeValue(buf *bytes.Buffer, value interface{}) error {
	switch v := value.(type) {
	case int32:
		// Format: [TypeInt32:1][value:4]
		buf.WriteByte(TypeInt32)
		return binary.Write(buf, binary.LittleEndian, v)

	case string:
		// Validate UTF-8
		if !utf8.ValidString(v) {
			return fmt.Errorf("invalid UTF-8 string")
		}
		// Check length constraint
		if len(v) > 1000000 {
			return fmt.Errorf("string exceeds maximum length of 1,000,000 bytes")
		}
		// Format: [TypeString:1][length:4][utf8_bytes:length]
		buf.WriteByte(TypeString)
		length := uint32(len(v))
		if err := binary.Write(buf, binary.LittleEndian, length); err != nil {
			return err
		}
		buf.WriteString(v)

	case Data:
		// Check length constraint
		if len(v) > 1000 {
			return fmt.Errorf("list exceeds maximum length of 1000")
		}
		// Format: [TypeList:1][count:4][element1][element2]...[elementN]
		buf.WriteByte(TypeList)
		count := uint32(len(v))
		if err := binary.Write(buf, binary.LittleEndian, count); err != nil {
			return err
		}
		for _, item := range v {
			if err := encodeValue(buf, item); err != nil {
				return err
			}
		}

	case []interface{}:
		// Handle slice converted to Data
		return encodeValue(buf, Data(v))

	default:
		return fmt.Errorf("unsupported type: %T", v)
	}

	return nil
}

// Decode deserializes binary data back into Data structure
// Time Complexity: O(n) where n is the total number of elements
// Space Complexity: O(n) for the result + O(d) for recursion stack where d is max depth
func Decode(data []byte) (Data, error) {
	buf := bytes.NewReader(data)
	value, err := decodeValue(buf)
	if err != nil {
		return nil, err
	}

	// The root must be a list
	if result, ok := value.(Data); ok {
		return result, nil
	}

	return nil, fmt.Errorf("root value must be a list")
}

// decodeValue recursively decodes a single value
func decodeValue(buf *bytes.Reader) (interface{}, error) {
	typeByte, err := buf.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to read type byte: %v", err)
	}

	switch typeByte {
	case TypeInt32:
		var value int32
		err := binary.Read(buf, binary.LittleEndian, &value)
		if err != nil {
			return nil, fmt.Errorf("failed to read int32: %v", err)
		}
		return value, nil

	case TypeString:
		var length uint32
		err := binary.Read(buf, binary.LittleEndian, &length)
		if err != nil {
			return nil, fmt.Errorf("failed to read string length: %v", err)
		}

		if length > 1000000 {
			return nil, fmt.Errorf("string length %d exceeds maximum of 1,000,000", length)
		}

		stringBytes := make([]byte, length)
		n, err := buf.Read(stringBytes)
		if err != nil || uint32(n) != length {
			return nil, fmt.Errorf("failed to read string data: %v", err)
		}

		str := string(stringBytes)
		if !utf8.ValidString(str) {
			return nil, fmt.Errorf("invalid UTF-8 string")
		}

		return str, nil

	case TypeList:
		var count uint32
		err := binary.Read(buf, binary.LittleEndian, &count)
		if err != nil {
			return nil, fmt.Errorf("failed to read list count: %v", err)
		}

		if count > 1000 {
			return nil, fmt.Errorf("list count %d exceeds maximum of 1000", count)
		}

		result := make(Data, count)
		for i := uint32(0); i < count; i++ {
			value, err := decodeValue(buf)
			if err != nil {
				return nil, fmt.Errorf("failed to decode list element %d: %v", i, err)
			}
			result[i] = value
		}

		return result, nil

	default:
		return nil, fmt.Errorf("unknown type byte: 0x%02x", typeByte)
	}
}

// Helper function to create Data from various inputs
func NewData(items ...interface{}) Data {
	return Data(items)
}

// String representation for debugging
func (d Data) String() string {
	return fmt.Sprintf("%v", []interface{}(d))
}
