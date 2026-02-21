// Package runner provides functions that execute JavaScript from
// various sources: files, stdin, strings, and syntax-only checking.
package runner

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/robertkrimen/otto"
	"github.com/robertkrimen/otto/parser"

	"gode/compat"
)

// RunFile reads path and executes its contents in vm.
func RunFile(vm *otto.Otto, path string) {
	src, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "gode: %v\n", err)
		os.Exit(1)
	}
	abs, _ := filepath.Abs(path)
	vm.Set("__filename", abs)
	vm.Set("__dirname", filepath.Dir(abs))
	if _, err := vm.Run(compat.Transform(string(src))); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

// RunStdin reads all of standard input and executes it in vm.
func RunStdin(vm *otto.Otto) {
	src, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "gode: error reading stdin: %v\n", err)
		os.Exit(1)
	}
	if _, err := vm.Run(compat.Transform(string(src))); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

// RunEval executes script as JavaScript in vm.
func RunEval(vm *otto.Otto, script string) {
	if _, err := vm.Run(compat.Transform(script)); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

// RunPrint executes script and prints the resulting value.
func RunPrint(vm *otto.Otto, script string) {
	value, err := vm.Run(compat.Transform(script))
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	fmt.Println(value)
}

// CheckSyntax parses the file at path without executing it.
// Exits with code 1 if the file contains syntax errors.
func CheckSyntax(path string) {
	src, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "gode: %v\n", err)
		os.Exit(1)
	}
	if _, err := parser.ParseFile(nil, path, string(src), 0); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
