package repl

import (
	"bufio"
	"fmt"
	"io"

	"github.com/darwin1224/saphire/interpreter"
	"github.com/darwin1224/saphire/lexer"
	"github.com/darwin1224/saphire/object"
	"github.com/darwin1224/saphire/parser"
)

const (
	Prompt = ">>"
)

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
	env := object.NewEnvironment()

	for {
		fmt.Printf(Prompt)
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()
		lexer := lexer.New(line)
		parser := parser.New(lexer)

		program := parser.ParseProgram()
		if len(parser.Errors()) > 0 {
			printParserErrors(out, parser.Errors())
			continue
		}

		result := interpreter.Eval(program, env)
		if result != nil {
			io.WriteString(out, result.Inspect())
			io.WriteString(out, "\n")
		}
	}
}

func printParserErrors(out io.Writer, errors []string) {
	for _, msg := range errors {
		io.WriteString(out, " parser errors:\n")
		io.WriteString(out, "\t"+msg+"\n")
	}
}
