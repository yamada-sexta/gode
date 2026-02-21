package modules

import (
	"testing"

	"github.com/dop251/goja"
)

func bufVM(t *testing.T) *goja.Runtime {
	t.Helper()
	vm := goja.New()
	NewLoader(vm)
	mustRunB(t, vm, `var Buffer = require('buffer').Buffer`)
	return vm
}

func mustRunB(t *testing.T, vm *goja.Runtime, js string) goja.Value {
	t.Helper()
	val, err := vm.RunString(js)
	if err != nil {
		t.Fatalf("unexpected error: %v\nscript: %s", err, js)
	}
	return val
}

func TestBufferRequire(t *testing.T) {
	vm := goja.New()
	NewLoader(vm)
	mustRunB(t, vm, `var b = require('buffer'); if (!b.Buffer) throw new Error('missing Buffer')`)
}

func TestBufferAlloc(t *testing.T) {
	vm := bufVM(t)
	mustRunB(t, vm, `var b = Buffer.alloc(10); if (b.length !== 10) throw new Error('wrong length')`)
	mustRunB(t, vm, `for (var i = 0; i < 10; i++) if (b[i] !== 0) throw new Error('not zeroed')`)
}

func TestBufferFrom(t *testing.T) {
	vm := bufVM(t)
	mustRunB(t, vm, `var b = Buffer.from([1,2,3]); if (b.length !== 3) throw new Error('wrong length')`)
	mustRunB(t, vm, `if (b[0] !== 1 || b[1] !== 2 || b[2] !== 3) throw new Error('wrong values')`)
}

func TestBufferFromString(t *testing.T) {
	vm := bufVM(t)
	mustRunB(t, vm, `var b = Buffer.from('hello'); if (b.length !== 5) throw new Error('wrong length: ' + b.length)`)
	mustRunB(t, vm, `if (b.toString() !== 'hello') throw new Error('wrong string')`)
}

func TestBufferToString(t *testing.T) {
	vm := bufVM(t)
	mustRunB(t, vm, `var b = Buffer.from('hello world')`)
	mustRunB(t, vm, `if (b.toString('utf8') !== 'hello world') throw new Error('utf8 failed')`)
	mustRunB(t, vm, `if (b.toString('hex') !== '68656c6c6f20776f726c64') throw new Error('hex failed')`)
}

func TestBufferByteLength(t *testing.T) {
	vm := bufVM(t)
	mustRunB(t, vm, `if (Buffer.byteLength('hello') !== 5) throw new Error('wrong byteLength')`)
}

func TestBufferCompare(t *testing.T) {
	vm := bufVM(t)
	mustRunB(t, vm, `
		var a = Buffer.from('abc'), b = Buffer.from('abc'), c = Buffer.from('abd');
		if (Buffer.compare(a, b) !== 0) throw new Error('should be equal');
		if (Buffer.compare(a, c) >= 0) throw new Error('a < c');
	`)
}

func TestBufferConcat(t *testing.T) {
	vm := bufVM(t)
	mustRunB(t, vm, `
		var a = Buffer.from('hel'), b = Buffer.from('lo');
		var c = Buffer.concat([a, b]);
		if (c.toString() !== 'hello') throw new Error('concat failed');
	`)
}

func TestBufferCopy(t *testing.T) {
	vm := bufVM(t)
	mustRunB(t, vm, `
		var a = Buffer.from('hello'), b = Buffer.alloc(3);
		a.copy(b, 0, 0, 3);
		if (b.toString() !== 'hel') throw new Error('copy failed');
	`)
}

func TestBufferSlice(t *testing.T) {
	vm := bufVM(t)
	mustRunB(t, vm, `
		var a = Buffer.from('hello');
		var s = a.slice(1, 4);
		if (s.toString() !== 'ell') throw new Error('slice failed');
	`)
}

func TestBufferFill(t *testing.T) {
	vm := bufVM(t)
	mustRunB(t, vm, `
		var b = Buffer.alloc(5);
		b.fill(65);
		if (b.toString() !== 'AAAAA') throw new Error('fill failed');
	`)
}

func TestBufferIsBuffer(t *testing.T) {
	vm := bufVM(t)
	mustRunB(t, vm, `
		if (!Buffer.isBuffer(Buffer.alloc(1))) throw new Error('should be buffer');
		if (Buffer.isBuffer('hello')) throw new Error('string is not buffer');
	`)
}
