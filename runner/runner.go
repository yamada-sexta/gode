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
	if err := ExecFile(vm, path); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

// ExecFile is the testable core of RunFile. It reads a file and executes
// it on the vm, returning any error.
func ExecFile(vm *goja.Runtime, path string) error {
	src, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("gode: %v", err)
	}
	abs, _ := filepath.Abs(path)
	vm.Set("__filename", abs)
	vm.Set("__dirname", filepath.Dir(abs))
	_, err = vm.RunString(string(src))
	return err
}

// RunStdin reads all of stdin and executes it.
func RunStdin(vm *goja.Runtime) {
	if err := ExecStdin(vm, os.Stdin); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

// ExecStdin is the testable core of RunStdin.
func ExecStdin(vm *goja.Runtime, r io.Reader) error {
	src, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("gode: %v", err)
	}
	_, err = vm.RunString(string(src))
	return err
}

// RunEval evaluates the given script string.
func RunEval(vm *goja.Runtime, script string) {
	if err := ExecEval(vm, script); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

// ExecEval is the testable core of RunEval.
func ExecEval(vm *goja.Runtime, script string) error {
	_, err := vm.RunString(script)
	return err
}

// RunPrint evaluates script like RunEval but also prints the result.
func RunPrint(vm *goja.Runtime, script string) {
	val, err := ExecPrint(vm, script)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	fmt.Println(val)
}

// ExecPrint is the testable core of RunPrint.
func ExecPrint(vm *goja.Runtime, script string) (goja.Value, error) {
	return vm.RunString(script)
}

// CheckSyntax parses path without executing, exiting 1 on error.
func CheckSyntax(path string) {
	if err := ValidateSyntax(path); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

// ValidateSyntax is the testable core of CheckSyntax.
func ValidateSyntax(path string) error {
	src, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("gode: %v", err)
	}
	_, err = parser.ParseFile(nil, path, string(src), 0)
	return err
}
