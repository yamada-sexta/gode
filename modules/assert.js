(function () {
  'use strict';

  // ---------------------------------------------------------------------------
  // AssertionError
  // ---------------------------------------------------------------------------

  function AssertionError(options) {
    if (!(this instanceof AssertionError)) {
      return new AssertionError(options);
    }
    options = options || {};
    this.name = 'AssertionError';
    this.actual = options.actual;
    this.expected = options.expected;
    this.operator = options.operator || '';
    this.message = options.message
      ? String(options.message)
      : _formatDefault(options.actual, options.operator, options.expected);
  }

  AssertionError.prototype = Object.create(Error.prototype);
  AssertionError.prototype.constructor = AssertionError;
  AssertionError.prototype.toString = function () {
    return this.name + ': ' + this.message;
  };

  // ---------------------------------------------------------------------------
  // Internal helpers
  // ---------------------------------------------------------------------------

  function _inspect(val) {
    if (val === null) return 'null';
    if (val === undefined) return 'undefined';
    if (typeof val === 'string') return JSON.stringify(val);
    if (typeof val === 'function') {
      return '[Function' + (val.name ? ': ' + val.name : '') + ']';
    }
    if (val instanceof RegExp) return String(val);
    if (val instanceof Date) return val.toISOString();
    try { return JSON.stringify(val); } catch (e) { return String(val); }
  }

  function _formatDefault(actual, operator, expected) {
    if (!operator) return 'Failed';
    return _inspect(actual) + ' ' + operator + ' ' + _inspect(expected);
  }

  // Deep comparison. When strict is true every leaf uses ===; when false
  // leaves use ==.  Handles primitives, Arrays, Dates, RegExps, and plain
  // Objects (by own enumerable keys).
  function _isDeepEqual(a, b, strict) {
    // Identical values / same reference.
    if (strict ? a === b : a == b) return true;

    // null / undefined mismatches.
    if (a === null || a === undefined || b === null || b === undefined) {
      return false;
    }

    // Type mismatch in strict mode.
    if (strict && typeof a !== typeof b) return false;

    // Dates.
    if (a instanceof Date && b instanceof Date) {
      return a.getTime() === b.getTime();
    }

    // RegExps.
    if (a instanceof RegExp && b instanceof RegExp) {
      return a.source === b.source &&
        a.global === b.global &&
        a.multiline === b.multiline &&
        a.ignoreCase === b.ignoreCase;
    }

    // Primitives after above checks.
    if (typeof a !== 'object' || typeof b !== 'object') {
      return strict ? a === b : a == b;
    }

    // Arrays.
    var aIsArr = Array.isArray(a);
    var bIsArr = Array.isArray(b);
    if (aIsArr !== bIsArr) return false;

    if (aIsArr) {
      if (a.length !== b.length) return false;
      for (var i = 0; i < a.length; i++) {
        if (!_isDeepEqual(a[i], b[i], strict)) return false;
      }
      return true;
    }

    // Plain objects – compare own enumerable keys.
    var aKeys = Object.keys(a).sort();
    var bKeys = Object.keys(b).sort();
    if (aKeys.length !== bKeys.length) return false;

    for (var j = 0; j < aKeys.length; j++) {
      if (aKeys[j] !== bKeys[j]) return false;
      if (!_isDeepEqual(a[aKeys[j]], b[bKeys[j]], strict)) return false;
    }
    return true;
  }

  // Partial deep strict comparison – expected is a subset of actual.
  function _isPartialDeepStrictEqual(actual, expected) {
    if (actual === expected) return true;
    if (expected === null || expected === undefined) return actual === expected;
    if (typeof expected !== 'object') return actual === expected;

    if (Array.isArray(expected)) {
      if (!Array.isArray(actual)) return false;
      for (var i = 0; i < expected.length; i++) {
        var found = false;
        for (var j = 0; j < actual.length; j++) {
          if (_isPartialDeepStrictEqual(actual[j], expected[i])) {
            found = true;
            break;
          }
        }
        if (!found) return false;
      }
      return true;
    }

    var keys = Object.keys(expected);
    for (var k = 0; k < keys.length; k++) {
      var key = keys[k];
      if (!actual.hasOwnProperty(key)) return false;
      if (!_isPartialDeepStrictEqual(actual[key], expected[key])) return false;
    }
    return true;
  }

  // ---------------------------------------------------------------------------
  // Public API
  // ---------------------------------------------------------------------------

  function ok(value, message) {
    if (!value) {
      throw new AssertionError({
        actual: value, expected: true, operator: '==',
        message: message || 'Expected value to be truthy'
      });
    }
  }

  function equal(actual, expected, message) {
    if (actual != expected) {
      throw new AssertionError({
        actual: actual, expected: expected, operator: '==', message: message
      });
    }
  }

  function notEqual(actual, expected, message) {
    if (actual == expected) {
      throw new AssertionError({
        actual: actual, expected: expected, operator: '!=', message: message
      });
    }
  }

  function strictEqual(actual, expected, message) {
    if (actual !== expected) {
      throw new AssertionError({
        actual: actual, expected: expected, operator: '===', message: message
      });
    }
  }

  function notStrictEqual(actual, expected, message) {
    if (actual === expected) {
      throw new AssertionError({
        actual: actual, expected: expected, operator: '!==', message: message
      });
    }
  }

  function deepEqual(actual, expected, message) {
    if (!_isDeepEqual(actual, expected, false)) {
      throw new AssertionError({
        actual: actual, expected: expected, operator: 'deepEqual', message: message
      });
    }
  }

  function deepStrictEqual(actual, expected, message) {
    if (!_isDeepEqual(actual, expected, true)) {
      throw new AssertionError({
        actual: actual, expected: expected, operator: 'deepStrictEqual',
        message: message
      });
    }
  }

  function notDeepEqual(actual, expected, message) {
    if (_isDeepEqual(actual, expected, false)) {
      throw new AssertionError({
        actual: actual, expected: expected, operator: 'notDeepEqual',
        message: message
      });
    }
  }

  function notDeepStrictEqual(actual, expected, message) {
    if (_isDeepEqual(actual, expected, true)) {
      throw new AssertionError({
        actual: actual, expected: expected, operator: 'notDeepStrictEqual',
        message: message
      });
    }
  }

  function fail(message) {
    throw new AssertionError({
      operator: 'fail', message: message || 'Failed'
    });
  }

  function ifError(value) {
    if (value !== null && value !== undefined) {
      if (value instanceof Error) throw value;
      throw new AssertionError({
        actual: value, expected: null, operator: 'ifError',
        message: 'ifError got unwanted exception: ' + _inspect(value)
      });
    }
  }

  function throws(fn, errorOrMessage, message) {
    var expected;
    if (typeof errorOrMessage === 'string') {
      message = errorOrMessage;
    } else {
      expected = errorOrMessage;
    }

    var thrown = false;
    var caught;
    try { fn(); } catch (e) { thrown = true; caught = e; }

    if (!thrown) {
      throw new AssertionError({
        actual: undefined, expected: expected, operator: 'throws',
        message: message || 'Missing expected exception'
      });
    }

    if (expected !== undefined && expected !== null) {
      if (expected instanceof RegExp) {
        var msg = (caught && caught.message) ? caught.message : String(caught);
        if (!expected.test(msg)) {
          throw new AssertionError({
            actual: caught, expected: expected, operator: 'throws',
            message: message || 'Error message did not match pattern'
          });
        }
      } else if (typeof expected === 'function') {
        if (!(caught instanceof expected)) {
          throw caught; // Re-throw: wrong error type.
        }
      } else if (typeof expected === 'object') {
        var keys = Object.keys(expected);
        for (var i = 0; i < keys.length; i++) {
          if (caught[keys[i]] !== expected[keys[i]]) {
            throw new AssertionError({
              actual: caught, expected: expected, operator: 'throws',
              message: message || 'Error properties did not match'
            });
          }
        }
      }
    }
  }

  function doesNotThrow(fn, errorOrMessage, message) {
    if (typeof errorOrMessage === 'string') {
      message = errorOrMessage;
    }
    try { fn(); } catch (e) {
      throw new AssertionError({
        actual: e, expected: undefined, operator: 'doesNotThrow',
        message: message || 'Got unwanted exception: ' +
          ((e && e.message) ? e.message : String(e))
      });
    }
  }

  function match(string, regexp, message) {
    if (typeof string !== 'string') {
      throw new AssertionError({
        actual: string, expected: regexp, operator: 'match',
        message: message || 'The "string" argument must be of type string'
      });
    }
    if (!(regexp instanceof RegExp)) {
      throw new AssertionError({
        actual: string, expected: regexp, operator: 'match',
        message: message || 'The "regexp" argument must be an instance of RegExp'
      });
    }
    if (!regexp.test(string)) {
      throw new AssertionError({
        actual: string, expected: regexp, operator: 'match', message: message
      });
    }
  }

  function doesNotMatch(string, regexp, message) {
    if (typeof string !== 'string') {
      throw new AssertionError({
        actual: string, expected: regexp, operator: 'doesNotMatch',
        message: message || 'The "string" argument must be of type string'
      });
    }
    if (!(regexp instanceof RegExp)) {
      throw new AssertionError({
        actual: string, expected: regexp, operator: 'doesNotMatch',
        message: message || 'The "regexp" argument must be an instance of RegExp'
      });
    }
    if (regexp.test(string)) {
      throw new AssertionError({
        actual: string, expected: regexp, operator: 'doesNotMatch',
        message: message
      });
    }
  }

  function partialDeepStrictEqual(actual, expected, message) {
    if (!_isPartialDeepStrictEqual(actual, expected)) {
      throw new AssertionError({
        actual: actual, expected: expected,
        operator: 'partialDeepStrictEqual', message: message
      });
    }
  }

  // ---------------------------------------------------------------------------
  // Build the module: assert is callable (alias for ok) and has properties.
  // ---------------------------------------------------------------------------

  var assert = function (value, message) { ok(value, message); };

  assert.ok = ok;
  assert.equal = equal;
  assert.notEqual = notEqual;
  assert.strictEqual = strictEqual;
  assert.notStrictEqual = notStrictEqual;
  assert.deepEqual = deepEqual;
  assert.deepStrictEqual = deepStrictEqual;
  assert.notDeepEqual = notDeepEqual;
  assert.notDeepStrictEqual = notDeepStrictEqual;
  assert.throws = throws;
  assert.doesNotThrow = doesNotThrow;
  assert.fail = fail;
  assert.ifError = ifError;
  assert.match = match;
  assert.doesNotMatch = doesNotMatch;
  assert.partialDeepStrictEqual = partialDeepStrictEqual;
  assert.AssertionError = AssertionError;

  return assert;
})();
