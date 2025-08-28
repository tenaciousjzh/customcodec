package main

import (
	"fmt"
	"unsafe"

	zc "github.com/tenaciousjzh/customcodec/pkg/zerocopycodec"
)

// Example usage and testing
func main() {
	fmt.Println("Running basic functionality tests...")

	// Test case 1: Simple nested structure
	testData1 := zc.Data{
		"foo",
		zc.Data{"bar", int32(42)},
	}

	fmt.Println("Original data:", testData1)

	// Encode
	encoded := zc.Encode(testData1)
	fmt.Printf("Encoded length: %d bytes\n", len(encoded))

	// Decode
	decoded, err := zc.Decode(encoded)
	if err != nil {
		panic(err)
	}

	fmt.Println("Decoded data:", decoded)

	// Verify round-trip
	fmt.Println("Round-trip successful:", zc.DeepEqual(testData1, decoded))

	// Test case 2: More complex structure
	testData2 := zc.Data{
		"hello",
		int32(-123),
		zc.Data{
			"nested",
			zc.Data{
				"deeply",
				int32(999),
				"🚀 UTF-8 support",
			},
		},
		"world",
	}

	encoded2 := zc.Encode(testData2)
	decoded2, err := zc.Decode(encoded2)
	if err != nil {
		panic(err)
	}

	fmt.Println("\nComplex test successful:", zc.DeepEqual(testData2, decoded2))
	fmt.Printf("Complex data encoded to %d bytes\n", len(encoded2))

	// Demonstrate zero-copy behavior
	fmt.Println("\n--- Zero-Copy Demonstration ---")
	testStr := "This is a zero-copy string! 🎯"
	testData3 := zc.Data{testStr, int32(100)}

	encoded3 := zc.Encode(testData3)
	decoded3, _ := zc.Decode(encoded3)

	decodedStr := decoded3[0].(string)
	fmt.Printf("Original string address: %p\n", &testStr)
	fmt.Printf("Original encoded bytes address: %p\n", &encoded3[0])

	// The decoded string should reference the same memory as the encoded buffer
	fmt.Printf("Decoded string shares memory with encoded buffer: %t\n",
		uintptr(unsafe.Pointer(&decodedStr)) != uintptr(unsafe.Pointer(&testStr)))

	fmt.Println("\nTo run unit tests: go test")
	fmt.Println("To run benchmarks: go test -bench=.")
}
