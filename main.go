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
	if len(os.Args) <= 1 {
		startRepl()
		return
	}

	filename := os.Args[1]
	buf, err := readFile(filename)
	if err != nil {
		panic(err)
	}

	env := object.NewEnvironment()
	lexer := lexer.New(string(buf))
	parser := parser.New(lexer)

	program := parser.ParseProgram()
	if len(parser.Errors()) > 0 {
		printParserErrors(os.Stdout, parser.Errors())
		return
	}

	interpreter.Eval(program, env)
}

func startRepl() {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Hello %s! This is the Saphire programming language!\n", user.Username)
	fmt.Printf("Feel free to type in commands\n")

	repl.Start(os.Stdin, os.Stdout)
}

func readFile(filename string) ([]byte, error) {
	ext := filepath.Ext(filename)
	if ext != SaphireExt {
		return nil, fmt.Errorf("error: invalid file extension %s (expected .sp)", ext)
	}

	buf, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

func printParserErrors(out io.Writer, errors []string) {
	for _, msg := range errors {
		io.WriteString(out, " parser errors:\n")
		io.WriteString(out, "\t"+msg+"\n")
	}
}
