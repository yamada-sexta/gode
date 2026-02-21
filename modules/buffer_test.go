package modules

import (
	"strings"
	"testing"

	"github.com/robertkrimen/otto"
)

func bufferVM(t *testing.T) *otto.Otto {
	t.Helper()
	vm := otto.New()
	NewLoader(vm)
	mustRunB(t, vm, `var buf = require('buffer'); var Buffer = buf.Buffer;`)
	return vm
}

func mustRunB(t *testing.T, vm *otto.Otto, js string) otto.Value {
	t.Helper()
	val, err := vm.Run(js)
	if err != nil {
		t.Fatalf("unexpected error: %v\nscript: %s", err, js)
	}
	return val
}

func mustFailB(t *testing.T, vm *otto.Otto, js string) string {
	t.Helper()
	_, err := vm.Run(js)
	if err == nil {
		t.Fatalf("expected error but got none\nscript: %s", js)
	}
	return err.Error()
}

// ---------------------------------------------------------------------------
// require('buffer') / require('node:buffer')
// ---------------------------------------------------------------------------

func TestBufferRequire(t *testing.T) {
	vm := otto.New()
	NewLoader(vm)
	mustRunB(t, vm, `var b = require('buffer'); if (!b.Buffer) throw new Error('missing Buffer')`)
}

func TestBufferRequireNodePrefix(t *testing.T) {
	vm := otto.New()
	NewLoader(vm)
	mustRunB(t, vm, `var b = require('node:buffer'); if (!b.Buffer) throw new Error('missing Buffer')`)
}

// ---------------------------------------------------------------------------
// Buffer.alloc
// ---------------------------------------------------------------------------

func TestBufferAlloc_ZeroFilled(t *testing.T) {
	vm := bufferVM(t)
	mustRunB(t, vm, `
		var b = Buffer.alloc(5);
		if (b.length !== 5) throw new Error('wrong length: ' + b.length);
		for (var i = 0; i < 5; i++) if (b[i] !== 0) throw new Error('not zero at ' + i);
	`)
}

func TestBufferAlloc_WithFill(t *testing.T) {
	vm := bufferVM(t)
	mustRunB(t, vm, `
		var b = Buffer.alloc(3, 0x41);
		if (b.toString() !== 'AAA') throw new Error('expected AAA, got: ' + b.toString());
	`)
}

func TestBufferAlloc_WithStringFill(t *testing.T) {
	vm := bufferVM(t)
	mustRunB(t, vm, `
		var b = Buffer.alloc(5, 'ab');
		if (b.toString() !== 'ababa') throw new Error('expected ababa, got: ' + b.toString());
	`)
}

// ---------------------------------------------------------------------------
// Buffer.from
// ---------------------------------------------------------------------------

func TestBufferFrom_Array(t *testing.T) {
	vm := bufferVM(t)
	mustRunB(t, vm, `
		var b = Buffer.from([0x68, 0x65, 0x6c, 0x6c, 0x6f]);
		if (b.toString() !== 'hello') throw new Error('expected hello');
		if (b.length !== 5) throw new Error('wrong length');
	`)
}

func TestBufferFrom_String_UTF8(t *testing.T) {
	vm := bufferVM(t)
	mustRunB(t, vm, `
		var b = Buffer.from('hello');
		if (b.length !== 5) throw new Error('wrong length');
		if (b[0] !== 0x68) throw new Error('wrong byte 0');
	`)
}

func TestBufferFrom_String_Hex(t *testing.T) {
	vm := bufferVM(t)
	mustRunB(t, vm, `
		var b = Buffer.from('68656c6c6f', 'hex');
		if (b.toString() !== 'hello') throw new Error('expected hello');
	`)
}

func TestBufferFrom_String_Base64(t *testing.T) {
	vm := bufferVM(t)
	mustRunB(t, vm, `
		var b = Buffer.from('aGVsbG8=', 'base64');
		if (b.toString() !== 'hello') throw new Error('expected hello');
	`)
}

