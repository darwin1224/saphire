package lexer

import (
	"strings"
	"testing"

	"github.com/darwin1224/saphire/token"
)

func TestCommentsAtEOF(t *testing.T) {
	for _, input := range []string{"// comment", "//", "// first\n // last", strings.Repeat("// comment\n", 10000) + "// last"} {
		t.Run(input[:min(len(input), 30)], func(t *testing.T) {
			if tok := New(input).NextToken(); tok.Type != token.EOF {
				t.Fatalf("expected EOF, got %+v", tok)
			}
		})
	}
}

func TestUnterminatedString(t *testing.T) {
	for _, input := range []string{`"`, `"abc`, "\"abc\n"} {
		l := New(input)
		if tok := l.NextToken(); tok.Type != token.ILLEGAL || tok.Literal != "unterminated string" {
			t.Fatalf("%q: expected unterminated string error, got %+v", input, tok)
		}
		if tok := l.NextToken(); tok.Type != token.EOF {
			t.Fatalf("%q: expected EOF after error, got %+v", input, tok)
		}
	}
}

func FuzzLexerTerminates(f *testing.F) {
	for _, input := range []string{"// comment", `"abc`, `let x = 1.5; // tail`, "// first\n// second", "", "\x00"} {
		f.Add(input)
	}
	f.Fuzz(func(t *testing.T, input string) {
		if len(input) > 4096 {
			t.Skip()
		}
		l := New(input)
		for n := 0; n <= len(input)+1; n++ {
			if l.NextToken().Type == token.EOF {
				return
			}
		}
		t.Fatal("lexer failed to reach EOF within the input length")
	})
}
