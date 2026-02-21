package modules

import (
	"strings"
	"testing"

	"github.com/robertkrimen/otto"
)

// newVM creates a fresh VM with the module loader installed.
func newVM(t *testing.T) *otto.Otto {
	t.Helper()
	vm := otto.New()
	NewLoader(vm)
	return vm
}

// mustRun evaluates js and fails the test if an error occurs.
func mustRun(t *testing.T, vm *otto.Otto, js string) otto.Value {
	t.Helper()
	val, err := vm.Run(js)
	if err != nil {
		t.Fatalf("unexpected error: %v\nscript: %s", err, js)
	}
	return val
}

// mustFail evaluates js and expects an error. Returns the error string.
func mustFail(t *testing.T, vm *otto.Otto, js string) string {
	t.Helper()
	_, err := vm.Run(js)
	if err == nil {
		t.Fatalf("expected error but got none\nscript: %s", js)
	}
	return err.Error()
}

// ---------------------------------------------------------------------------
// require()
// ---------------------------------------------------------------------------

func TestRequireAssert(t *testing.T) {
	vm := newVM(t)
	mustRun(t, vm, `var assert = require('assert')`)
	mustRun(t, vm, `if (typeof assert !== 'function') throw new Error('assert should be a function')`)
}

func TestRequireNodeAssert(t *testing.T) {
	vm := newVM(t)
	mustRun(t, vm, `var assert = require('node:assert')`)
	mustRun(t, vm, `assert.ok(true)`)
}

func TestRequireUnknownModule(t *testing.T) {
	vm := newVM(t)
	errStr := mustFail(t, vm, `require('nonexistent')`)
	if !strings.Contains(errStr, "Cannot find module") {
		t.Fatalf("expected 'Cannot find module' error, got: %s", errStr)
	}
}

func TestRequireCaching(t *testing.T) {
	vm := newVM(t)
	mustRun(t, vm, `
		var a1 = require('assert');
		var a2 = require('assert');
		if (a1 !== a2) throw new Error('require should return cached module');
	`)
}

// ---------------------------------------------------------------------------
// assert() / assert.ok()
// ---------------------------------------------------------------------------

func TestAssertOk_Passing(t *testing.T) {
	vm := newVM(t)
	mustRun(t, vm, `var assert = require('assert')`)

	cases := []string{`true`, `1`, `"hello"`, `{}`, `[]`, `function(){}`}
	for _, c := range cases {
		t.Run(c, func(t *testing.T) {
			mustRun(t, vm, `assert.ok(`+c+`)`)
			mustRun(t, vm, `assert(`+c+`)`)
		})
	}
}

func TestAssertOk_Failing(t *testing.T) {
	vm := newVM(t)
	mustRun(t, vm, `var assert = require('assert')`)

	cases := []string{`false`, `0`, `""`, `null`, `undefined`}
	for _, c := range cases {
		t.Run(c, func(t *testing.T) {
			mustFail(t, vm, `assert.ok(`+c+`)`)
			mustFail(t, vm, `assert(`+c+`)`)
		})
	}
}

func TestAssertOk_CustomMessage(t *testing.T) {
	vm := newVM(t)
	mustRun(t, vm, `var assert = require('assert')`)
	errStr := mustFail(t, vm, `assert.ok(false, 'custom msg')`)
	if !strings.Contains(errStr, "custom msg") {
		t.Fatalf("expected custom message in error, got: %s", errStr)
	}
}

// ---------------------------------------------------------------------------
// assert.equal / assert.notEqual
// ---------------------------------------------------------------------------

func TestAssertEqual_Passing(t *testing.T) {
	vm := newVM(t)
	mustRun(t, vm, `var assert = require('assert')`)

	tests := []string{
		`assert.equal(1, 1)`,
		`assert.equal('1', 1)`,          // loose: coercion
		`assert.equal(null, undefined)`, // loose
		`assert.equal(0, false)`,        // loose
	}
	for _, js := range tests {
		t.Run(js, func(t *testing.T) { mustRun(t, vm, js) })
	}
}

func TestAssertEqual_Failing(t *testing.T) {
	vm := newVM(t)
	mustRun(t, vm, `var assert = require('assert')`)
	mustFail(t, vm, `assert.equal(1, 2)`)
	mustFail(t, vm, `assert.equal('a', 'b')`)
}

