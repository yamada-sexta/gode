package process

import (
	"os"
	"strings"
	"testing"

	"github.com/robertkrimen/otto"
)

func newVM(t *testing.T) *otto.Otto {
	t.Helper()
	vm := otto.New()
	Setup(vm, "v1.2.3", "test.js", []string{"arg1", "arg2"})
	return vm
}

func mustRun(t *testing.T, vm *otto.Otto, js string) otto.Value {
	t.Helper()
	val, err := vm.Run(js)
	if err != nil {
		t.Fatalf("unexpected error: %v\nscript: %s", err, js)
	}
	return val
}

func TestProcessVersion(t *testing.T) {
	vm := newVM(t)
	val := mustRun(t, vm, `process.version`)
	if val.String() != "v1.2.3" {
		t.Fatalf("expected 'v1.2.3', got %q", val.String())
	}
}

func TestProcessArgv(t *testing.T) {
	vm := newVM(t)
	mustRun(t, vm, `
		if (process.argv.length !== 4) throw new Error('wrong length: ' + process.argv.length);
		if (process.argv[0] !== 'gode') throw new Error('argv[0]: ' + process.argv[0]);
		if (process.argv[1] !== 'test.js') throw new Error('argv[1]: ' + process.argv[1]);
		if (process.argv[2] !== 'arg1') throw new Error('argv[2]: ' + process.argv[2]);
		if (process.argv[3] !== 'arg2') throw new Error('argv[3]: ' + process.argv[3]);
	`)
}

func TestProcessArgv_NoScript(t *testing.T) {
	vm := otto.New()
	Setup(vm, "v0.0.0", "", nil)
	mustRun(t, vm, `
		if (process.argv.length !== 1) throw new Error('wrong length: ' + process.argv.length);
		if (process.argv[0] !== 'gode') throw new Error('argv[0]: ' + process.argv[0]);
	`)
}

func TestProcessCwd(t *testing.T) {
	vm := newVM(t)
	val := mustRun(t, vm, `process.cwd()`)
	cwd, _ := os.Getwd()
	if val.String() != cwd {
		t.Fatalf("expected %q, got %q", cwd, val.String())
	}
}

func TestProcessEnv(t *testing.T) {
	os.Setenv("GODE_TEST_VAR", "test_value_42")
	defer os.Unsetenv("GODE_TEST_VAR")

	vm := otto.New()
	Setup(vm, "v0.0.0", "", nil)

	val := mustRun(t, vm, `process.env.GODE_TEST_VAR`)
	if val.String() != "test_value_42" {
		t.Fatalf("expected 'test_value_42', got %q", val.String())
	}
}

func TestProcessEnv_PATH(t *testing.T) {
	vm := newVM(t)
	val := mustRun(t, vm, `process.env.PATH`)
	// PATH should exist and be non-empty on any system.
	if val.String() == "" || val.String() == "undefined" {
		t.Fatal("expected process.env.PATH to be set")
	}
}

func TestProcessExit_IsFunction(t *testing.T) {
	vm := newVM(t)
	mustRun(t, vm, `
		if (typeof process.exit !== 'function') throw new Error('exit is not a function');
	`)
}

func TestProcessObject_AllProperties(t *testing.T) {
	vm := newVM(t)
	// Verify all expected properties exist.
	props := []string{"version", "argv", "exit", "cwd", "env"}
	for _, p := range props {
		t.Run(p, func(t *testing.T) {
			val := mustRun(t, vm, `typeof process.`+p)
			s := val.String()
			if s == "undefined" {
				t.Fatalf("process.%s should be defined", p)
			}
			if p == "exit" || p == "cwd" {
				if s != "function" {
					t.Fatalf("process.%s should be a function, got %s", p, s)
				}
			}
		})
	}
}

func TestProcessEnv_ContainsHOME(t *testing.T) {
	vm := newVM(t)
	val := mustRun(t, vm, `process.env.HOME`)
	home := os.Getenv("HOME")
	if !strings.Contains(val.String(), home) {
		t.Fatalf("expected HOME %q, got %q", home, val.String())
	}
}
