package main

import (
	"fmt"

	codec "github.com/tenaciousjzh/customcodec/pkg/claudecodec"
)

// Example usage and tests
func main() {
	// Test case 1: Simple nested structure
	original := codec.NewData("foo", codec.NewData("bar", int32(42)))
	fmt.Printf("Original: %v\n", original)

	encoded, err := codec.Encode(original)
	if err != nil {
		fmt.Printf("Encode error: %v\n", err)
		return
	}

	fmt.Printf("Encoded length: %d bytes\n", len(encoded))

	decoded, err := codec.Decode(encoded)
	if err != nil {
		fmt.Printf("Decode error: %v\n", err)
		return
	}

	fmt.Printf("Decoded: %v\n", decoded)
	fmt.Printf("Round trip successful: %v\n", fmt.Sprintf("%v", original) == fmt.Sprintf("%v", decoded))

	// Test case 2: More complex nested structure
	complex := codec.NewData(
		"hello",
		int32(-123),
		codec.NewData(
			"nested",
			codec.NewData("deeply", int32(999)),
			"world",
		),
		int32(0),
	)

	fmt.Printf("\nComplex original: %v\n", complex)

	encoded2, err := codec.Encode(complex)
	if err != nil {
		fmt.Printf("Encode error: %v\n", err)
		return
	}

	fmt.Printf("Complex encoded length: %d bytes\n", len(encoded2))

	decoded2, err := codec.Decode(encoded2)
	if err != nil {
		fmt.Printf("Decode error: %v\n", err)
		return
	}

	fmt.Printf("Complex decoded: %v\n", decoded2)
	fmt.Printf("Complex round trip successful: %v\n", fmt.Sprintf("%v", complex) == fmt.Sprintf("%v", decoded2))

	// Test case 3: Empty structures
	empty := codec.NewData()
	fmt.Printf("\nEmpty original: %v\n", empty)

	encoded3, err := codec.Encode(empty)
	if err != nil {
		fmt.Printf("Encode error: %v\n", err)
		return
	}

	decoded3, err := codec.Decode(encoded3)
	if err != nil {
		fmt.Printf("Decode error: %v\n", err)
		return
	}

	fmt.Printf("Empty decoded: %v\n", decoded3)
	fmt.Printf("Empty round trip successful: %v\n", len(empty) == len(decoded3))

	// Test case 4: Unicode strings
	unicode := codec.NewData("Hello 世界", "🚀 emoji", int32(42))
	fmt.Printf("\nUnicode original: %v\n", unicode)

	encoded4, err := codec.Encode(unicode)
	if err != nil {
		fmt.Printf("Encode error: %v\n", err)
		return
	}

	decoded4, err := codec.Decode(encoded4)
	if err != nil {
		fmt.Printf("Decode error: %v\n", err)
		return
	}

	fmt.Printf("Unicode decoded: %v\n", decoded4)
	fmt.Printf("Unicode round trip successful: %v\n", fmt.Sprintf("%v", unicode) == fmt.Sprintf("%v", decoded4))
}
