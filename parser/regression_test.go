package parser

import (
	"strings"
	"testing"

	"github.com/darwin1224/saphire/ast"
	"github.com/darwin1224/saphire/lexer"
)

func TestRejectMalformedSource(t *testing.T) {
	tests := []struct{ input, message string }{
		{`fn() { 1`, "expected closing }"},
		{`if (true) { 1`, "expected closing }"},
		{`if (false) {} else { 1`, "expected closing }"},
		{`fn() { if (true) { 1 }`, "expected closing }"},
		{`"abc`, "unterminated string"},
		{`fn(1) { 42 }`, "expected next token to be IDENT"},
		{`fn(x, true) { x }`, "expected next token to be IDENT"},
		{`fn(x,) { x }`, "expected next token to be IDENT"},
		{`fn(x, x) { x }`, "duplicate parameter: x"},
		{`[1, 2`, "expected next token to be ]"},
		{`{"x": 1`, "expected next token to be ,"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			p := New(lexer.New(tt.input))
			p.ParseProgram()
			if !strings.Contains(strings.Join(p.Errors(), "\n"), tt.message) {
				t.Fatalf("expected %q, got %v", tt.message, p.Errors())
			}
		})
	}
}

func TestPowerPrecedence(t *testing.T) {
	for _, tt := range []struct{ input, expected string }{
		{"2 * 3 ** 2", "(2 * (3 ** 2))"},
		{"2 ** 3 ** 2", "(2 ** (3 ** 2))"},
		{"-2 ** 2", "(-(2 ** 2))"},
		{"2 ** -2", "(2 ** (-2))"},
		{"(-2) ** 2", "((-2) ** 2)"},
		{"2 ** 3 * 4", "((2 ** 3) * 4)"},
	} {
		p := New(lexer.New(tt.input))
		program := p.ParseProgram()
		checkParserErrors(t, p)
		if got := program.String(); got != tt.expected {
			t.Errorf("%s: got %s, want %s", tt.input, got, tt.expected)
		}
	}
}

func TestHashLiteralPreservesSourceOrder(t *testing.T) {
	p := New(lexer.New(`{"b": 1, "a": 2, "b": 3}`))
	program := p.ParseProgram()
	checkParserErrors(t, p)
	hash := program.Statements[0].(*ast.ExpressionStatement).Expression.(*ast.HashLiteral)
	if len(hash.Pairs) != 3 || hash.String() != "{b:1, a:2, b:3}" {
		t.Fatalf("hash entries lost source order: %s", hash.String())
	}
}

func FuzzParserDoesNotPanic(f *testing.F) {
	for _, input := range []string{"// comment", `"abc`, `fn(1) {`, `if (true) {`, `{"x": 1}`, `fn(x) { x ** 2 }(3)`, "", "[1,]"} {
		f.Add(input)
	}
	f.Fuzz(func(t *testing.T, input string) {
		if len(input) > 4096 {
			t.Skip()
		}
		p := New(lexer.New(input))
		program := p.ParseProgram()
		if len(p.Errors()) == 0 {
			_ = program.String()
		}
	})
}
