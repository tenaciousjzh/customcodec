# High-Performance Custom Codec

A zero-copy, memory-efficient binary serialization codec for Go, designed for high-performance network communication.

## Building and Installation

### Prerequisites
- Go 1.18 or higher
- A working Go environment with `$GOPATH` set

### Building from Source

1. Clone the repository:
```bash
git clone https://github.com/tenaciousjzh/customcodec.git
cd customcodec
```

2. Build the project:

Option 1 - Build in place:
```bash
# Creates binary in the local bin directory
mkdir -p bin
go build -o bin/zero_copy_codec_app ./cmd/zero_copy_codec_app/
```

Option 2 - Install to $GOPATH:
```bash
# Installs binary to $GOPATH/bin
go install ./cmd/zero_copy_codec_app/
```

### Verifying the Installation

After building, you can verify the installation:

For local build:
```bash
./bin/zero_copy_codec_app
```

For Go installation:
```bash
zero_copy_codec_app
```

Note: If using `go install`, ensure `$GOPATH/bin` is in your system's PATH.

## Original Requirements

Design a high-performance data encoding algorithm for network communication with the following interface:

```typescript
type Data = Array<string | int32 | Data>
function encode(to_send: Data) => string
function decode(received: string) => Data
```

### Constraints
- Input is a list of integer, string and nested list types
- Maximum list length: 1,000
- Maximum string length: 1,000,000
- Strings can contain any UTF-8 symbols
- Integer type is a valid int32 value

Example:
```go
decode(encode(["foo", ["bar", 42]])) == ["foo", ["bar", 42]]
```

## Implementation

This implementation provides a high-performance solution with the following features:

- Zero-copy string deserialization
- Buffer pooling for reduced allocations
- Consistent cross-platform binary format
- Support for recursive data structures
- Minimal allocations during encoding/decoding

### Data Format

The binary format consists of:
```
[4 bytes: total length][remaining bytes: data]

Each element is encoded as:
[1 byte: type][4 bytes: length/value][data bytes if applicable]
```

Supported types:
- Int32 (type 0x01)
- String (type 0x02)
- Array (type 0x03)

### Usage Example

```go
package main

import (
    "fmt"
    "github.com/tenaciousjzh/customcodec/pkg/zerocopycodec"
)

func main() {
    // Create a sample data structure
    data := zerocopycodec.Data{
        int32(42),
        "hello",
        zerocopycodec.Data{
            int32(1),
            "nested",
        },
    }

    // Encode
    encoded := zerocopycodec.Encode(data)

    // Decode
    decoded, err := zerocopycodec.Decode(encoded)
    if err != nil {
        panic(err)
    }

    // Compare
    if zerocopycodec.DeepEqual(data, decoded) {
        fmt.Println("Successfully encoded and decoded!")
    }
}
```

### Performance Optimization

For better performance in high-throughput scenarios, use the encoder pool:

```go
encoder := zerocopycodec.GetPooledEncoder()
encoded := encoder.Encode(data)
zerocopycodec.ReturnPooledEncoder(encoder)
```

## Testing

### Unit Tests

Run all unit tests:
```bash
go test ./...
```

Run tests with verbose output:
```bash
go test -v ./...
```

### Benchmark Tests

Run all benchmarks:
```bash
go test -bench=. ./...
```

Run benchmarks with memory allocation statistics:
```bash
go test -bench=. -benchmem ./...
```

Run a specific benchmark:
```bash
go test -bench=BenchmarkEncode ./...
```

## Performance Characteristics

## Complexity Analysis

### Time Complexity

#### Encoding (O(n))
- **Base case**: O(1) for primitive types (int32)
- **String handling**: O(k) where k is the string length
- **Array handling**: O(n) where n is the total number of elements in the array and its nested structures
- **Overall**: O(n) where n represents the total size of all elements (including string lengths)

Detailed breakdown of encoding operations:
```
Operation                Time Complexity
----------------------------------------
Type marker writing     O(1)
Integer encoding        O(1)
String encoding         O(k) where k is string length
Array length writing    O(1)
Recursive array calls   O(n) total across all elements
```

#### Decoding (O(n))
- **Base case**: O(1) for primitive types (int32)
- **String handling**: O(1) due to zero-copy optimization
- **Array handling**: O(n) where n is the number of elements
- **Overall**: O(n) where n is the total number of elements

