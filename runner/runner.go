// Package runner executes JavaScript from files, stdin, or strings.
package runner

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/dop251/goja"
	"github.com/dop251/goja/parser"
)

// RunFile reads path and executes its contents in vm.
func RunFile(vm *goja.Runtime, path string) {
	src, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "gode: %v\n", err)
		os.Exit(1)
	}
	abs, _ := filepath.Abs(path)
	vm.Set("__filename", abs)
	vm.Set("__dirname", filepath.Dir(abs))
	if _, err := vm.RunString(string(src)); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

// RunStdin reads all of stdin and executes it.
func RunStdin(vm *goja.Runtime) {
	src, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "gode: %v\n", err)
		os.Exit(1)
	}
	if _, err := vm.RunString(string(src)); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

// RunEval evaluates the given script string.
func RunEval(vm *goja.Runtime, script string) {
	if _, err := vm.RunString(script); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

// RunPrint evaluates script like RunEval but also prints the result.
func RunPrint(vm *goja.Runtime, script string) {
	val, err := vm.RunString(script)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	fmt.Println(val)
}

// CheckSyntax parses path without executing, exiting 1 on error.
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
