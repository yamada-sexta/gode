package modules

import (
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/robertkrimen/otto"
)

func osVM(t *testing.T) *otto.Otto {
	t.Helper()
	vm := otto.New()
	NewLoader(vm)
	mustRunO(t, vm, `var os = require('os')`)
	return vm
}

func mustRunO(t *testing.T, vm *otto.Otto, js string) otto.Value {
	t.Helper()
	val, err := vm.Run(js)
	if err != nil {
		t.Fatalf("unexpected error: %v\nscript: %s", err, js)
	}
	return val
}

// ---------------------------------------------------------------------------
// require
// ---------------------------------------------------------------------------

func TestOSRequire(t *testing.T) {
	vm := otto.New()
	NewLoader(vm)
	mustRunO(t, vm, `var o = require('os'); if (!o.hostname) throw new Error('missing hostname')`)
}

func TestOSRequireNodePrefix(t *testing.T) {
	vm := otto.New()
	NewLoader(vm)
	mustRunO(t, vm, `var o = require('node:os'); if (!o.hostname) throw new Error('missing hostname')`)
}

// ---------------------------------------------------------------------------
// EOL / devNull
// ---------------------------------------------------------------------------

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

// ---------------------------------------------------------------------------
// hostname
// ---------------------------------------------------------------------------

func TestOS_Hostname(t *testing.T) {
	vm := osVM(t)
	val := mustRunO(t, vm, `os.hostname()`)
	expected, _ := os.Hostname()
	if val.String() != expected {
		t.Fatalf("expected %q, got %q", expected, val.String())
	}
}

// ---------------------------------------------------------------------------
// homedir
// ---------------------------------------------------------------------------

func TestOS_Homedir(t *testing.T) {
	vm := osVM(t)
	val := mustRunO(t, vm, `os.homedir()`)
	expected, _ := os.UserHomeDir()
	if val.String() != expected {
		t.Fatalf("expected %q, got %q", expected, val.String())
	}
}

// ---------------------------------------------------------------------------
// tmpdir
// ---------------------------------------------------------------------------

func TestOS_Tmpdir(t *testing.T) {
	vm := osVM(t)
	val := mustRunO(t, vm, `os.tmpdir()`)
	expected := os.TempDir()
	if val.String() != expected {
		t.Fatalf("expected %q, got %q", expected, val.String())
	}
}

// ---------------------------------------------------------------------------
// platform
// ---------------------------------------------------------------------------

func TestOS_Platform(t *testing.T) {
	vm := osVM(t)
	val := mustRunO(t, vm, `os.platform()`)
	if val.String() != runtime.GOOS {
		t.Fatalf("expected %q, got %q", runtime.GOOS, val.String())
	}
}

// ---------------------------------------------------------------------------
// arch
// ---------------------------------------------------------------------------

func TestOS_Arch(t *testing.T) {
	vm := osVM(t)
	val := mustRunO(t, vm, `os.arch()`)
	expected := runtime.GOARCH
	if expected == "amd64" {
		expected = "x64"
	} else if expected == "386" {
		expected = "ia32"
	}
	if val.String() != expected {
		t.Fatalf("expected %q, got %q", expected, val.String())
	}
}

// ---------------------------------------------------------------------------
// type
// ---------------------------------------------------------------------------

func TestOS_Type(t *testing.T) {
	vm := osVM(t)
	val := mustRunO(t, vm, `os.type()`)
	if runtime.GOOS == "linux" && val.String() != "Linux" {
		t.Fatalf("expected Linux, got %q", val.String())
	}
}

// ---------------------------------------------------------------------------
// release / version / machine
// ---------------------------------------------------------------------------

func TestOS_Release(t *testing.T) {
	vm := osVM(t)
	val := mustRunO(t, vm, `os.release()`)
	if val.String() == "" || val.String() == "undefined" {
		t.Fatal("release should not be empty")
	}
}

func TestOS_Version(t *testing.T) {
	vm := osVM(t)
	val := mustRunO(t, vm, `os.version()`)
	if val.String() == "" || val.String() == "undefined" {
		t.Fatal("version should not be empty")
	}
}

func TestOS_Machine(t *testing.T) {
	vm := osVM(t)
	val := mustRunO(t, vm, `os.machine()`)
	if val.String() == "" || val.String() == "undefined" {
		t.Fatal("machine should not be empty")
	}
}