func TestBufferFrom_String_Latin1(t *testing.T) {
	vm := bufferVM(t)
	mustRunB(t, vm, `
		var b = Buffer.from('hello', 'latin1');
		if (b.length !== 5) throw new Error('wrong length');
		if (b.toString('latin1') !== 'hello') throw new Error('roundtrip failed');
	`)
}

func TestBufferFrom_Buffer(t *testing.T) {
	vm := bufferVM(t)
	mustRunB(t, vm, `
		var a = Buffer.from('hello');
		var b = Buffer.from(a);
		if (!a.equals(b)) throw new Error('should be equal');
		a[0] = 0; // modify original
		if (b[0] === 0) throw new Error('should be a copy');
	`)
}

// ---------------------------------------------------------------------------
// Buffer.isBuffer / Buffer.isEncoding
// ---------------------------------------------------------------------------

func TestBufferIsBuffer(t *testing.T) {
	vm := bufferVM(t)
	mustRunB(t, vm, `
		if (!Buffer.isBuffer(Buffer.alloc(0))) throw new Error('should be true');
		if (Buffer.isBuffer('hello')) throw new Error('should be false for string');
		if (Buffer.isBuffer(null)) throw new Error('should be false for null');
	`)
}

func TestBufferIsEncoding(t *testing.T) {
	vm := bufferVM(t)
	mustRunB(t, vm, `
		var valid = ['utf8', 'utf-8', 'ascii', 'latin1', 'binary', 'hex', 'base64', 'base64url', 'ucs2', 'utf16le'];
		for (var i = 0; i < valid.length; i++) {
			if (!Buffer.isEncoding(valid[i])) throw new Error('should support: ' + valid[i]);
		}
		if (Buffer.isEncoding('nope')) throw new Error('should not support nope');
	`)
}

// ---------------------------------------------------------------------------
// toString with encodings
// ---------------------------------------------------------------------------

func TestBufferToString_Hex(t *testing.T) {
	vm := bufferVM(t)
	mustRunB(t, vm, `
		var b = Buffer.from('hello');
		if (b.toString('hex') !== '68656c6c6f') throw new Error('wrong hex: ' + b.toString('hex'));
	`)
}

func TestBufferToString_Base64(t *testing.T) {
	vm := bufferVM(t)
	mustRunB(t, vm, `
		var b = Buffer.from('hello');
		if (b.toString('base64') !== 'aGVsbG8=') throw new Error('wrong base64: ' + b.toString('base64'));
	`)
}

func TestBufferToString_Base64url(t *testing.T) {
	vm := bufferVM(t)
	mustRunB(t, vm, `
		var b = Buffer.from([0xFB, 0xFF, 0xFE]);
		var b64 = b.toString('base64url');
		if (b64.indexOf('+') !== -1 || b64.indexOf('/') !== -1 || b64.indexOf('=') !== -1) {
			throw new Error('base64url should not contain +/=: ' + b64);
		}
	`)
}

func TestBufferToString_UTF16LE(t *testing.T) {
	vm := bufferVM(t)
	mustRunB(t, vm, `
		var b = Buffer.from('hi', 'utf16le');
		if (b.length !== 4) throw new Error('wrong length');
		if (b.toString('utf16le') !== 'hi') throw new Error('roundtrip failed');
	`)
}

func TestBufferToString_Range(t *testing.T) {
	vm := bufferVM(t)
	mustRunB(t, vm, `
		var b = Buffer.from('hello world');
		if (b.toString('utf8', 0, 5) !== 'hello') throw new Error('wrong range');
	`)
}

// ---------------------------------------------------------------------------
// Buffer.byteLength
// ---------------------------------------------------------------------------

func TestBufferByteLength(t *testing.T) {
	vm := bufferVM(t)
	mustRunB(t, vm, `
		if (Buffer.byteLength('hello') !== 5) throw new Error('wrong utf8 length');
		if (Buffer.byteLength('68656c6c6f', 'hex') !== 5) throw new Error('wrong hex length');
		if (Buffer.byteLength('aGVsbG8=', 'base64') !== 5) throw new Error('wrong base64 length');
	`)
}

