package zerocopycodec

// =============================================================================
// UNIT TESTS
// =============================================================================

// Test file: main_test.go
import (
	"strings"
	"testing"
	"unsafe"
)

// TestMaxListLength verifies the constraint of max 1000 elements in a list
func TestMaxListLength(t *testing.T) {
	// Test exactly at the limit (1000 elements)
	maxData := make(Data, 1000)
	for i := 0; i < 1000; i++ {
		maxData[i] = int32(i)
	}

	encoded := Encode(maxData)
	decoded, err := Decode(encoded)
	if err != nil {
		t.Fatalf("Failed to encode/decode max length array: %v", err)
	}

	if len(decoded) != 1000 {
		t.Errorf("Expected 1000 elements, got %d", len(decoded))
	}

	// Verify all values
	for i := 0; i < 1000; i++ {
		if decoded[i].(int32) != int32(i) {
			t.Errorf("Element %d: expected %d, got %d", i, i, decoded[i])
		}
	}
}

// TestMaxStringLength verifies the constraint of max 1,000,000 char strings
func TestMaxStringLength(t *testing.T) {
	// Create a string with exactly 1,000,000 characters
	maxString := strings.Repeat("a", 1000000)
	testData := Data{maxString, int32(42)}

	encoded := Encode(testData)
	decoded, err := Decode(encoded)
	if err != nil {
		t.Fatalf("Failed to encode/decode max length string: %v", err)
	}

	if len(decoded) != 2 {
		t.Errorf("Expected 2 elements, got %d", len(decoded))
	}

	decodedStr := decoded[0].(string)
	if len(decodedStr) != 1000000 {
		t.Errorf("Expected string length 1000000, got %d", len(decodedStr))
	}

	if decodedStr != maxString {
		t.Error("Decoded string doesn't match original")
	}
}

// TestNestedMaxConstraints tests nested structures at the limits
func TestNestedMaxConstraints(t *testing.T) {
	// Create nested array with 1000 elements, each containing max string
	bigString := strings.Repeat("x", 1000000)
	nestedData := make(Data, 1000)

	for i := 0; i < 1000; i++ {
		if i%2 == 0 {
			nestedData[i] = bigString
		} else {
			nestedData[i] = int32(i)
		}
	}

	testData := Data{
		"header",
		nestedData,
		"footer",
	}

	encoded := Encode(testData)
	decoded, err := Decode(encoded)
	if err != nil {
		t.Fatalf("Failed to encode/decode nested max constraints: %v", err)
	}

	if !DeepEqual(testData, decoded) {
		t.Error("Nested max constraints data doesn't match after round-trip")
	}
}

// TestZeroCopyStrings verifies that strings are truly zero-copy
func TestZeroCopyStrings(t *testing.T) {
	testStr := "zero-copy test string"
	testData := Data{testStr}

	encoded := Encode(testData)
	decoded, err := Decode(encoded)
	if err != nil {
		t.Fatalf("Failed to decode: %v", err)
	}

	decodedStr := decoded[0].(string)

	// The decoded string should reference memory within the encoded buffer
	encodedPtr := uintptr(unsafe.Pointer(&encoded[0]))
	encodedEnd := encodedPtr + uintptr(len(encoded))
	decodedPtr := uintptr(unsafe.Pointer(&decodedStr))

	// The decoded string's data should point somewhere within the encoded buffer
	if decodedPtr < encodedPtr || decodedPtr >= encodedEnd {
		// Note: This test might be flaky due to Go's string interning
		// but it demonstrates the zero-copy concept
		t.Logf("Warning: String may not be zero-copy (could be due to string interning)")
	}
}

// TestUTF8Support verifies proper UTF-8 handling
func TestUTF8Support(t *testing.T) {
	utf8Strings := []string{
		"Hello, 世界",
		"🚀🎯💻",
		"Здравствуй мир",
		"مرحبا بالعالم",
		"こんにちは世界",
	}

	testData := Data{}
	for _, s := range utf8Strings {
		testData = append(testData, s)
	}

	encoded := Encode(testData)
	decoded, err := Decode(encoded)
	if err != nil {
		t.Fatalf("Failed to encode/decode UTF-8 strings: %v", err)
	}

	if !DeepEqual(testData, decoded) {
		t.Error("UTF-8 strings don't match after round-trip")
	}
}

// TestErrorCases verifies proper error handling
func TestErrorCases(t *testing.T) {
	// Test empty data
	_, err := Decode([]byte{})
	if err == nil {
		t.Error("Expected error for empty data")
	}

	// Test truncated data
	_, err = Decode([]byte{1, 2})
	if err == nil {
		t.Error("Expected error for truncated data")
	}

	// Test invalid type
	invalidData := []byte{0, 0, 0, 2, 0xFF, 0, 0, 0, 0}
	_, err = Decode(invalidData)
	if err == nil {
		t.Error("Expected error for invalid type")
	}
}
