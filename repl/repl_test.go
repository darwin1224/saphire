package repl

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestSessionStateAndOutput(t *testing.T) {
	var out bytes.Buffer
	if err := Start(strings.NewReader("let x = 2;\nprint(x);\nx + 1;\n"), &out); err != nil {
		t.Fatal(err)
	}
	if got, want := out.String(), ">>>>2\nnil\n>>3\n>>"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestCommentsAndBlankLinesAreSilent(t *testing.T) {
	var out bytes.Buffer
	if err := Start(strings.NewReader("\n// comment"), &out); err != nil {
		t.Fatal(err)
	}
	if got := out.String(); got != ">>>>>>" {
		t.Fatalf("unexpected results for blank/comment input: %q", got)
	}
}

func TestErrorsDoNotEndSession(t *testing.T) {
	var out bytes.Buffer
	input := "if (true) { 1\nlet bad = len(fn() {}());\n2;\n"
	if err := Start(strings.NewReader(input), &out); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"parser errors:", "expected closing }", "ERROR: argument to `len` not supported, got NIL", ">>2\n>>"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("output %q missing %q", out.String(), want)
		}
	}
}

func TestLongLines(t *testing.T) {
	var out bytes.Buffer
	input := "1;" + strings.Repeat(" ", 65536) + "\n2;\n"
	if err := Start(strings.NewReader(input), &out); err != nil {
		t.Fatal(err)
	}
	if out.String() != ">>1\n>>2\n>>" {
		t.Fatalf("long line ended session: %q", out.String())
	}
	err := Start(strings.NewReader(strings.Repeat(" ", MaxLineSize)+"\n"), io.Discard)
	if !errors.Is(err, bufio.ErrTooLong) {
		t.Fatalf("expected line limit error, got %v", err)
	}
}

var errStream = errors.New("stream unavailable")

type failingReader struct{}

func (failingReader) Read([]byte) (int, error) { return 0, errStream }

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, errStream }

func TestStreamErrors(t *testing.T) {
	if err := Start(failingReader{}, io.Discard); !errors.Is(err, errStream) {
		t.Fatalf("expected reader error, got %v", err)
	}
	if err := Start(strings.NewReader("1;"), failingWriter{}); !errors.Is(err, errStream) {
		t.Fatalf("expected writer error, got %v", err)
	}
}
