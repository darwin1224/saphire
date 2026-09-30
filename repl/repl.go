package repl

import (
	"bufio"
	"fmt"
	"io"

	"github.com/darwin1224/saphire/ast"
	"github.com/darwin1224/saphire/interpreter"
	"github.com/darwin1224/saphire/lexer"
	"github.com/darwin1224/saphire/object"
	"github.com/darwin1224/saphire/parser"
)

const (
	Prompt = ">>"
	// MaxLineSize includes the line's delimiter and bounds memory use per scan.
	MaxLineSize = 1 << 20
)

func Start(in io.Reader, out io.Writer) error {
	scanner := bufio.NewScanner(in)
	scanner.Buffer(make([]byte, 4096), MaxLineSize)
	env := object.NewEnvironment()
	evaluator := interpreter.New(out)

	for {
		if _, err := fmt.Fprint(out, Prompt); err != nil {
			return err
		}
		scanned := scanner.Scan()
		if !scanned {
			if err := scanner.Err(); err != nil {
				return fmt.Errorf("read input: %w", err)
			}
			return nil
		}

		line := scanner.Text()
		lexer := lexer.New(line)
		parser := parser.New(lexer)

		program := parser.ParseProgram()
		if len(parser.Errors()) > 0 {
			if err := printParserErrors(out, parser.Errors()); err != nil {
				return err
			}
			continue
		}
		if len(program.Statements) == 0 {
			continue
		}

		result := evaluator.Eval(program, env)
		_, declaration := program.Statements[len(program.Statements)-1].(*ast.LetStatement)
		if !declaration || result.Type() == object.ERROR_OBJ {
			if _, err := fmt.Fprintln(out, result.Inspect()); err != nil {
				return err
			}
		}
	}
}

func printParserErrors(out io.Writer, errors []string) error {
	if _, err := fmt.Fprintln(out, "parser errors:"); err != nil {
		return err
	}
	for _, msg := range errors {
		if _, err := fmt.Fprintln(out, "\t"+msg); err != nil {
			return err
		}
	}
	return nil
}
