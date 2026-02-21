package modules

import (
	"testing"

	"github.com/dop251/goja"
)

func newVM(t *testing.T) *goja.Runtime {
	t.Helper()
	vm := goja.New()
	NewLoader(vm)
	return vm
}

func mustRunA(t *testing.T, vm *goja.Runtime, js string) goja.Value {
	t.Helper()
	val, err := vm.RunString(js)
	if err != nil {
		t.Fatalf("unexpected error: %v\nscript: %s", err, js)
	}
	return val
}

func TestRequireAssert(t *testing.T) {
	vm := newVM(t)
	mustRunA(t, vm, `var assert = require('assert')`)
	mustRunA(t, vm, `if (typeof assert !== 'function') throw new Error('assert should be a function')`)
}

func TestRequireAssertNodePrefix(t *testing.T) {
	vm := newVM(t)
	mustRunA(t, vm, `var assert = require('node:assert')`)
	mustRunA(t, vm, `if (typeof assert !== 'function') throw new Error('assert should be a function')`)
}

func TestAssertOk(t *testing.T) {
	vm := newVM(t)
	mustRunA(t, vm, `var assert = require('assert')`)
	mustRunA(t, vm, `assert.ok(true)`)
	mustRunA(t, vm, `assert.ok(1)`)
	mustRunA(t, vm, `assert.ok('hello')`)
}

func TestAssertEqual(t *testing.T) {
	vm := newVM(t)
	mustRunA(t, vm, `var assert = require('assert')`)
	mustRunA(t, vm, `assert.equal(1, 1)`)
	mustRunA(t, vm, `assert.equal('a', 'a')`)
	mustRunA(t, vm, `assert.equal(1, '1')`)
}

func TestAssertStrictEqual(t *testing.T) {
	vm := newVM(t)
	mustRunA(t, vm, `var assert = require('assert')`)
	mustRunA(t, vm, `assert.strictEqual(1, 1)`)
	mustRunA(t, vm, `assert.strictEqual('a', 'a')`)
}

func TestAssertDeepEqual(t *testing.T) {
	vm := newVM(t)
	mustRunA(t, vm, `var assert = require('assert')`)
	mustRunA(t, vm, `assert.deepEqual({a: 1}, {a: 1})`)
	mustRunA(t, vm, `assert.deepEqual([1,2], [1,2])`)
}

func TestAssertDeepStrictEqual(t *testing.T) {
	vm := newVM(t)
	mustRunA(t, vm, `var assert = require('assert')`)
	mustRunA(t, vm, `assert.deepStrictEqual({a: 1, b: [2]}, {a: 1, b: [2]})`)
}

func TestAssertThrows(t *testing.T) {
	vm := newVM(t)
	mustRunA(t, vm, `var assert = require('assert')`)
	mustRunA(t, vm, `assert.throws(function() { throw new Error('boom') })`)
}

func TestAssertDoesNotThrow(t *testing.T) {
	vm := newVM(t)
	mustRunA(t, vm, `var assert = require('assert')`)
	mustRunA(t, vm, `assert.doesNotThrow(function() { return 42 })`)
}

func TestAssertFail(t *testing.T) {
	vm := newVM(t)
	mustRunA(t, vm, `var assert = require('assert')`)
	_, err := vm.RunString(`assert.fail('intentional')`)
	if err == nil {
		t.Fatal("expected assert.fail to throw")
	}
}

func TestAssertIfError(t *testing.T) {
	vm := newVM(t)
	mustRunA(t, vm, `var assert = require('assert')`)
	mustRunA(t, vm, `assert.ifError(null)`)
	mustRunA(t, vm, `assert.ifError(undefined)`)
}

func TestAssertMatch(t *testing.T) {
	vm := newVM(t)
	mustRunA(t, vm, `var assert = require('assert')`)
	mustRunA(t, vm, `assert.match('hello world', /world/)`)
}

func TestAssertDoesNotMatch(t *testing.T) {
	vm := newVM(t)
	mustRunA(t, vm, `var assert = require('assert')`)
	mustRunA(t, vm, `assert.doesNotMatch('hello', /world/)`)
}
