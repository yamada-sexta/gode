package modules

import (
	"testing"

	"github.com/dop251/goja"
)

func zlibVM(t *testing.T) *goja.Runtime {
	t.Helper()
	vm := goja.New()
	NewLoader(vm)
	mustRunZ(t, vm, `var zlib = require('zlib')`)
	mustRunZ(t, vm, `var Buffer = require('buffer').Buffer`)
	return vm
}

func mustRunZ(t *testing.T, vm *goja.Runtime, js string) goja.Value {
	t.Helper()
	val, err := vm.RunString(js)
	if err != nil {
		t.Fatalf("unexpected error: %v\nscript: %s", err, js)
	}
	return val
}

func TestZlibRequire(t *testing.T) {
	vm := goja.New()
	NewLoader(vm)
	mustRunZ(t, vm, `var z = require('zlib'); if (!z.deflateSync) throw new Error('missing deflateSync')`)
}

func TestZlib_DeflateInflateSync(t *testing.T) {
	vm := zlibVM(t)
	mustRunZ(t, vm, `
		var input = Buffer.from('Hello, World!');
		var compressed = zlib.deflateSync(input);
		if (!Buffer.isBuffer(compressed)) throw new Error('should return Buffer');
		var decompressed = zlib.inflateSync(compressed);
		if (decompressed.toString() !== 'Hello, World!') throw new Error('roundtrip failed');
	`)
}

func TestZlib_GzipGunzipSync(t *testing.T) {
	vm := zlibVM(t)
	mustRunZ(t, vm, `
		var compressed = zlib.gzipSync('hello gzip');
		var decompressed = zlib.gunzipSync(compressed);
		if (decompressed.toString() !== 'hello gzip') throw new Error('gzip roundtrip failed');
	`)
}

func TestZlib_DeflateRawInflateRawSync(t *testing.T) {
	vm := zlibVM(t)
	mustRunZ(t, vm, `
		var compressed = zlib.deflateRawSync('Raw test');
		var decompressed = zlib.inflateRawSync(compressed);
		if (decompressed.toString() !== 'Raw test') throw new Error('raw roundtrip failed');
	`)
}

func TestZlib_BrotliSync(t *testing.T) {
	vm := zlibVM(t)
	mustRunZ(t, vm, `
		var compressed = zlib.brotliCompressSync('Brotli test');
		var decompressed = zlib.brotliDecompressSync(compressed);
		if (decompressed.toString() !== 'Brotli test') throw new Error('brotli roundtrip failed');
	`)
}

func TestZlib_CRC32(t *testing.T) {
	vm := zlibVM(t)
	mustRunZ(t, vm, `
		var checksum = zlib.crc32('hello');
		if (checksum !== 907060870) throw new Error('wrong crc32: ' + checksum);
	`)
}

func TestZlib_CRC32_Incremental(t *testing.T) {
	vm := zlibVM(t)
	mustRunZ(t, vm, `
		var c1 = zlib.crc32('hello');
		var c2 = zlib.crc32(' world', c1);
		var full = zlib.crc32('hello world');
		if (c2 !== full) throw new Error('incremental crc32 failed');
	`)
}

func TestZlib_DeflateCallback(t *testing.T) {
	vm := zlibVM(t)
	mustRunZ(t, vm, `
		var result;
		zlib.deflate('callback test', function(err, buf) {
			if (err) throw err;
			result = zlib.inflateSync(buf).toString();
		});
		if (result !== 'callback test') throw new Error('callback deflate failed');
	`)
}

func TestZlib_Constants(t *testing.T) {
	vm := zlibVM(t)
	mustRunZ(t, vm, `
		if (typeof zlib.constants !== 'object') throw new Error('missing constants');
		if (zlib.constants.Z_DEFAULT_COMPRESSION !== -1) throw new Error('wrong constant');
	`)
}

func TestZlib_BinaryRoundtrip(t *testing.T) {
	vm := zlibVM(t)
	mustRunZ(t, vm, `
		var data = [];
		for (var i = 0; i < 256; i++) data.push(i);
		var buf = Buffer.from(data);
		var compressed = zlib.gzipSync(buf);
		var decompressed = zlib.gunzipSync(compressed);
		if (decompressed.length !== 256) throw new Error('wrong length');
		for (var j = 0; j < 256; j++) {
			if (decompressed[j] !== j) throw new Error('byte ' + j + ' mismatch');
		}
	`)
}

func TestZlib_AllMethodsExist(t *testing.T) {
	vm := zlibVM(t)
	methods := []string{
		"deflateSync", "inflateSync", "deflateRawSync", "inflateRawSync",
		"gzipSync", "gunzipSync", "unzipSync",
		"brotliCompressSync", "brotliDecompressSync",
		"deflate", "inflate", "gzip", "gunzip",
		"brotliCompress", "brotliDecompress",
		"crc32",
	}
	for _, m := range methods {
		t.Run(m, func(t *testing.T) {
			mustRunZ(t, vm, `if (typeof zlib.`+m+` !== 'function') throw new Error('missing: `+m+`')`)
		})
	}
}
