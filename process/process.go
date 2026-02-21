// Package process registers the global `process` object on an otto VM,
// providing process.env, process.argv, process.exit, process.cwd, and
// process.version — similar to the Node.js process global.
package process

import (
	"os"
	"strings"

	"github.com/robertkrimen/otto"
)

// Setup creates the `process` global on vm.
//
//   - version  — the string returned by process.version
//   - script   — path to the JS file being run (empty for REPL/eval)
//   - args     — extra arguments passed after the script
func Setup(vm *otto.Otto, version, script string, args []string) {
	proc, _ := vm.Object(`process = {}`)

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
	proc.Set("exit", func(call otto.FunctionCall) otto.Value {
		code, _ := call.Argument(0).ToInteger()
		os.Exit(int(code))
		return otto.UndefinedValue()
	})

	// process.cwd()
	proc.Set("cwd", func(call otto.FunctionCall) otto.Value {
		cwd, _ := os.Getwd()
		result, _ := vm.ToValue(cwd)
		return result
	})

	// process.env — snapshot of the current environment
	envObj, _ := vm.Object(`process.env = {}`)
	for _, e := range os.Environ() {
		if k, v, ok := strings.Cut(e, "="); ok {
			envObj.Set(k, v)
		}
	}
}
