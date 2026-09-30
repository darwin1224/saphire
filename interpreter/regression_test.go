package interpreter

import (
	"bytes"
	"errors"
	"testing"

	"github.com/darwin1224/saphire/lexer"
	"github.com/darwin1224/saphire/object"
	"github.com/darwin1224/saphire/parser"
)

func TestExplicitReturnsCompose(t *testing.T) {
	for _, tt := range []struct {
		input string
		want  float64
	}{
		{`let f = fn(x) { return x; }; f(5) + 1;`, 6},
		{`let f = fn() { return 5; }; f(); 99;`, 99},
		{`let f = fn(x) { return x; }; let n = f(5); n + 1;`, 6},
		{`let f = fn(x) { return x; }; f(f(5));`, 5},
		{`let f = fn(x) { if (x > 0) { return x; } return 0; }; f(5) + 1;`, 6},
		{`let f = fn(x) { if (x == 0) { return 1; } return x * f(x - 1); }; f(5);`, 120},
	} {
		testNumberObject(t, testEval(t, tt.input), tt.want)
	}
}

func TestExpressionsAlwaysProduceLanguageValues(t *testing.T) {
	for _, input := range []string{"", "// comment", `fn() {}()`, `if (true) {}`, `if (false) {} else {}`, `fn() { let x = 1; }()`, `let x = 1;`, `let x = fn() {}(); x;`} {
		t.Run(input, func(t *testing.T) {
			testNullObject(t, testEval(t, input))
		})
	}
	testBooleanObject(t, testEval(t, `!fn() {}()`), true)
	result := testEval(t, `[fn() {}(), fn() { let x = 1; }()]`)
	if result.Inspect() != "[nil, nil]" {
		t.Fatalf("unexpected empty function results: %s", result.Inspect())
	}
}

func TestRuntimeErrorsDoNotPanic(t *testing.T) {
	for _, tt := range []struct{ input, message string }{
		{`fn(x, y) { x + y; }(1)`, "wrong number of arguments, got=1, want=2"},
		{`fn(x) { x; }(1, 2)`, "wrong number of arguments, got=2, want=1"},
		{`fn() {}(1)`, "wrong number of arguments, got=1, want=0"},
		{`1 % 0`, "modulo by zero"},
		{`1 % -0`, "modulo by zero"},
		{`len(fn() {}())`, "argument to `len` not supported, got NIL"},
		{`if (true) {} + 1`, "type mismatch: NIL + NUMBER"},
		{`fn() {}()()`, "not a function: NIL"},
		{`[42][0 / 0]`, "array index must be a finite integer"},
		{`[42][1 / 0]`, "array index must be a finite integer"},
		{`[42][-1 / 0]`, "array index must be a finite integer"},
		{`[10, 20][0.5]`, "array index must be a finite integer"},
		{`{0 / 0: 42}`, "unusable as hash key: NaN"},
		{`{}[0 / 0]`, "unusable as hash key: NaN"},
	} {
		t.Run(tt.input, func(t *testing.T) {
			result := testEval(t, tt.input)
			err, ok := result.(*object.Error)
			if !ok || err.Message != tt.message {
				t.Fatalf("expected %q, got %T (%s)", tt.message, result, result.Inspect())
			}
		})
	}
}

func TestDecimalArithmeticAndPower(t *testing.T) {
	for _, tt := range []struct {
		input string
		want  float64
	}{
		{"5.5 % 2", 1.5}, {"1 % 0.5", 0}, {"-5.5 % 2", -1.5}, {"5.5 % -2", 1.5},
		{"2 * 3 ** 2", 18}, {"2 ** 3 ** 2", 512}, {"-2 ** 2", -4}, {"2 ** -2", 0.25},
	} {
		testNumberObject(t, testEval(t, tt.input), tt.want)
	}
}