// ---------------------------------------------------------------------------
// Buffer.compare / buf.compare / buf.equals
// ---------------------------------------------------------------------------

func TestBufferCompare(t *testing.T) {
	vm := bufferVM(t)
	mustRunB(t, vm, `
		var a = Buffer.from('abc');
		var b = Buffer.from('abc');
		var c = Buffer.from('abd');
		if (Buffer.compare(a, b) !== 0) throw new Error('a === b');
		if (Buffer.compare(a, c) >= 0) throw new Error('a < c');
		if (Buffer.compare(c, a) <= 0) throw new Error('c > a');
	`)
}

func TestBufferEquals(t *testing.T) {
	vm := bufferVM(t)
	mustRunB(t, vm, `
		var a = Buffer.from('hello');
		var b = Buffer.from('hello');
		var c = Buffer.from('world');
		if (!a.equals(b)) throw new Error('should be equal');
		if (a.equals(c)) throw new Error('should not be equal');
	`)
}

// ---------------------------------------------------------------------------
// Buffer.concat
// ---------------------------------------------------------------------------

func TestBufferConcat(t *testing.T) {
	vm := bufferVM(t)
	mustRunB(t, vm, `
		var a = Buffer.from('hello');
		var b = Buffer.from(' world');
		var c = Buffer.concat([a, b]);
		if (c.toString() !== 'hello world') throw new Error('wrong concat: ' + c.toString());
		if (c.length !== 11) throw new Error('wrong length: ' + c.length);
	`)
}

func TestBufferConcat_WithLength(t *testing.T) {
	vm := bufferVM(t)
	mustRunB(t, vm, `
		var c = Buffer.concat([Buffer.from('hello'), Buffer.from(' world')], 5);
		if (c.toString() !== 'hello') throw new Error('wrong truncated concat');
	`)
}

// ---------------------------------------------------------------------------
// copy / slice / fill
// ---------------------------------------------------------------------------

func TestBufferCopy(t *testing.T) {
	vm := bufferVM(t)
	mustRunB(t, vm, `
		var a = Buffer.from('hello');
		var b = Buffer.alloc(3);
		a.copy(b, 0, 1, 4);
		if (b.toString() !== 'ell') throw new Error('wrong copy: ' + b.toString());
	`)
}

func TestBufferSlice(t *testing.T) {
	vm := bufferVM(t)
	mustRunB(t, vm, `
		var a = Buffer.from('hello world');
		var b = a.slice(6);
		if (b.toString() !== 'world') throw new Error('wrong slice: ' + b.toString());
	`)
}

func TestBufferFill(t *testing.T) {
	vm := bufferVM(t)
	mustRunB(t, vm, `
		var b = Buffer.alloc(5);
		b.fill(0x41);
		if (b.toString() !== 'AAAAA') throw new Error('wrong fill');
	`)
}

// ---------------------------------------------------------------------------
// indexOf / lastIndexOf / includes
// ---------------------------------------------------------------------------

func TestBufferIndexOf(t *testing.T) {
	vm := bufferVM(t)
	mustRunB(t, vm, `
		var b = Buffer.from('hello world hello');
		if (b.indexOf('world') !== 6) throw new Error('wrong indexOf');
		if (b.lastIndexOf('hello') !== 12) throw new Error('wrong lastIndexOf');
		if (!b.includes('world')) throw new Error('should include world');
		if (b.includes('xyz')) throw new Error('should not include xyz');
	`)
}

func TestBufferIndexOf_Byte(t *testing.T) {
	vm := bufferVM(t)
	mustRunB(t, vm, `
		var b = Buffer.from([1, 2, 3, 4]);
		if (b.indexOf(3) !== 2) throw new Error('wrong byte indexOf');
	`)
}

// ---------------------------------------------------------------------------
// write
// ---------------------------------------------------------------------------

