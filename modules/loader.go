// Package modules provides a require()-based module loading system
// for the gode JavaScript runtime. Built-in modules are embedded at
// compile time and evaluated lazily on first require().
package modules

import (
	_ "embed"
	"fmt"

	"github.com/robertkrimen/otto"
)

//go:embed assert.js
var assertSource string

// Loader manages built-in module registration, caching, and the
// require() function installed on the VM.
type Loader struct {
	vm       *otto.Otto
	registry map[string]string     // name → JS source
	cache    map[string]otto.Value // name → cached exports
}

// NewLoader creates a Loader, registers all built-in modules, and
// installs require() on vm.
func NewLoader(vm *otto.Otto) *Loader {
	l := &Loader{
		vm:       vm,
		registry: make(map[string]string),
		cache:    make(map[string]otto.Value),
	}

	// Built-in modules.
	l.Register("assert", assertSource)
	l.Register("node:assert", assertSource)

	// Install require().
	vm.Set("require", l.require)

	return l
}

// Register adds a built-in module. source must be a JS expression
// (typically an IIFE) that evaluates to the module's exports value.
func (l *Loader) Register(name, source string) {
	l.registry[name] = source
}

func (l *Loader) require(call otto.FunctionCall) otto.Value {
	name := call.Argument(0).String()

	// Return from cache.
	if val, ok := l.cache[name]; ok {
		return val
	}

	src, ok := l.registry[name]
	if !ok {
		panic(l.vm.MakeCustomError("Error", fmt.Sprintf("Cannot find module '%s'", name)))
	}

	val, err := l.vm.Run(src)
	if err != nil {
		panic(l.vm.MakeCustomError("Error", fmt.Sprintf("Error loading module '%s': %v", name, err)))
	}

	l.cache[name] = val
	return val
}
