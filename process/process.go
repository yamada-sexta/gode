// Package process registers the global `process` object on a goja VM,
// providing process.env, process.argv, process.exit, process.cwd, and
// process.version — similar to the Node.js process global.
package process

import (
	"fmt"
	"os"
	"strings"

	"github.com/dop251/goja"
)

// Setup creates the `process` global on vm.
//
//   - version  — the string returned by process.version
//   - script   — path to the JS file being run (empty for REPL/eval)
//   - args     — extra arguments passed after the script
func Setup(vm *goja.Runtime, version, script string, args []string) {
	proc := vm.NewObject()
	vm.Set("process", proc)

	// process.version
	proc.Set("version", version)

	// process.argv — ["gode", scriptPath?, ...userArgs]
	argv := []string{"gode"}
	if script != "" {
		argv = append(argv, script)
	}
	argv = append(argv, args...)
	proc.Set("argv", argv)

	// process.exit(code)
	proc.Set("exit", func(call goja.FunctionCall) goja.Value {
		code := call.Argument(0).ToInteger()
		os.Exit(int(code))
		return goja.Undefined()
	})

	// process.cwd()
	proc.Set("cwd", func(call goja.FunctionCall) goja.Value {
		cwd, _ := os.Getwd()
		return vm.ToValue(cwd)
	})

	// process.env — snapshot of the current environment
	envObj := vm.NewObject()
	proc.Set("env", envObj)
	for _, e := range os.Environ() {
		if k, v, ok := strings.Cut(e, "="); ok {
			envObj.Set(k, v)
		}
	}

	// process.nextTick(callback, ...args)
	proc.Set("nextTick", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) == 0 {
			return goja.Undefined()
		}
		callback, ok := goja.AssertFunction(call.Argument(0))
		if !ok {
			panic(vm.ToValue("TypeError [ERR_INVALID_CALLBACK]: Callback must be a function"))
		}
		// Synchronously call for now as gode has no event loop
		if _, err := callback(goja.Undefined(), call.Arguments[1:]...); err != nil {
			panic(err)
		}
		return goja.Undefined()
	})

	// console — goja does not provide a built-in console.
	consolePrint := func(w *os.File) func(goja.FunctionCall) goja.Value {
		return func(call goja.FunctionCall) goja.Value {
			args := make([]interface{}, len(call.Arguments))
			for i, a := range call.Arguments {
				args[i] = a.Export()
			}
			fmt.Fprintln(w, args...)
			return goja.Undefined()
		}
	}
	con := vm.NewObject()
	con.Set("log", consolePrint(os.Stdout))
	con.Set("info", consolePrint(os.Stdout))
	con.Set("debug", consolePrint(os.Stdout))
	con.Set("error", consolePrint(os.Stderr))
	con.Set("warn", consolePrint(os.Stderr))
	vm.Set("console", con)
}