// ---------------------------------------------------------------------------
// endianness
// ---------------------------------------------------------------------------

func TestOS_Endianness(t *testing.T) {
	vm := osVM(t)
	val := mustRunO(t, vm, `os.endianness()`)
	if val.String() != "LE" && val.String() != "BE" {
		t.Fatalf("expected LE or BE, got %q", val.String())
	}
}

// ---------------------------------------------------------------------------
// uptime / freemem / totalmem
// ---------------------------------------------------------------------------

func TestOS_Uptime(t *testing.T) {
	vm := osVM(t)
	mustRunO(t, vm, `
		var u = os.uptime();
		if (typeof u !== 'number' || u <= 0) throw new Error('bad uptime: ' + u);
	`)
}

func TestOS_Freemem(t *testing.T) {
	vm := osVM(t)
	mustRunO(t, vm, `
		var f = os.freemem();
		if (typeof f !== 'number' || f <= 0) throw new Error('bad freemem: ' + f);
	`)
}

func TestOS_Totalmem(t *testing.T) {
	vm := osVM(t)
	mustRunO(t, vm, `
		var m = os.totalmem();
		if (typeof m !== 'number' || m <= 0) throw new Error('bad totalmem: ' + m);
	`)
}

// ---------------------------------------------------------------------------
// loadavg
// ---------------------------------------------------------------------------

func TestOS_Loadavg(t *testing.T) {
	vm := osVM(t)
	mustRunO(t, vm, `
		var l = os.loadavg();
		if (!Array.isArray(l) || l.length !== 3) throw new Error('bad loadavg');
		if (typeof l[0] !== 'number') throw new Error('loadavg[0] not number');
	`)
}

// ---------------------------------------------------------------------------
// availableParallelism
// ---------------------------------------------------------------------------

func TestOS_AvailableParallelism(t *testing.T) {
	vm := osVM(t)
	mustRunO(t, vm, `
		var n = os.availableParallelism();
		if (typeof n !== 'number' || n < 1) throw new Error('bad parallelism: ' + n);
	`)
}

// ---------------------------------------------------------------------------
// cpus
// ---------------------------------------------------------------------------

func TestOS_CPUs(t *testing.T) {
	vm := osVM(t)
	mustRunO(t, vm, `
		var c = os.cpus();
		if (!Array.isArray(c) || c.length === 0) throw new Error('empty cpus');
		if (typeof c[0].times !== 'object') throw new Error('missing times');
	`)
}

// ---------------------------------------------------------------------------
// networkInterfaces
// ---------------------------------------------------------------------------

func TestOS_NetworkInterfaces(t *testing.T) {
	vm := osVM(t)
	mustRunO(t, vm, `
		var ni = os.networkInterfaces();
		if (typeof ni !== 'object') throw new Error('bad networkInterfaces');
		// Should have at least loopback
		var keys = Object.keys(ni);
		if (keys.length === 0) throw new Error('no interfaces');
	`)
}

func TestOS_NetworkInterfaces_HasAddress(t *testing.T) {
	vm := osVM(t)
	mustRunO(t, vm, `
		var ni = os.networkInterfaces();
		var found = false;
		var keys = Object.keys(ni);
		for (var i = 0; i < keys.length; i++) {
			var addrs = ni[keys[i]];
			for (var j = 0; j < addrs.length; j++) {
				if (addrs[j].address && addrs[j].family) found = true;
			}
		}
		if (!found) throw new Error('no addresses found');
	`)
}

// ---------------------------------------------------------------------------
// userInfo
// ---------------------------------------------------------------------------

func TestOS_UserInfo(t *testing.T) {
	vm := osVM(t)
	val := mustRunO(t, vm, `JSON.stringify(os.userInfo())`)
	s := val.String()
	if !strings.Contains(s, "username") || !strings.Contains(s, "homedir") {
		t.Fatalf("userInfo missing fields: %s", s)
	}
}

// ---------------------------------------------------------------------------
// constants
// ---------------------------------------------------------------------------

func TestOS_Constants(t *testing.T) {
	vm := osVM(t)
	mustRunO(t, vm, `
		if (typeof os.constants !== 'object') throw new Error('missing constants');
		if (os.constants.priority.PRIORITY_NORMAL !== 0) throw new Error('wrong PRIORITY_NORMAL');
	`)
}

// ---------------------------------------------------------------------------
// All functions exist
// ---------------------------------------------------------------------------

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
