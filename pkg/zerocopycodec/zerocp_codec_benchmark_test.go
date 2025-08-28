package zerocopycodec

import (
	"strings"
	"testing"
)

// =============================================================================
// BENCHMARK TESTS
// =============================================================================

// BenchmarkSmallData tests performance with small datasets
func BenchmarkSmallData(b *testing.B) {
	smallData := Data{
		"hello",
		int32(42),
		Data{"nested", int32(-1)},
	}

	b.Run("Encode", func(b *testing.B) {
		encoder := NewEncoder(256)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			encoded := encoder.Encode(smallData)
			_ = encoded
		}
	})

	encoded := Encode(smallData)
	b.Run("Decode", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			decoded, err := Decode(encoded)
			if err != nil {
				b.Fatal(err)
			}
			_ = decoded
		}
	})

	b.Run("RoundTrip", func(b *testing.B) {
		encoder := NewEncoder(256)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			encoded := encoder.Encode(smallData)
			decoded, err := Decode(encoded)
			if err != nil {
				b.Fatal(err)
			}
			_ = decoded
		}
	})
}

// BenchmarkMediumData tests performance with medium datasets (100 elements, 10KB strings)
func BenchmarkMediumData(b *testing.B) {
	mediumString := strings.Repeat("medium test data ", 588) // ~10KB
	mediumData := make(Data, 100)

	for i := 0; i < 100; i++ {
		if i%3 == 0 {
			mediumData[i] = mediumString
		} else if i%3 == 1 {
			mediumData[i] = int32(i)
		} else {
			mediumData[i] = Data{mediumString, int32(i)}
		}
	}

	b.Run("Encode", func(b *testing.B) {
		encoder := NewEncoder(1024 * 1024) // 1MB initial capacity
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			encoded := encoder.Encode(mediumData)
			_ = encoded
		}
	})

	encoded := Encode(mediumData)
	b.Run("Decode", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			decoded, err := Decode(encoded)
			if err != nil {
				b.Fatal(err)
			}
			_ = decoded
		}
	})

	b.Run("RoundTrip", func(b *testing.B) {
		encoder := NewEncoder(1024 * 1024)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			encoded := encoder.Encode(mediumData)
			decoded, err := Decode(encoded)
			if err != nil {
				b.Fatal(err)
			}
			_ = decoded
		}
	})
}

// BenchmarkLargeData tests performance with large datasets (1000 elements, 1MB strings)
func BenchmarkLargeData(b *testing.B) {
	// Create 1MB string (max allowed)
	largeString := strings.Repeat("This is a large test string for benchmarking purposes. ", 18519) // ~1MB

	// Create array with 1000 elements (max allowed)
	largeData := make(Data, 1000)

	for i := 0; i < 1000; i++ {
		if i%10 == 0 {
			largeData[i] = largeString // 1MB string
		} else if i%10 < 5 {
			largeData[i] = int32(i)
		} else {
			// Nested structure
			largeData[i] = Data{
				"nested",
				int32(i),
				strings.Repeat("nest", 250), // 1KB string
			}
		}
	}

	b.Run("Encode", func(b *testing.B) {
		encoder := NewEncoder(10 * 1024 * 1024) // 10MB initial capacity
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			encoded := encoder.Encode(largeData)
			_ = encoded
		}
	})

	encoded := Encode(largeData)
	b.Logf("Large data encoded to %d bytes (%.2f MB)", len(encoded), float64(len(encoded))/(1024*1024))

	b.Run("Decode", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			decoded, err := Decode(encoded)
			if err != nil {
				b.Fatal(err)
			}
			_ = decoded
		}
	})

	b.Run("RoundTrip", func(b *testing.B) {
		encoder := NewEncoder(10 * 1024 * 1024)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			encoded := encoder.Encode(largeData)
			decoded, err := Decode(encoded)
			if err != nil {
				b.Fatal(err)
			}
			_ = decoded
		}
	})
}

// BenchmarkPooledEncoder tests the performance benefit of encoder pooling
func BenchmarkPooledEncoder(b *testing.B) {
	testData := Data{
		"pooled test",
		int32(123),
		Data{"nested", strings.Repeat("data", 250)},
	}

	b.Run("WithoutPooling", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			encoder := NewEncoder(1024)
			encoded := encoder.Encode(testData)
			_ = encoded
		}
	})

	b.Run("WithPooling", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			encoder := GetPooledEncoder()
			encoded := encoder.Encode(testData)
			ReturnPooledEncoder(encoder)
			_ = encoded
		}
	})
}

// BenchmarkMemoryComparison compares our format against a naive approach
func BenchmarkMemoryComparison(b *testing.B) {
	testString := strings.Repeat("comparison test ", 6250) // 100KB string
	testData := Data{testString, int32(42), testString}

	b.Run("ZeroCopyDecode", func(b *testing.B) {
		encoded := Encode(testData)
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			decoded, err := Decode(encoded)
			if err != nil {
				b.Fatal(err)
			}
			_ = decoded
		}
	})
}