func TestBufferWrite(t *testing.T) {
	vm := bufferVM(t)
	mustRunB(t, vm, `
		var b = Buffer.alloc(10);
		var n = b.write('hello');
		if (n !== 5) throw new Error('wrong bytes written: ' + n);
		if (b.toString('utf8', 0, 5) !== 'hello') throw new Error('wrong content');
	`)
}

// ---------------------------------------------------------------------------
// Read/Write integers
// ---------------------------------------------------------------------------

func TestBufferReadWriteUInt8(t *testing.T) {
	vm := bufferVM(t)
	mustRunB(t, vm, `
		var b = Buffer.alloc(1);
		b.writeUInt8(255, 0);
		if (b.readUInt8(0) !== 255) throw new Error('wrong uint8');
	`)
}

func TestBufferReadWriteUInt16(t *testing.T) {
	vm := bufferVM(t)
	mustRunB(t, vm, `
		var b = Buffer.alloc(4);
		b.writeUInt16BE(0x1234, 0);
		b.writeUInt16LE(0x5678, 2);
		if (b.readUInt16BE(0) !== 0x1234) throw new Error('wrong uint16be');
		if (b.readUInt16LE(2) !== 0x5678) throw new Error('wrong uint16le');
	`)
}

func TestBufferReadWriteUInt32(t *testing.T) {
	vm := bufferVM(t)
	mustRunB(t, vm, `
		var b = Buffer.alloc(8);
		b.writeUInt32BE(0xDEADBEEF, 0);
		b.writeUInt32LE(0xCAFEBABE, 4);
		if (b.readUInt32BE(0) !== 0xDEADBEEF) throw new Error('wrong uint32be: ' + b.readUInt32BE(0));
		if (b.readUInt32LE(4) !== 0xCAFEBABE) throw new Error('wrong uint32le: ' + b.readUInt32LE(4));
	`)
}

func TestBufferReadWriteInt8(t *testing.T) {
	vm := bufferVM(t)
	mustRunB(t, vm, `
		var b = Buffer.alloc(2);
		b.writeInt8(-1, 0);
		b.writeInt8(127, 1);
		if (b.readInt8(0) !== -1) throw new Error('wrong int8: ' + b.readInt8(0));
		if (b.readInt8(1) !== 127) throw new Error('wrong int8: ' + b.readInt8(1));
	`)
}

func TestBufferReadWriteInt16(t *testing.T) {
	vm := bufferVM(t)
	mustRunB(t, vm, `
		var b = Buffer.alloc(4);
		b.writeInt16BE(-256, 0);
		b.writeInt16LE(-256, 2);
		if (b.readInt16BE(0) !== -256) throw new Error('wrong int16be');
		if (b.readInt16LE(2) !== -256) throw new Error('wrong int16le');
	`)
}

func TestBufferReadWriteInt32(t *testing.T) {
	vm := bufferVM(t)
	mustRunB(t, vm, `
		var b = Buffer.alloc(8);
		b.writeInt32BE(-123456789, 0);
		b.writeInt32LE(-123456789, 4);
		if (b.readInt32BE(0) !== -123456789) throw new Error('wrong int32be');
		if (b.readInt32LE(4) !== -123456789) throw new Error('wrong int32le');
	`)
}

func TestBufferReadWriteUIntBE_LE(t *testing.T) {
	vm := bufferVM(t)
	mustRunB(t, vm, `
		var b = Buffer.alloc(6);
		b.writeUIntBE(0x010203, 0, 3);
		if (b.readUIntBE(0, 3) !== 0x010203) throw new Error('wrong uintbe 3');
		b.writeUIntLE(0x040506, 3, 3);
		if (b.readUIntLE(3, 3) !== 0x040506) throw new Error('wrong uintle 3');
	`)
}

func TestBufferReadWriteIntBE_LE(t *testing.T) {
	vm := bufferVM(t)
	mustRunB(t, vm, `
		var b = Buffer.alloc(6);
		b.writeIntBE(-1000, 0, 3);
		if (b.readIntBE(0, 3) !== -1000) throw new Error('wrong intbe');
		b.writeIntLE(-1000, 3, 3);
		if (b.readIntLE(3, 3) !== -1000) throw new Error('wrong intle');
	`)
}