func TestNumericHashIdentity(t *testing.T) {
	for _, tt := range []struct {
		input string
		want  float64
	}{
		{`len({1.1: "a", 1.9: "b", -1.1: "c", -1.9: "d"})`, 4},
		{`{1.1: 11, 1.9: 19}[1.1]`, 11},
		{`{1.1: 11, 1.9: 19}[1.9]`, 19},
		{`{-0: 5}[0]`, 5}, {`{0: 5}[-0]`, 5},
	} {
		testNumberObject(t, testEval(t, tt.input), tt.want)
	}
	testNullObject(t, testEval(t, `{1.1: "a"}[1.9]`))
}

func evalWithOutput(t *testing.T, i *Interpreter, input string) object.Object {
	t.Helper()
	p := parser.New(lexer.New(input))
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}
	return i.Eval(program, object.NewEnvironment())
}

func TestHashEvaluationOrder(t *testing.T) {
	// Repeated evaluations exercise the map-order regression as well as duplicate keys.
	for n := 0; n < 50; n++ {
		var out bytes.Buffer
		evalWithOutput(t, New(&out), `let key = fn(x) { print(x); x; }; {key("a"): print(1), key("b"): print(2)};`)
		if got := out.String(); got != "a\n1\nb\n2\n" {
			t.Fatalf("hash entries evaluated out of source order: %q", got)
		}
		testNumberObject(t, testEval(t, `{"x": 1, "x": 2}["x"]`), 2)
	}
}

func TestPrintUsesSessionOutput(t *testing.T) {
	for n := 0; n < 2; n++ {
		t.Run(string(rune('a'+n)), func(t *testing.T) {
			t.Parallel()
			var out bytes.Buffer
			result := evalWithOutput(t, New(&out), `let p = print; let f = fn() { p(0.001); }; f(); print(fn() { let x = 1; }());`)
			testNullObject(t, result)
			if got := out.String(); got != "0.001\nnil\n" {
				t.Fatalf("print output escaped session writer: %q", got)
			}
		})
	}
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, errors.New("output unavailable") }

func TestPrintReportsWriterErrors(t *testing.T) {
	result := evalWithOutput(t, New(failingWriter{}), `print(1); 2;`)
	err, ok := result.(*object.Error)
	if !ok || err.Message != "print: output unavailable" {
		t.Fatalf("expected print error, got %s", result.Inspect())
	}
}

func TestArrayBuiltins(t *testing.T) {
	for _, tt := range []struct{ input, want string }{
		{`len([1, 2])`, "2"}, {`len({"x": 1})`, "1"},
		{`first([1, 2])`, "1"}, {`first([])`, "nil"},
		{`last([1, 2])`, "2"}, {`last([])`, "nil"},
		{`rest([1, 2])`, "[2]"}, {`rest([1])`, "[]"}, {`rest([])`, "nil"},
		{`push([], 1)`, "[1]"}, {`let a = [1]; let b = push(a, 2); [a, b]`, "[[1], [1, 2]]"},
		{`let a = [1, 2]; let b = rest(a); [a, b]`, "[[1, 2], [2]]"},
		{`first(1)`, "ERROR: argument to `first` must be ARRAY, got NUMBER"},
		{`last(1)`, "ERROR: argument to `last` must be ARRAY, got NUMBER"},
		{`rest(1)`, "ERROR: argument to `rest` must be ARRAY, got NUMBER"},
		{`push(1, 2)`, "ERROR: argument to `push` must be ARRAY, got NUMBER"},
		{`first()`, "ERROR: wrong number of arguments. got=0, want=1"},
		{`last(1, 2)`, "ERROR: wrong number of arguments. got=2, want=1"},
		{`rest()`, "ERROR: wrong number of arguments. got=0, want=1"},
		{`push([])`, "ERROR: wrong number of arguments. got=1, want=2"},
	} {
		t.Run(tt.input, func(t *testing.T) {
			if got := testEval(t, tt.input).Inspect(); got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}
