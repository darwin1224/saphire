package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"

	"github.com/darwin1224/saphire/compiler"
	"github.com/darwin1224/saphire/interpreter"
	"github.com/darwin1224/saphire/lexer"
	"github.com/darwin1224/saphire/object"
	"github.com/darwin1224/saphire/parser"
	"github.com/darwin1224/saphire/repl"
	"github.com/darwin1224/saphire/vm"
)

const (
	SaphireExt = ".sp"

	VMMode       = "vm"
	TreeWalkMode = "treewalk"
)

var (
	interpreterMode = flag.String("mode", "treewalk", "Interpreter mode. Default: treewalk.")
)

func main() {
	flag.Parse()

	if len(os.Args) <= 1 {
		startRepl(*interpreterMode)
		return
	}

	filename := os.Args[1]
	buf, err := readFile(filename)
	if err != nil {
		panic(err)
	}

	switch *interpreterMode {
	case VMMode:
		runInterpreterInVM(buf)
	case TreeWalkMode:
		runInterpreterInTreeWalk(buf)
	}
}

func startRepl(mode string) {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Hello %s! This is the Saphire programming language!\n", user.Username)
	fmt.Printf("Feel free to type in commands\n")

	repl.Start(os.Stdin, os.Stdout, mode)
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

func runInterpreterInVM(buf []byte) error {
	constants := make([]object.Object, 0)
	globals := make([]object.Object, vm.GlobalSize)

	symbolTable := compiler.NewSymbolTable()
	for i, v := range object.Builtins {
		symbolTable.DefineBuiltin(i, v.Name)
	}

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

	return nil
}

func runInterpreterInTreeWalk(buf []byte) {
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

func printParserErrors(out io.Writer, errors []string) {
	for _, msg := range errors {
		io.WriteString(out, " parser errors:\n")
		io.WriteString(out, "\t"+msg+"\n")
	}
}