func TestAssertNotEqual_Passing(t *testing.T) {
	vm := newVM(t)
	mustRun(t, vm, `var assert = require('assert')`)
	mustRun(t, vm, `assert.notEqual(1, 2)`)
}

func TestAssertNotEqual_Failing(t *testing.T) {
	vm := newVM(t)
	mustRun(t, vm, `var assert = require('assert')`)
	mustFail(t, vm, `assert.notEqual(1, 1)`)
	mustFail(t, vm, `assert.notEqual('1', 1)`) // loose coercion
}

// ---------------------------------------------------------------------------
// assert.strictEqual / assert.notStrictEqual
// ---------------------------------------------------------------------------

func TestAssertStrictEqual_Passing(t *testing.T) {
	vm := newVM(t)
	mustRun(t, vm, `var assert = require('assert')`)
	mustRun(t, vm, `assert.strictEqual(1, 1)`)
	mustRun(t, vm, `assert.strictEqual('hello', 'hello')`)
	mustRun(t, vm, `assert.strictEqual(null, null)`)
	mustRun(t, vm, `assert.strictEqual(undefined, undefined)`)
}

func TestAssertStrictEqual_Failing(t *testing.T) {
	vm := newVM(t)
	mustRun(t, vm, `var assert = require('assert')`)
	mustFail(t, vm, `assert.strictEqual(1, '1')`)          // no coercion
	mustFail(t, vm, `assert.strictEqual(null, undefined)`) // strict
	mustFail(t, vm, `assert.strictEqual(0, false)`)
}

func TestAssertNotStrictEqual_Passing(t *testing.T) {
	vm := newVM(t)
	mustRun(t, vm, `var assert = require('assert')`)
	mustRun(t, vm, `assert.notStrictEqual(1, '1')`)
	mustRun(t, vm, `assert.notStrictEqual(null, undefined)`)
}

func TestAssertNotStrictEqual_Failing(t *testing.T) {
	vm := newVM(t)
	mustRun(t, vm, `var assert = require('assert')`)
	mustFail(t, vm, `assert.notStrictEqual(1, 1)`)
}

// ---------------------------------------------------------------------------
// assert.deepEqual / assert.notDeepEqual
// ---------------------------------------------------------------------------

func TestAssertDeepEqual_Passing(t *testing.T) {
	vm := newVM(t)
	mustRun(t, vm, `var assert = require('assert')`)

	tests := []string{
		`assert.deepEqual({a: 1}, {a: 1})`,
		`assert.deepEqual([1, 2, 3], [1, 2, 3])`,
		`assert.deepEqual({a: {b: 1}}, {a: {b: 1}})`,
		`assert.deepEqual({a: '1'}, {a: 1})`, // loose: '1' == 1
		`assert.deepEqual([], [])`,
		`assert.deepEqual({}, {})`,
	}
	for _, js := range tests {
		t.Run(js, func(t *testing.T) { mustRun(t, vm, js) })
	}
}

func TestAssertDeepEqual_Failing(t *testing.T) {
	vm := newVM(t)
	mustRun(t, vm, `var assert = require('assert')`)
	mustFail(t, vm, `assert.deepEqual({a: 1}, {a: 2})`)
	mustFail(t, vm, `assert.deepEqual([1], [1, 2])`)
	mustFail(t, vm, `assert.deepEqual({a: 1}, {b: 1})`)
}

func TestAssertNotDeepEqual_Passing(t *testing.T) {
	vm := newVM(t)
	mustRun(t, vm, `var assert = require('assert')`)
	mustRun(t, vm, `assert.notDeepEqual({a: 1}, {a: 2})`)
	mustRun(t, vm, `assert.notDeepEqual([1], [2])`)
}

func TestAssertNotDeepEqual_Failing(t *testing.T) {
	vm := newVM(t)
	mustRun(t, vm, `var assert = require('assert')`)
	mustFail(t, vm, `assert.notDeepEqual({a: 1}, {a: 1})`)
}

// ---------------------------------------------------------------------------
// assert.deepStrictEqual / assert.notDeepStrictEqual
// ---------------------------------------------------------------------------