Detailed breakdown of decoding operations:
```
Operation                Time Complexity
----------------------------------------
Type marker reading     O(1)
Integer decoding        O(1)
String decoding         O(1) [zero-copy]
Array length reading    O(1)
Recursive array calls   O(n) total across all elements
```

### Space Complexity

#### Encoding (O(m))
- **Buffer allocation**: O(m) where m is the total serialized size
- **Stack space**: O(d) where d is the maximum nesting depth
- **Temporary allocations**: O(1) due to buffer reuse

Space usage breakdown:
```
Component               Space Usage
----------------------------------------
Output buffer          O(m) total size
Recursive call stack   O(d) nesting depth
Temporary variables    O(1)
```

#### Decoding (O(k))
- **Zero-copy strings**: O(1) per string (points to input buffer)
- **Array structures**: O(k) where k is the number of arrays
- **Stack space**: O(d) where d is the maximum nesting depth

Space optimization details:
```
Component               Space Usage
----------------------------------------
String references      O(1) per string
Array structures       O(k) total arrays
Input buffer          O(1) [shared reference]
Recursive call stack   O(d) nesting depth
```

### Performance Guarantees

- No heap allocations for string data during decoding (zero-copy)
- Constant memory overhead per array structure
- Buffer pooling reduces allocation overhead for repeated operations
- Predictable memory usage based on input structure
- Linear time scaling with data size

### Optimization Notes

1. **Buffer Pooling**
   - Reduces allocation overhead
   - Amortizes memory allocation cost
   - Pool size limited to prevent memory bloat

2. **Zero-Copy String Handling**
   - Eliminates string data copying
   - Reduces memory pressure
   - Improves decode performance

3. **Fixed-Size Headers**
   - Constant-time length calculations
   - Predictable memory layout
   - Efficient random access

### Memory Optimizations
1. Zero-Copy Strings:
   - Uses unsafe pointers to create strings without copying bytes
   - Significantly reduces memory allocations during deserialization

2. Buffer Pooling:
   - Maintains a pool of reusable buffers
   - Buffers larger than 64KB are not pooled to prevent memory bloat

3. Consistent Endianness:
   - All integers stored in little-endian format
   - Ensures consistent cross-platform compatibility

## Thread Safety Notes

- Encoder and Decoder are not thread-safe
- The encoder pool is thread-safe
- For concurrent use, create separate encoder/decoder instances per goroutine

## Benchmark Results

Benchmarks run on:
- OS: Linux
- Architecture: amd64
- CPU: Intel(R) Core(TM) i7-4810MQ CPU @ 2.80GHz

### Small Data Performance
```
BenchmarkSmallData/Encode-8         31,394,000     38.42 ns/op
BenchmarkSmallData/Decode-8          2,577,895    732.7 ns/op
BenchmarkSmallData/RoundTrip-8       1,570,861    753.9 ns/op
```
Small data operations are extremely fast, with encoding taking only ~38ns and a full round-trip under 1μs.

### Medium Data Performance
```
BenchmarkMediumData/Encode-8           42,034     26,879 ns/op
BenchmarkMediumData/Decode-8          189,003     13,560 ns/op
BenchmarkMediumData/RoundTrip-8        17,556     74,273 ns/op
```
Medium-sized data shows good throughput with decode operations being notably faster than encode operations.

### Large Data Performance
```
BenchmarkLargeData/Encode-8               78     13.85 ms/op
BenchmarkLargeData/Decode-8            12,372     91.42 μs/op
BenchmarkLargeData/RoundTrip-8             88     13.25 ms/op
```
For large data sets, the zero-copy strategy shows its strength with decode operations being significantly faster than encode operations.

### Encoder Pooling Impact
```
BenchmarkPooledEncoder/WithoutPooling-8    1,000,000      1,738 ns/op
BenchmarkPooledEncoder/WithPooling-8      11,097,712        106.8 ns/op
```
Encoder pooling provides a dramatic performance improvement, being about 16x faster than without pooling.

### Memory Allocation
```
BenchmarkMemoryComparison/ZeroCopyDecode-8    2,710,084    477.4 ns/op    104 B/op    4 allocs/op
```
The zero-copy implementation is extremely memory efficient, requiring only 104 bytes and 4 allocations per operation.
