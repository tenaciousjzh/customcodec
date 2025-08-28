package gptcodec

import (
	"reflect"
	"testing"
)

func roundTrip(t *testing.T, v Data) {
	b, err := Encode(v)
	if err != nil {
		t.Fatalf("encode error: %v", err)
	}
	out, err := Decode(b)
	if err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if !reflect.DeepEqual(out, v) {
		t.Fatalf("mismatch: want %#v got %#v", v, out)
	}
}

func TestRoundTrips(t *testing.T) {
	cases := []Data{
		"foo",
		int32(-42),
		[]Data{"bar", int32(123)},
		[]Data{"αβγ", []Data{"nested", []Data{"deep"}}},
	}
	for _, cse := range cases {
		roundTrip(t, cse)
	}
}

func TestConstraints(t *testing.T) {
	// oversize string
	bigStr := make([]byte, MaxStringLen+1)
	if _, err := Encode(string(bigStr)); err == nil {
		t.Fatal("expected oversize string error")
	}
	// oversize list
	bigList := make([]Data, MaxListLen+1)
	if _, err := Encode(bigList); err == nil {
		t.Fatal("expected oversize list error")
	}
}