func TestAssertDeepStrictEqual_Passing(t *testing.T) {
	vm := newVM(t)
	mustRun(t, vm, `var assert = require('assert')`)

	tests := []string{
		`assert.deepStrictEqual({a: 1}, {a: 1})`,
		`assert.deepStrictEqual([1, 'two'], [1, 'two'])`,
		`assert.deepStrictEqual({nested: {x: true}}, {nested: {x: true}})`,
	}
	for _, js := range tests {
		t.Run(js, func(t *testing.T) { mustRun(t, vm, js) })
	}
}

func TestAssertDeepStrictEqual_Failing(t *testing.T) {
	vm := newVM(t)
	mustRun(t, vm, `var assert = require('assert')`)
	mustFail(t, vm, `assert.deepStrictEqual({a: '1'}, {a: 1})`) // strict: no coercion
	mustFail(t, vm, `assert.deepStrictEqual({a: 1}, {a: 1, b: 2})`)
}

func TestAssertNotDeepStrictEqual_Passing(t *testing.T) {
	vm := newVM(t)
	mustRun(t, vm, `var assert = require('assert')`)
	mustRun(t, vm, `assert.notDeepStrictEqual({a: '1'}, {a: 1})`)
}

func TestAssertNotDeepStrictEqual_Failing(t *testing.T) {
	vm := newVM(t)
	mustRun(t, vm, `var assert = require('assert')`)
	mustFail(t, vm, `assert.notDeepStrictEqual({a: 1}, {a: 1})`)
}

// ---------------------------------------------------------------------------
// Deep comparison: Dates & RegExps
// ---------------------------------------------------------------------------

func TestDeepEqual_Dates(t *testing.T) {
	vm := newVM(t)
	mustRun(t, vm, `var assert = require('assert')`)
	mustRun(t, vm, `assert.deepStrictEqual(new Date(0), new Date(0))`)
	mustFail(t, vm, `assert.deepStrictEqual(new Date(0), new Date(1))`)
}

func TestDeepEqual_RegExps(t *testing.T) {
	vm := newVM(t)
	mustRun(t, vm, `var assert = require('assert')`)
	mustRun(t, vm, `assert.deepStrictEqual(/abc/gi, /abc/gi)`)
	mustFail(t, vm, `assert.deepStrictEqual(/abc/g, /abc/i)`)
	mustFail(t, vm, `assert.deepStrictEqual(/abc/, /def/)`)
}

// ---------------------------------------------------------------------------
// assert.throws / assert.doesNotThrow
// ---------------------------------------------------------------------------

func TestAssertThrows_Passing(t *testing.T) {
	vm := newVM(t)
	mustRun(t, vm, `var assert = require('assert')`)

	// Basic: function throws
	mustRun(t, vm, `assert.throws(function() { throw new Error('boom'); })`)

	// With RegExp validator
	mustRun(t, vm, `assert.throws(function() { throw new Error('boom'); }, /boom/)`)

	// With constructor validator
	mustRun(t, vm, `assert.throws(function() { throw new TypeError('x'); }, TypeError)`)

	// With object validator
	mustRun(t, vm, `assert.throws(function() { throw {code: 42}; }, {code: 42})`)
}

func TestAssertThrows_Failing_NoThrow(t *testing.T) {
	vm := newVM(t)
	mustRun(t, vm, `var assert = require('assert')`)
	errStr := mustFail(t, vm, `assert.throws(function() {})`)
	if !strings.Contains(errStr, "Missing expected exception") {
		t.Fatalf("expected 'Missing expected exception', got: %s", errStr)
	}
}

func TestAssertThrows_Failing_WrongPattern(t *testing.T) {
	vm := newVM(t)
	mustRun(t, vm, `var assert = require('assert')`)
	mustFail(t, vm, `assert.throws(function() { throw new Error('hello'); }, /world/)`)
}

func TestAssertThrows_WithMessage(t *testing.T) {
	vm := newVM(t)
	mustRun(t, vm, `var assert = require('assert')`)
	// String as second arg is the message, not a validator
	errStr := mustFail(t, vm, `assert.throws(function() {}, 'should have thrown')`)
	if !strings.Contains(errStr, "should have thrown") {
		t.Fatalf("expected custom message, got: %s", errStr)
	}
}

func TestAssertDoesNotThrow_Passing(t *testing.T) {
	vm := newVM(t)
	mustRun(t, vm, `var assert = require('assert')`)
	mustRun(t, vm, `assert.doesNotThrow(function() { return 1; })`)
}

