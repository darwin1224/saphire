package main

import (
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"

	"github.com/darwin1224/saphire/interpreter"
	"github.com/darwin1224/saphire/lexer"
	"github.com/darwin1224/saphire/object"
	"github.com/darwin1224/saphire/parser"
	"github.com/darwin1224/saphire/repl"
)

const (
	SaphireExt = ".sp"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func run(args []string, in io.Reader, out, errOut io.Writer) int {
	if len(args) == 0 {
		if err := startRepl(in, out); err != nil {
			fmt.Fprintln(errOut, "error:", err)
			return 1
		}
		return 0
	}
	if len(args) != 1 {
		fmt.Fprintln(errOut, "usage: saphire [file.sp]")
		return 1
	}

	filename := args[0]
	buf, err := readFile(filename)
	if err != nil {
		fmt.Fprintln(errOut, "error:", err)
		return 1
	}

	env := object.NewEnvironment()
	lexer := lexer.New(string(buf))
	parser := parser.New(lexer)

	program := parser.ParseProgram()
	if len(parser.Errors()) > 0 {
		printParserErrors(errOut, parser.Errors())
		return 1
	}

	result := interpreter.New(out).Eval(program, env)
	if result.Type() == object.ERROR_OBJ {
		fmt.Fprintln(errOut, result.Inspect())
		return 1
	}
	return 0
}

func startRepl(in io.Reader, out io.Writer) error {
	name := "there"
	user, err := user.Current()
	if err == nil {
		name = user.Username
	}

	if _, err := fmt.Fprintf(out, "Hello %s! This is the Saphire programming language!\nFeel free to type in commands\n", name); err != nil {
		return err
	}

	return repl.Start(in, out)
}

func readFile(filename string) ([]byte, error) {
	ext := filepath.Ext(filename)
	if ext != SaphireExt {
		return nil, fmt.Errorf("invalid file extension %s (expected .sp)", ext)
	}

	buf, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

func printParserErrors(out io.Writer, errors []string) {
	fmt.Fprintln(out, "parser errors:")
	for _, msg := range errors {
		fmt.Fprintln(out, "\t"+msg)
	}
}