// ---------------------------------------------------------------------------
// Read/Write floats
// ---------------------------------------------------------------------------

func TestBufferReadWriteFloat(t *testing.T) {
	vm := bufferVM(t)
	mustRunB(t, vm, `
		var b = Buffer.alloc(8);
		b.writeFloatBE(1.5, 0);
		b.writeFloatLE(1.5, 4);
		var be = b.readFloatBE(0);
		var le = b.readFloatLE(4);
		if (Math.abs(be - 1.5) > 0.001) throw new Error('wrong floatBE: ' + be);
		if (Math.abs(le - 1.5) > 0.001) throw new Error('wrong floatLE: ' + le);
	`)
}

func TestBufferReadWriteDouble(t *testing.T) {
	vm := bufferVM(t)
	mustRunB(t, vm, `
		var b = Buffer.alloc(16);
		b.writeDoubleBE(3.141592653589793, 0);
		b.writeDoubleLE(2.718281828459045, 8);
		var pi = b.readDoubleBE(0);
		var e  = b.readDoubleLE(8);
		if (Math.abs(pi - 3.141592653589793) > 1e-10) throw new Error('wrong doubleBE: ' + pi);
		if (Math.abs(e - 2.718281828459045) > 1e-10) throw new Error('wrong doubleLE: ' + e);
	`)
}

func TestBufferFloat_SpecialValues(t *testing.T) {
	vm := bufferVM(t)
	mustRunB(t, vm, `
		var b = Buffer.alloc(8);
		b.writeDoubleBE(Infinity, 0);
		if (b.readDoubleBE(0) !== Infinity) throw new Error('inf roundtrip');
		b.writeDoubleBE(-Infinity, 0);
		if (b.readDoubleBE(0) !== -Infinity) throw new Error('-inf roundtrip');
		b.writeDoubleBE(NaN, 0);
		if (!isNaN(b.readDoubleBE(0))) throw new Error('NaN roundtrip');
		b.writeDoubleBE(0, 0);
		if (b.readDoubleBE(0) !== 0) throw new Error('zero roundtrip');
	`)
}

// ---------------------------------------------------------------------------
// swap
// ---------------------------------------------------------------------------

func TestBufferSwap16(t *testing.T) {
	vm := bufferVM(t)
	mustRunB(t, vm, `
		var b = Buffer.from([0x01, 0x02, 0x03, 0x04]);
		b.swap16();
		if (b[0] !== 2 || b[1] !== 1 || b[2] !== 4 || b[3] !== 3) throw new Error('wrong swap16');
	`)
}

func TestBufferSwap32(t *testing.T) {
	vm := bufferVM(t)
	mustRunB(t, vm, `
		var b = Buffer.from([1, 2, 3, 4]);
		b.swap32();
		if (b[0] !== 4 || b[1] !== 3 || b[2] !== 2 || b[3] !== 1) throw new Error('wrong swap32');
	`)
}

// ---------------------------------------------------------------------------
// toJSON
// ---------------------------------------------------------------------------

func TestBufferToJSON(t *testing.T) {
	vm := bufferVM(t)
	mustRunB(t, vm, `
		var b = Buffer.from([1, 2, 3]);
		var j = b.toJSON();
		if (j.type !== 'Buffer') throw new Error('wrong type');
		if (j.data.length !== 3) throw new Error('wrong data length');
		if (j.data[0] !== 1 || j.data[1] !== 2 || j.data[2] !== 3) throw new Error('wrong data');
	`)
}

// ---------------------------------------------------------------------------
// keys / values / entries
// ---------------------------------------------------------------------------

func TestBufferIteration(t *testing.T) {
	vm := bufferVM(t)
	mustRunB(t, vm, `
		var b = Buffer.from([10, 20, 30]);
		var k = b.keys();
		if (k[0] !== 0 || k[1] !== 1 || k[2] !== 2) throw new Error('wrong keys');
		var v = b.values();
		if (v[0] !== 10 || v[1] !== 20 || v[2] !== 30) throw new Error('wrong values');
		var e = b.entries();
		if (e[0][0] !== 0 || e[0][1] !== 10) throw new Error('wrong entry');
	`)
}