func TestAssertDoesNotThrow_Failing(t *testing.T) {
	vm := newVM(t)
	mustRun(t, vm, `var assert = require('assert')`)
	errStr := mustFail(t, vm, `assert.doesNotThrow(function() { throw new Error('oops'); })`)
	if !strings.Contains(errStr, "oops") {
		t.Fatalf("expected error message 'oops', got: %s", errStr)
	}
}

// ---------------------------------------------------------------------------
// assert.fail
// ---------------------------------------------------------------------------

func TestAssertFail(t *testing.T) {
	vm := newVM(t)
	mustRun(t, vm, `var assert = require('assert')`)
	mustFail(t, vm, `assert.fail()`)
}

func TestAssertFail_CustomMessage(t *testing.T) {
	vm := newVM(t)
	mustRun(t, vm, `var assert = require('assert')`)
	errStr := mustFail(t, vm, `assert.fail('intentional')`)
	if !strings.Contains(errStr, "intentional") {
		t.Fatalf("expected 'intentional', got: %s", errStr)
	}
}

// ---------------------------------------------------------------------------
// assert.ifError
// ---------------------------------------------------------------------------

func TestAssertIfError_Passing(t *testing.T) {
	vm := newVM(t)
	mustRun(t, vm, `var assert = require('assert')`)
	mustRun(t, vm, `assert.ifError(null)`)
	mustRun(t, vm, `assert.ifError(undefined)`)
}

func TestAssertIfError_Failing(t *testing.T) {
	vm := newVM(t)
	mustRun(t, vm, `var assert = require('assert')`)
	mustFail(t, vm, `assert.ifError(new Error('err'))`)
	mustFail(t, vm, `assert.ifError('string error')`)
	mustFail(t, vm, `assert.ifError(1)`)
}

// ---------------------------------------------------------------------------
// assert.match / assert.doesNotMatch
// ---------------------------------------------------------------------------

func TestAssertMatch_Passing(t *testing.T) {
	vm := newVM(t)
	mustRun(t, vm, `var assert = require('assert')`)
	mustRun(t, vm, `assert.match('hello world', /hello/)`)
	mustRun(t, vm, `assert.match('abc123', /\d+/)`)
}

func TestAssertMatch_Failing(t *testing.T) {
	vm := newVM(t)
	mustRun(t, vm, `var assert = require('assert')`)
	mustFail(t, vm, `assert.match('hello', /world/)`)
}

func TestAssertMatch_InvalidArgs(t *testing.T) {
	vm := newVM(t)
	mustRun(t, vm, `var assert = require('assert')`)
	// Non-string first arg
	errStr := mustFail(t, vm, `assert.match(123, /x/)`)
	if !strings.Contains(errStr, "string") {
		t.Fatalf("expected type error about string, got: %s", errStr)
	}
	// Non-regexp second arg
	errStr = mustFail(t, vm, `assert.match('hello', 'hello')`)
	if !strings.Contains(errStr, "RegExp") {
		t.Fatalf("expected type error about RegExp, got: %s", errStr)
	}
}

func TestAssertDoesNotMatch_Passing(t *testing.T) {
	vm := newVM(t)
	mustRun(t, vm, `var assert = require('assert')`)
	mustRun(t, vm, `assert.doesNotMatch('hello', /world/)`)
}

func TestAssertDoesNotMatch_Failing(t *testing.T) {
	vm := newVM(t)
	mustRun(t, vm, `var assert = require('assert')`)
	mustFail(t, vm, `assert.doesNotMatch('hello world', /hello/)`)
}

// ---------------------------------------------------------------------------
// assert.partialDeepStrictEqual
// ---------------------------------------------------------------------------

func TestPartialDeepStrictEqual_Passing(t *testing.T) {
	vm := newVM(t)
	mustRun(t, vm, `var assert = require('assert')`)

	tests := []string{
		// Subset of keys
		`assert.partialDeepStrictEqual({a: 1, b: 2, c: 3}, {a: 1, c: 3})`,
		// Nested subset
		`assert.partialDeepStrictEqual({x: {a: 1, b: 2}}, {x: {a: 1}})`,
		// Array contains expected elements
		`assert.partialDeepStrictEqual([1, 2, 3], [1, 3])`,
		// Exact match also passes
		`assert.partialDeepStrictEqual({a: 1}, {a: 1})`,
	}
	for _, js := range tests {
		t.Run(js, func(t *testing.T) { mustRun(t, vm, js) })
	}
}

