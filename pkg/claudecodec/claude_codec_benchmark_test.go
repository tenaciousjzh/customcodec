package claudecodec

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"testing"
)

// Benchmark test data generators
func generateString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789 "
	b := make([]byte, length)
	for i := range b {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		b[i] = charset[n.Int64()]
	}
	return string(b)
}

func generateInt32() int32 {
	n, _ := rand.Int(rand.Reader, big.NewInt(4294967296))
	return int32(n.Int64() - 2147483648)
}

func generateSimpleData(size int) Data {
	data := make(Data, size)
	for i := 0; i < size; i++ {
		if i%2 == 0 {
			data[i] = generateString(10 + i%90) // Strings 10-100 chars
		} else {
			data[i] = generateInt32()
		}
	}
	return data
}

func generateNestedData(depth, breadth int) Data {
	if depth <= 0 {
		return NewData(generateString(20), generateInt32())
	}

	data := make(Data, breadth)
	for i := 0; i < breadth; i++ {
		if i%3 == 0 {
			data[i] = generateNestedData(depth-1, breadth)
		} else if i%3 == 1 {
			data[i] = generateString(10 + i%40)
		} else {
			data[i] = generateInt32()
		}
	}
	return data
}

func generateLargeStringData(stringSize int) Data {
	return NewData(
		generateString(stringSize),
		int32(42),
		generateString(stringSize/2),
	)
}

// Benchmark tests
func BenchmarkEncodeSimpleSmall(b *testing.B) {
	data := generateSimpleData(10)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := Encode(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEncodeSimpleMedium(b *testing.B) {
	data := generateSimpleData(100)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := Encode(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEncodeSimpleLarge(b *testing.B) {
	data := generateSimpleData(1000)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := Encode(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEncodeNestedShallow(b *testing.B) {
	data := generateNestedData(2, 5) // 2 levels deep, 5 items per level
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := Encode(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEncodeNestedDeep(b *testing.B) {
	data := generateNestedData(5, 3) // 5 levels deep, 3 items per level
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := Encode(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEncodeLargeStrings(b *testing.B) {
	data := generateLargeStringData(10000) // 10KB strings
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := Encode(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEncodeUnicodeStrings(b *testing.B) {
	data := NewData(
		"Hello 世界 🌍 Привет мир",
		"🚀🌟💫⭐🌙🌞",
		int32(42),
		"これは日本語のテストです",
	)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := Encode(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDecodeSimpleSmall(b *testing.B) {
	data := generateSimpleData(10)
	encoded, _ := Encode(data)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := Decode(encoded)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDecodeSimpleMedium(b *testing.B) {
	data := generateSimpleData(100)
	encoded, _ := Encode(data)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := Decode(encoded)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDecodeSimpleLarge(b *testing.B) {
	data := generateSimpleData(1000)
	encoded, _ := Encode(data)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := Decode(encoded)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDecodeNestedShallow(b *testing.B) {
	data := generateNestedData(2, 5)
	encoded, _ := Encode(data)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := Decode(encoded)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDecodeNestedDeep(b *testing.B) {
	data := generateNestedData(5, 3)
	encoded, _ := Encode(data)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := Decode(encoded)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDecodeLargeStrings(b *testing.B) {
	data := generateLargeStringData(10000)
	encoded, _ := Encode(data)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := Decode(encoded)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRoundTripSmall(b *testing.B) {
	data := generateSimpleData(10)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		encoded, err := Encode(data)
		if err != nil {
			b.Fatal(err)
		}
		_, err = Decode(encoded)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRoundTripMedium(b *testing.B) {
	data := generateSimpleData(100)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		encoded, err := Encode(data)
		if err != nil {
			b.Fatal(err)
		}
		_, err = Decode(encoded)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRoundTripLarge(b *testing.B) {
	data := generateSimpleData(1000)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		encoded, err := Encode(data)
		if err != nil {
			b.Fatal(err)
		}
		_, err = Decode(encoded)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Memory allocation benchmarks
func BenchmarkEncodeMemoryAlloc(b *testing.B) {
	data := generateSimpleData(100)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := Encode(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDecodeMemoryAlloc(b *testing.B) {
	data := generateSimpleData(100)
	encoded, _ := Encode(data)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := Decode(encoded)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Compression ratio test (not a benchmark but useful)
func TestCompressionRatio(t *testing.T) {
	testCases := []struct {
		name string
		data Data
	}{
		{"Simple", NewData("hello", int32(42), "world")},
		{"Nested", NewData("foo", NewData("bar", int32(123)), "baz")},
		{"Large String", generateLargeStringData(1000)},
		{"Many Small Items", generateSimpleData(50)},
		{"Deep Nesting", generateNestedData(4, 3)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			encoded, err := Encode(tc.data)
			if err != nil {
				t.Fatal(err)
			}

			// Rough JSON equivalent size estimate
			jsonEstimate := len(fmt.Sprintf("%v", tc.data))
			binarySize := len(encoded)

			ratio := float64(binarySize) / float64(jsonEstimate)

			t.Logf("%s: Binary=%d bytes, JSON estimate=%d bytes, Ratio=%.2f",
				tc.name, binarySize, jsonEstimate, ratio)
		})
	}
}
