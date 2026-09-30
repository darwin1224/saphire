package main

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func writeScript(t *testing.T, source string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.sp")
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestRunScript(t *testing.T) {
	for _, tt := range []struct {
		name, source, stdout, errorText string
		status                          int
	}{
		{"success", `print(0.001);`, "0.001\n", "", 0},
		{"empty", "", "", "", 0},
		{"comment", "// no trailing newline", "", "", 0},
		{"runtime error", `5 + true;`, "", "ERROR: type mismatch: NUMBER + BOOLEAN", 1},
		{"identifier error", `doesNotExist;`, "", "ERROR: identifier not found: doesNotExist", 1},
		{"syntax error", `if (true) { print(42);`, "", "expected closing }", 1},
		{"unterminated string", `print("abc`, "", "unterminated string", 1},
		{"argument count", `fn(x) { x; }();`, "", "wrong number of arguments", 1},
		{"NaN index", `print([42][0 / 0]);`, "", "array index must be a finite integer", 1},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var out, errOut bytes.Buffer
			status := run([]string{writeScript(t, tt.source)}, strings.NewReader(""), &out, &errOut)
			if status != tt.status || out.String() != tt.stdout {
				t.Fatalf("got status=%d output=%q; want status=%d output=%q", status, out.String(), tt.status, tt.stdout)
			}
			if tt.errorText == "" && errOut.Len() != 0 || !strings.Contains(errOut.String(), tt.errorText) {
				t.Fatalf("unexpected error output: %q (want %q)", errOut.String(), tt.errorText)
			}
		})
	}
}

func TestRunFileErrors(t *testing.T) {
	for _, args := range [][]string{
		{filepath.Join(t.TempDir(), "missing.sp")},
		{filepath.Join(t.TempDir(), "wrong.txt")},
		{"one.sp", "two.sp"},
	} {
		var errOut bytes.Buffer
		if status := run(args, strings.NewReader(""), io.Discard, &errOut); status != 1 || errOut.Len() == 0 {
			t.Fatalf("expected reported file/usage failure for %v, got status=%d error=%q", args, status, errOut.String())
		}
	}
}

func TestRunREPL(t *testing.T) {
	var out, errOut bytes.Buffer
	status := run(nil, strings.NewReader("print(1);\n// end"), &out, &errOut)
	if status != 0 || errOut.Len() != 0 || !strings.Contains(out.String(), ">>1\nnil\n>>>>") {
		t.Fatalf("unexpected REPL result: status=%d output=%q error=%q", status, out.String(), errOut.String())
	}
}

// Run the actual entry point in a subprocess so os.Exit behavior is verified.
func TestCLIProcess(t *testing.T) {
	if os.Getenv("SAPHIRE_TEST_CLI") != "1" {
		return
	}
	for n, arg := range os.Args {
		if arg == "--" {
			os.Args = append([]string{"saphire"}, os.Args[n+1:]...)
			main()
			t.Fatal("main returned without setting an exit status")
		}
	}
	t.Fatal("missing CLI arguments")
}

func TestCLIExitStatus(t *testing.T) {
	for _, tt := range []struct {
		source string
		status int
	}{
		{`print(1);`, 0}, {`5 + true;`, 1}, {`let x = ;`, 1},
	} {
		cmd := exec.Command(os.Args[0], "-test.run=^TestCLIProcess$", "--", writeScript(t, tt.source))
		cmd.Env = append(os.Environ(), "SAPHIRE_TEST_CLI=1")
		var out, errOut bytes.Buffer
		cmd.Stdout, cmd.Stderr = &out, &errOut
		err := cmd.Run()
		status := 0
		if exit, ok := err.(*exec.ExitError); ok {
			status = exit.ExitCode()
		} else if err != nil {
			t.Fatal(err)
		}
		if status != tt.status || tt.status != 0 && errOut.Len() == 0 {
			t.Fatalf("%q: status=%d, want=%d, stdout=%q, stderr=%q", tt.source, status, tt.status, out.String(), errOut.String())
		}
	}
}
