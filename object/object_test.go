package object

import (
	"math"
	"strconv"
	"testing"
)

func TestNumberHashKeys(t *testing.T) {
	values := []float64{0, 1, -1, 1.1, 1.9, -1.1, -1.9, math.Inf(1), math.Inf(-1)}
	for _, a := range values {
		for _, b := range values {
			if equalKeys := (&Number{Value: a}).HashKey() == (&Number{Value: b}).HashKey(); equalKeys != (a == b) {
				t.Errorf("hash identity disagrees with numeric equality: %g, %g", a, b)
			}
		}
	}
	if (&Number{Value: 0}).HashKey() != (&Number{Value: math.Copysign(0, -1)}).HashKey() {
		t.Error("signed zero must address the same hash entry")
	}
}

func TestNumberInspectPreservesPrecision(t *testing.T) {
	for _, value := range []float64{0, 1, -1, 0.001, math.Pi, 1e20, 1e-20, math.SmallestNonzeroFloat64} {
		text := (&Number{Value: value}).Inspect()
		got, err := strconv.ParseFloat(text, 64)
		if err != nil || got != value {
			t.Errorf("%g did not round-trip through Inspect (%q): got %g, err %v", value, text, got, err)
		}
	}
}

func TestStringHashKey(t *testing.T) {
	hello1 := &String{Value: "Hello World"}
	hello2 := &String{Value: "Hello World"}
	diff1 := &String{Value: "My name is johnny"}
	diff2 := &String{Value: "My name is johnny"}

	if hello1.HashKey() != hello2.HashKey() {
		t.Errorf("strings with same content have different hash keys")
	}

	if diff1.HashKey() != diff2.HashKey() {
		t.Errorf("strings with same content have different hash keys")
	}

	if hello1.HashKey() == diff1.HashKey() {
		t.Errorf("strings with different content have same hash keys")
	}
}
