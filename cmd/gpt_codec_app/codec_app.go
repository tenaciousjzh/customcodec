package main

import (
	"fmt"
	"reflect"

	codec "github.com/tenaciousjzh/customcodec/pkg/gptcodec"
)

// ----------------------------------------------------------------------------
// Demo / basic self-check
// ----------------------------------------------------------------------------

func main() {
	// Example: ["foo", ["bar", 42]]
	var data codec.Data = []codec.Data{
		"foo",
		[]codec.Data{"bar", int32(42)},
	}

	enc, err := codec.Encode(data)
	if err != nil {
		panic(err)
	}

	dec, err := codec.Decode(enc)
	if err != nil {
		panic(err)
	}

	if !reflect.DeepEqual(dec, data) {
		panic(fmt.Sprintf("round-trip mismatch:\nwant: %#v\n got: %#v", data, dec))
	}

	fmt.Println("Round-trip OK. Bytes:", len(enc))

	// Extra quick checks
	cases := []codec.Data{
		int32(-7),
		"",
		make([]codec.Data, 0),
		[]codec.Data{"αβγ", int32(123456), []codec.Data{"nested", []codec.Data{"deep"}}},
	}
	for i, cse := range cases {
		b, err := codec.Encode(cse)
		if err != nil {
			panic(err)
		}
		v, err := codec.Decode(b)
		if err != nil {
			panic(err)
		}
		if !reflect.DeepEqual(v, cse) {
			panic(fmt.Sprintf("case %d mismatch", i))
		}
	}
	fmt.Println("All demo cases passed.")
}