// ---------------------------------------------------------------------------
// Index access
// ---------------------------------------------------------------------------

func TestBufferIndexAccess(t *testing.T) {
	vm := bufferVM(t)
	mustRunB(t, vm, `
		var b = Buffer.from([1, 2, 3]);
		if (b[0] !== 1 || b[1] !== 2 || b[2] !== 3) throw new Error('wrong index read');
		b[1] = 42;
		if (b[1] !== 42) throw new Error('index write failed');
		b[0] = 300; // should be masked to 44
		if (b[0] !== 44) throw new Error('byte masking failed: ' + b[0]);
	`)
}

// ---------------------------------------------------------------------------
// Module-level APIs
// ---------------------------------------------------------------------------

func TestBufferAtob(t *testing.T) {
	vm := bufferVM(t)
	mustRunB(t, vm, `
		if (buf.atob('aGVsbG8=') !== 'hello') throw new Error('atob failed');
	`)
}

func TestBufferBtoa(t *testing.T) {
	vm := bufferVM(t)
	mustRunB(t, vm, `
		if (buf.btoa('hello') !== 'aGVsbG8=') throw new Error('btoa failed: ' + buf.btoa('hello'));
	`)
}

func TestBufferIsAscii(t *testing.T) {
	vm := bufferVM(t)
	mustRunB(t, vm, `
		if (!buf.isAscii(Buffer.from('hello'))) throw new Error('should be ascii');
		if (buf.isAscii(Buffer.from([0x80]))) throw new Error('0x80 is not ascii');
	`)
}

func TestBufferIsUtf8(t *testing.T) {
	vm := bufferVM(t)
	mustRunB(t, vm, `
		if (!buf.isUtf8(Buffer.from('hello'))) throw new Error('should be utf8');
		if (!buf.isUtf8(Buffer.from([0xC3, 0xA9]))) throw new Error('valid utf8 2-byte');
		if (buf.isUtf8(Buffer.from([0xFF, 0xFE]))) throw new Error('invalid utf8');
	`)
}

func TestBufferConstants(t *testing.T) {
	vm := bufferVM(t)
	mustRunB(t, vm, `
		if (buf.constants.MAX_LENGTH !== 0x7FFFFFFF) throw new Error('wrong MAX_LENGTH');
		if (buf.kMaxLength !== 0x7FFFFFFF) throw new Error('wrong kMaxLength');
	`)
}

// ---------------------------------------------------------------------------
// Edge cases
// ---------------------------------------------------------------------------

func TestBufferAllocZero(t *testing.T) {
	vm := bufferVM(t)
	mustRunB(t, vm, `
		var b = Buffer.alloc(0);
		if (b.length !== 0) throw new Error('wrong length');
		if (b.toString() !== '') throw new Error('wrong string');
	`)
}

func TestBufferFrom_TruncateBytes(t *testing.T) {
	vm := bufferVM(t)
	mustRunB(t, vm, `
		var b = Buffer.from([257, -1, 300]);
		if (b[0] !== 1) throw new Error('257 should be 1');
		if (b[1] !== 255) throw new Error('-1 should be 255: ' + b[1]);
		if (b[2] !== 44) throw new Error('300 should be 44: ' + b[2]);
	`)
}

func TestBufferConcat_Empty(t *testing.T) {
	vm := bufferVM(t)
	mustRunB(t, vm, `
		var b = Buffer.concat([]);
		if (b.length !== 0) throw new Error('empty concat should be 0');
	`)
}

func TestBufferUnknownEncoding(t *testing.T) {
	vm := bufferVM(t)
	errStr := mustFailB(t, vm, `Buffer.from('hello', 'nope')`)
	if !strings.Contains(errStr, "Unknown encoding") {
		t.Fatalf("expected unknown encoding error, got: %s", errStr)
	}
}
