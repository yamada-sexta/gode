package modules

import (
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/dop251/goja"
)

func osVM(t *testing.T) *goja.Runtime {
	t.Helper()
	vm := goja.New()
	NewLoader(vm)
	mustRunO(t, vm, `var os = require('os')`)
	return vm
}

func mustRunO(t *testing.T, vm *goja.Runtime, js string) goja.Value {
	t.Helper()
	val, err := vm.RunString(js)
	if err != nil {
		t.Fatalf("unexpected error: %v\nscript: %s", err, js)
	}
	return val
}

func TestOSRequire(t *testing.T) {
	vm := goja.New()
	NewLoader(vm)
	mustRunO(t, vm, `var o = require('os'); if (!o.hostname) throw new Error('missing hostname')`)
}

func TestOS_EOL(t *testing.T) {
	vm := osVM(t)
	val := mustRunO(t, vm, `os.EOL`)
	if val.String() != "\n" {
		t.Fatalf("expected \\n, got %q", val.String())
	}
}

func TestOS_DevNull(t *testing.T) {
	vm := osVM(t)
	val := mustRunO(t, vm, `os.devNull`)
	if val.String() != "/dev/null" {
		t.Fatalf("expected /dev/null, got %q", val.String())
	}
}

func TestOS_Hostname(t *testing.T) {
	vm := osVM(t)
	val := mustRunO(t, vm, `os.hostname()`)
	expected, _ := os.Hostname()
	if val.String() != expected {
		t.Fatalf("expected %q, got %q", expected, val.String())
	}
}

func TestOS_Homedir(t *testing.T) {
	vm := osVM(t)
	val := mustRunO(t, vm, `os.homedir()`)
	expected, _ := os.UserHomeDir()
	if val.String() != expected {
		t.Fatalf("expected %q, got %q", expected, val.String())
	}
}

func TestOS_Platform(t *testing.T) {
	vm := osVM(t)
	val := mustRunO(t, vm, `os.platform()`)
	if val.String() != runtime.GOOS {
		t.Fatalf("expected %q, got %q", runtime.GOOS, val.String())
	}
}

func TestOS_Arch(t *testing.T) {
	vm := osVM(t)
	val := mustRunO(t, vm, `os.arch()`)
	expected := runtime.GOARCH
	if expected == "amd64" {
		expected = "x64"
	}
	if val.String() != expected {
		t.Fatalf("expected %q, got %q", expected, val.String())
	}
}

func TestOS_Uptime(t *testing.T) {
	vm := osVM(t)
	mustRunO(t, vm, `
		var u = os.uptime();
		if (typeof u !== 'number' || u <= 0) throw new Error('bad uptime: ' + u);
	`)
}

func TestOS_Loadavg(t *testing.T) {
	vm := osVM(t)
	mustRunO(t, vm, `
		var l = os.loadavg();
		if (!Array.isArray(l) || l.length !== 3) throw new Error('bad loadavg');
	`)
}

func TestOS_CPUs(t *testing.T) {
	vm := osVM(t)
	mustRunO(t, vm, `
		var c = os.cpus();
		if (!Array.isArray(c) || c.length === 0) throw new Error('empty cpus');
	`)
}

func TestOS_NetworkInterfaces(t *testing.T) {
	vm := osVM(t)
	mustRunO(t, vm, `
		var ni = os.networkInterfaces();
		if (typeof ni !== 'object') throw new Error('bad networkInterfaces');
	`)
}

func TestOS_UserInfo(t *testing.T) {
	vm := osVM(t)
	val := mustRunO(t, vm, `JSON.stringify(os.userInfo())`)
	s := val.String()
	if !strings.Contains(s, "username") || !strings.Contains(s, "homedir") {
		t.Fatalf("userInfo missing fields: %s", s)
	}
}

func TestOS_Constants(t *testing.T) {
	vm := osVM(t)
	mustRunO(t, vm, `
		if (typeof os.constants !== 'object') throw new Error('missing constants');
		if (os.constants.priority.PRIORITY_NORMAL !== 0) throw new Error('wrong PRIORITY_NORMAL');
	`)
}

func TestOS_AllFunctionsExist(t *testing.T) {
	vm := osVM(t)
	fns := []string{
		"hostname", "homedir", "tmpdir", "platform", "arch", "type",
		"release", "version", "machine", "endianness", "uptime",
		"freemem", "totalmem", "availableParallelism", "loadavg",
		"cpus", "networkInterfaces", "userInfo",
	}
	for _, fn := range fns {
		t.Run(fn, func(t *testing.T) {
			mustRunO(t, vm, `if (typeof os.`+fn+` !== 'function') throw new Error('missing: `+fn+`')`)
		})
	}
}
