package process

import (
	"os"
	"runtime"
	"testing"

	"github.com/dop251/goja"
)

func testVM(t *testing.T) *goja.Runtime {
	t.Helper()
	vm := goja.New()
	Setup(vm, "v0.1.0-test", "", nil)
	return vm
}

func mustRun(t *testing.T, vm *goja.Runtime, js string) goja.Value {
	t.Helper()
	val, err := vm.RunString(js)
	if err != nil {
		t.Fatalf("unexpected error: %v\nscript: %s", err, js)
	}
	return val
}

func TestProcessVersion(t *testing.T) {
	vm := testVM(t)
	val := mustRun(t, vm, `process.version`)
	if val.String() != "v0.1.0-test" {
		t.Fatalf("expected v0.1.0-test, got %s", val.String())
	}
}

func TestProcessArgvDefault(t *testing.T) {
	vm := testVM(t)
	mustRun(t, vm, `if (process.argv[0] !== 'gode') throw new Error('argv[0] should be gode')`)
	mustRun(t, vm, `if (process.argv.length !== 1) throw new Error('default argv length should be 1')`)
}

func TestProcessArgvWithScript(t *testing.T) {
	vm := goja.New()
	Setup(vm, "v0.1.0-test", "app.js", []string{"--port", "3000"})
	mustRun(t, vm, `if (process.argv[0] !== 'gode') throw new Error('argv[0]')`)
	mustRun(t, vm, `if (process.argv[1] !== 'app.js') throw new Error('argv[1]')`)
	mustRun(t, vm, `if (process.argv[2] !== '--port') throw new Error('argv[2]')`)
	mustRun(t, vm, `if (process.argv[3] !== '3000') throw new Error('argv[3]')`)
}

func TestProcessCwd(t *testing.T) {
	vm := testVM(t)
	val := mustRun(t, vm, `process.cwd()`)
	expected, _ := os.Getwd()
	if val.String() != expected {
		t.Fatalf("expected %q, got %q", expected, val.String())
	}
}

func TestProcessEnv(t *testing.T) {
	vm := testVM(t)
	mustRun(t, vm, `if (typeof process.env !== 'object') throw new Error('env should be object')`)
	home := os.Getenv("HOME")
	if home != "" {
		val := mustRun(t, vm, `process.env.HOME`)
		if val.String() != home {
			t.Fatalf("expected HOME=%q, got %q", home, val.String())
		}
	}
}

func TestProcessEnvCustom(t *testing.T) {
	os.Setenv("GODE_TEST_VAR", "hello123")
	defer os.Unsetenv("GODE_TEST_VAR")
	vm := goja.New()
	Setup(vm, "v0.1.0-test", "", nil)
	val := mustRun(t, vm, `process.env.GODE_TEST_VAR`)
	if val.String() != "hello123" {
		t.Fatalf("expected hello123, got %s", val.String())
	}
}

func TestProcessExitType(t *testing.T) {
	vm := testVM(t)
	mustRun(t, vm, `if (typeof process.exit !== 'function') throw new Error('exit should be function')`)
}

func TestProcessProperties(t *testing.T) {
	vm := testVM(t)
	props := []string{"version", "argv", "exit", "cwd", "env"}
	for _, p := range props {
		t.Run(p, func(t *testing.T) {
			mustRun(t, vm, `if (process.`+p+` === undefined) throw new Error('missing: `+p+`')`)
		})
	}
	_ = runtime.GOOS // keep import
}
