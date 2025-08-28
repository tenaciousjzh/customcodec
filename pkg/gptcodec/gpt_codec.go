package gptcodec

import (
	"errors"
	"fmt"
)

// ----------------------------------------------------------------------------
// Custom Type-Length-Value wire format (no JSON/Protobuf/etc.)
// ----------------------------------------------------------------------------
// Each value is encoded as: [Tag:1][Length:4][Payload]
// Tags:
//   'S' (0x53): String    -> Length = byte length of UTF-8 payload, Payload = bytes
//   'I' (0x49): Int32     -> Length = 4, Payload = 4 bytes big-endian two's complement
//   'L' (0x4C): List<Data>-> Length = element count (u32). Then exactly N concatenated TLV values
//
// Constraints enforced:
//  - Max list length: 1000
//  - Max string length: 1_000_000 bytes
//
// Complexity:
//  - Encode:  O(n) time in the total size of the output; O(d) auxiliary space (recursion depth)
//  - Decode:  O(n) time in the total size of the input;  O(d) auxiliary space (recursion depth)
// where n is total bytes on the wire, d is nesting depth. No extra copies beyond building output.
// ----------------------------------------------------------------------------

const (
	TagString byte = 'S'
	TagInt32  byte = 'I'
	TagList   byte = 'L'

	MaxListLen   = 1000
	MaxStringLen = 1_000_000
)

// Data is one of: string | int32 | []Data
// (Using any for simplicity; validate at runtime.)
type Data = any

// ----------------------------------------------------------------------------
// Public API
// ----------------------------------------------------------------------------

// Encode encodes Data into the custom binary wire format.
func Encode(v Data) ([]byte, error) {
	buf := make([]byte, 0, 64)
	return encodeValue(buf, v)
}

// EncodeString mirrors the spec signature (string as raw bytes container in Go).
func EncodeString(v Data) (string, error) {
	b, err := Encode(v)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Decode parses a byte slice produced by Encode and returns Data.
func Decode(b []byte) (Data, error) {
	v, off, err := decodeValue(b, 0)
	if err != nil {
		return nil, err
	}
	if off != len(b) {
		return nil, fmt.Errorf("trailing bytes: decoded %d of %d", off, len(b))
	}
	return v, nil
}

// DecodeString mirrors the spec signature.
func DecodeString(s string) (Data, error) { return Decode([]byte(s)) }

// ----------------------------------------------------------------------------
// Encoding helpers
// ----------------------------------------------------------------------------

func encodeValue(dst []byte, v Data) ([]byte, error) {
	switch x := v.(type) {
	case string:
		bs := []byte(x)
		if len(bs) > MaxStringLen {
			return nil, fmt.Errorf("string too long: %d > %d", len(bs), MaxStringLen)
		}
		dst = append(dst, TagString)
		dst = writeU32(dst, uint32(len(bs)))
		dst = append(dst, bs...)
		return dst, nil
	case int32:
		dst = append(dst, TagInt32)
		dst = writeU32(dst, 4)
		dst = writeI32(dst, x)
		return dst, nil
	case []Data:
		if len(x) > MaxListLen {
			return nil, fmt.Errorf("list too long: %d > %d", len(x), MaxListLen)
		}
		dst = append(dst, TagList)
		dst = writeU32(dst, uint32(len(x))) // element count
		for _, elem := range x {
			var err error
			dst, err = encodeValue(dst, elem)
			if err != nil {
				return nil, err
			}
		}
		return dst, nil
	default:
		return nil, fmt.Errorf("unsupported type %T (allowed: string | int32 | []Data)", v)
	}
}

func writeU32(dst []byte, v uint32) []byte {
	return append(dst,
		byte(v>>24),
		byte(v>>16),
		byte(v>>8),
		byte(v),
	)
}

func writeI32(dst []byte, v int32) []byte { return writeU32(dst, uint32(v)) }

// ----------------------------------------------------------------------------
// Decoding helpers
// ----------------------------------------------------------------------------

type cursor struct {
	b   []byte
	off int
}

func (c *cursor) need(n int) error {
	if c.off+n > len(c.b) {
		return errors.New("unexpected EOF")
	}
	return nil
}

func (c *cursor) readByte() (byte, error) {
	if err := c.need(1); err != nil {
		return 0, err
	}
	v := c.b[c.off]
	c.off++
	return v, nil
}

func (c *cursor) readU32() (uint32, error) {
	if err := c.need(4); err != nil {
		return 0, err
	}
	b0, b1, b2, b3 := c.b[c.off], c.b[c.off+1], c.b[c.off+2], c.b[c.off+3]
	c.off += 4
	return (uint32(b0) << 24) | (uint32(b1) << 16) | (uint32(b2) << 8) | uint32(b3), nil
}

func (c *cursor) readN(n int) ([]byte, error) {
	if err := c.need(n); err != nil {
		return nil, err
	}
	v := c.b[c.off : c.off+n]
	c.off += n
	return v, nil
}

func decodeValue(b []byte, start int) (Data, int, error) {
	c := &cursor{b: b, off: start}
	tag, err := c.readByte()
	if err != nil {
		return nil, start, err
	}
	switch tag {
	case TagString:
		ln, err := c.readU32()
		if err != nil {
			return nil, start, err
		}
		if ln > MaxStringLen {
			return nil, start, fmt.Errorf("string too long: %d > %d", ln, MaxStringLen)
		}
		payload, err := c.readN(int(ln))
		if err != nil {
			return nil, start, err
		}
		return string(payload), c.off, nil
	case TagInt32:
		ln, err := c.readU32()
		if err != nil {
			return nil, start, err
		}
		if ln != 4 {
			return nil, start, fmt.Errorf("int32 length must be 4, got %d", ln)
		}
		u, err := c.readU32()
		if err != nil {
			return nil, start, err
		}
		return int32(u), c.off, nil
	case TagList:
		count, err := c.readU32()
		if err != nil {
			return nil, start, err
		}
		if count > MaxListLen {
			return nil, start, fmt.Errorf("list too long: %d > %d", count, MaxListLen)
		}
		res := make([]Data, 0, count)
		for i := uint32(0); i < count; i++ {
			v, off, err := decodeValue(b, c.off)
			if err != nil {
				return nil, start, err
			}
			c.off = off
			res = append(res, v)
		}
		return res, c.off, nil
	default:
		return nil, start, fmt.Errorf("unknown tag 0x%X at offset %d", tag, start)
	}
}