func TestPartialDeepStrictEqual_Failing(t *testing.T) {
	vm := newVM(t)
	mustRun(t, vm, `var assert = require('assert')`)
	mustFail(t, vm, `assert.partialDeepStrictEqual({a: 1}, {a: 2})`)
	mustFail(t, vm, `assert.partialDeepStrictEqual({a: 1}, {b: 1})`)
	mustFail(t, vm, `assert.partialDeepStrictEqual([1, 2], [3])`)
}

// ---------------------------------------------------------------------------
// AssertionError
// ---------------------------------------------------------------------------

func TestAssertionError_Properties(t *testing.T) {
	vm := newVM(t)
	mustRun(t, vm, `var assert = require('assert')`)
	mustRun(t, vm, `
		try {
			assert.strictEqual(1, 2);
		} catch (e) {
			if (e.name !== 'AssertionError') throw new Error('wrong name: ' + e.name);
			if (e.actual !== 1) throw new Error('wrong actual: ' + e.actual);
			if (e.expected !== 2) throw new Error('wrong expected: ' + e.expected);
			if (e.operator !== '===') throw new Error('wrong operator: ' + e.operator);
		}
	`)
}

func TestAssertionError_IsInstanceOfError(t *testing.T) {
	vm := newVM(t)
	mustRun(t, vm, `var assert = require('assert')`)
	mustRun(t, vm, `
		try {
			assert.fail('test');
		} catch (e) {
			if (!(e instanceof Error)) throw new Error('should be instanceof Error');
			if (!(e instanceof assert.AssertionError)) throw new Error('should be instanceof AssertionError');
		}
	`)
}

func TestAssertionError_ToString(t *testing.T) {
	vm := newVM(t)
	mustRun(t, vm, `var assert = require('assert')`)
	mustRun(t, vm, `
		try {
			assert.fail('my message');
		} catch (e) {
			var str = e.toString();
			if (str.indexOf('AssertionError') === -1) throw new Error('toString missing name: ' + str);
			if (str.indexOf('my message') === -1) throw new Error('toString missing message: ' + str);
		}
	`)
}

// ---------------------------------------------------------------------------
// Deep comparison: nested arrays and mixed structures
// ---------------------------------------------------------------------------

func TestDeepEqual_NestedArrays(t *testing.T) {
	vm := newVM(t)
	mustRun(t, vm, `var assert = require('assert')`)
	mustRun(t, vm, `assert.deepStrictEqual([[1, 2], [3, 4]], [[1, 2], [3, 4]])`)
	mustFail(t, vm, `assert.deepStrictEqual([[1, 2], [3, 4]], [[1, 2], [3, 5]])`)
}

func TestDeepEqual_MixedStructures(t *testing.T) {
	vm := newVM(t)
	mustRun(t, vm, `var assert = require('assert')`)
	mustRun(t, vm, `assert.deepStrictEqual(
		{users: [{name: 'a', age: 1}, {name: 'b', age: 2}]},
		{users: [{name: 'a', age: 1}, {name: 'b', age: 2}]}
	)`)
	mustFail(t, vm, `assert.deepStrictEqual(
		{users: [{name: 'a'}]},
		{users: [{name: 'b'}]}
	)`)
}

// ---------------------------------------------------------------------------
// Edge cases
// ---------------------------------------------------------------------------

func TestAssertEqual_NaN(t *testing.T) {
	vm := newVM(t)
	mustRun(t, vm, `var assert = require('assert')`)
	// NaN != NaN in JS, so equal(NaN, NaN) should fail
	mustFail(t, vm, `assert.equal(NaN, NaN)`)
	// And notEqual(NaN, NaN) should pass
	mustRun(t, vm, `assert.notEqual(NaN, NaN)`)
}

func TestAssertDeepEqual_EmptyVsNonEmpty(t *testing.T) {
	vm := newVM(t)
	mustRun(t, vm, `var assert = require('assert')`)
	mustFail(t, vm, `assert.deepStrictEqual({}, {a: 1})`)
	mustFail(t, vm, `assert.deepStrictEqual([], [1])`)
}
