package repl

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/darwin1224/saphire/compiler"
	"github.com/darwin1224/saphire/interpreter"
	"github.com/darwin1224/saphire/lexer"
	"github.com/darwin1224/saphire/object"
	"github.com/darwin1224/saphire/parser"
	"github.com/darwin1224/saphire/vm"
)

const (
	Prompt = ">>"

	VMMode       = "vm"
	TreeWalkMode = "treewalk"
)

func Start(in io.Reader, out io.Writer, mode string) {
	scanner := bufio.NewScanner(in)

	constants := make([]object.Object, 0)
	globals := make([]object.Object, vm.GlobalSize)

	symbolTable := compiler.NewSymbolTable()
	for i, v := range object.Builtins {
		symbolTable.DefineBuiltin(i, v.Name)
	}

	env := object.NewEnvironment()

	for {
		fmt.Printf(Prompt)
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()

		switch mode {
		case VMMode:
			runInterpreterInVM([]byte(line), symbolTable, constants, globals, out)
		case TreeWalkMode:
			runInterpreterInTreeWalk([]byte(line), env, out)
		}
	}
}

func runInterpreterInVM(buf []byte, symbolTable *compiler.SymbolTable, constants, globals []object.Object, out io.Writer) error {
	lexer := lexer.New(string(buf))
	parser := parser.New(lexer)

	program := parser.ParseProgram()
	if len(parser.Errors()) > 0 {
		printParserErrors(os.Stdout, parser.Errors())
		return nil
	}

	comp := compiler.NewWithState(symbolTable, constants)
	err := comp.Compile(program)
	if err != nil {
		return err
	}

	code := comp.Bytecode()

	machine := vm.NewWithGlobalStore(code, globals)
	err = machine.Run()
	if err != nil {
		return err
	}

	lastPopped := machine.LastPoppedStackElem()
	if lastPopped != nil {
		io.WriteString(out, lastPopped.Inspect())
		io.WriteString(out, "\n")
	}

	return nil
}

func runInterpreterInTreeWalk(buf []byte, env *object.Environment, out io.Writer) {
	lexer := lexer.New(string(buf))
	parser := parser.New(lexer)

	program := parser.ParseProgram()
	if len(parser.Errors()) > 0 {
		printParserErrors(os.Stdout, parser.Errors())
		return
	}

	result := interpreter.Eval(program, env)
	if result != nil {
		io.WriteString(out, result.Inspect())
		io.WriteString(out, "\n")
	}
}

func printParserErrors(out io.Writer, errors []string) {
	for _, msg := range errors {
		io.WriteString(out, " parser errors:\n")
		io.WriteString(out, "\t"+msg+"\n")
	}
}
