package modules

import (
	"testing"

	"github.com/robertkrimen/otto"
)

func zlibVM(t *testing.T) *otto.Otto {
	t.Helper()
	vm := otto.New()
	NewLoader(vm)
	mustRunZ(t, vm, `var zlib = require('zlib')`)
	mustRunZ(t, vm, `var Buffer = require('buffer').Buffer`)
	return vm
}

func mustRunZ(t *testing.T, vm *otto.Otto, js string) otto.Value {
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

func TestZlibRequire(t *testing.T) {
	vm := otto.New()
	NewLoader(vm)
	mustRunZ(t, vm, `var z = require('zlib'); if (!z.deflateSync) throw new Error('missing deflateSync')`)
}

func TestZlibRequireNodePrefix(t *testing.T) {
	vm := otto.New()
	NewLoader(vm)
	mustRunZ(t, vm, `var z = require('node:zlib'); if (!z.gzipSync) throw new Error('missing gzipSync')`)
}

// ---------------------------------------------------------------------------
// deflateSync / inflateSync (zlib format)
// ---------------------------------------------------------------------------

func TestZlib_DeflateInflateSync(t *testing.T) {
	vm := zlibVM(t)
	mustRunZ(t, vm, `
		var input = Buffer.from('Hello, World!');
		var compressed = zlib.deflateSync(input);
		if (!Buffer.isBuffer(compressed)) throw new Error('should return Buffer');
		if (compressed.length === 0) throw new Error('compressed should not be empty');
		var decompressed = zlib.inflateSync(compressed);
		if (decompressed.toString() !== 'Hello, World!') throw new Error('roundtrip failed: ' + decompressed.toString());
	`)
}

func TestZlib_DeflateInflateSync_String(t *testing.T) {
	vm := zlibVM(t)
	mustRunZ(t, vm, `
		var compressed = zlib.deflateSync('hello zlib');
		var decompressed = zlib.inflateSync(compressed);
		if (decompressed.toString() !== 'hello zlib') throw new Error('string roundtrip failed');
	`)
}

func TestZlib_DeflateSync_Level(t *testing.T) {
	vm := zlibVM(t)
	mustRunZ(t, vm, `
		var data = 'aaaaaaaaaaaaaaaaaaaaaaaaaaaa';
		var best = zlib.deflateSync(data, { level: 9 });
		var fast = zlib.deflateSync(data, { level: 1 });
		// Both should roundtrip correctly
		if (zlib.inflateSync(best).toString() !== data) throw new Error('best level failed');
		if (zlib.inflateSync(fast).toString() !== data) throw new Error('fast level failed');
	`)
}

// ---------------------------------------------------------------------------
// deflateRawSync / inflateRawSync
// ---------------------------------------------------------------------------

func TestZlib_DeflateRawInflateRawSync(t *testing.T) {
	vm := zlibVM(t)
	mustRunZ(t, vm, `
		var input = Buffer.from('Raw deflate test');
		var compressed = zlib.deflateRawSync(input);
		var decompressed = zlib.inflateRawSync(compressed);
		if (decompressed.toString() !== 'Raw deflate test') throw new Error('raw roundtrip failed');
	`)
}

// ---------------------------------------------------------------------------
// gzipSync / gunzipSync
// ---------------------------------------------------------------------------

func TestZlib_GzipGunzipSync(t *testing.T) {
	vm := zlibVM(t)
	mustRunZ(t, vm, `
		var input = Buffer.from('Gzip test data');
		var compressed = zlib.gzipSync(input);
		if (compressed[0] !== 0x1f || compressed[1] !== 0x8b) throw new Error('missing gzip magic');
		var decompressed = zlib.gunzipSync(compressed);
		if (decompressed.toString() !== 'Gzip test data') throw new Error('gzip roundtrip failed');
	`)
}

func TestZlib_GzipSync_String(t *testing.T) {
	vm := zlibVM(t)
	mustRunZ(t, vm, `
		var compressed = zlib.gzipSync('hello gzip');
		var decompressed = zlib.gunzipSync(compressed);
		if (decompressed.toString() !== 'hello gzip') throw new Error('gzip string roundtrip failed');
	`)
}

// ---------------------------------------------------------------------------
// unzipSync
// ---------------------------------------------------------------------------

func TestZlib_UnzipSync_Gzip(t *testing.T) {
	vm := zlibVM(t)
	mustRunZ(t, vm, `
		var compressed = zlib.gzipSync('unzip gzip test');
		var decompressed = zlib.unzipSync(compressed);
		if (decompressed.toString() !== 'unzip gzip test') throw new Error('unzip gzip failed');
	`)
}

func TestZlib_UnzipSync_Zlib(t *testing.T) {
	vm := zlibVM(t)
	mustRunZ(t, vm, `
		var compressed = zlib.deflateSync('unzip zlib test');
		var decompressed = zlib.unzipSync(compressed);
		if (decompressed.toString() !== 'unzip zlib test') throw new Error('unzip zlib failed');
	`)
}

// ---------------------------------------------------------------------------
// brotliCompressSync / brotliDecompressSync
// ---------------------------------------------------------------------------

func TestZlib_BrotliSync(t *testing.T) {
	vm := zlibVM(t)
	mustRunZ(t, vm, `
		var input = Buffer.from('Brotli compression test');
		var compressed = zlib.brotliCompressSync(input);
		if (compressed.length === 0) throw new Error('brotli compressed should not be empty');
		var decompressed = zlib.brotliDecompressSync(compressed);
		if (decompressed.toString() !== 'Brotli compression test') throw new Error('brotli roundtrip failed');
	`)
}

func TestZlib_BrotliSync_String(t *testing.T) {
	vm := zlibVM(t)
	mustRunZ(t, vm, `
		var compressed = zlib.brotliCompressSync('hello brotli');
		var decompressed = zlib.brotliDecompressSync(compressed);
		if (decompressed.toString() !== 'hello brotli') throw new Error('brotli string roundtrip failed');
	`)
}

// ---------------------------------------------------------------------------
// crc32
// ---------------------------------------------------------------------------

func TestZlib_CRC32(t *testing.T) {
	vm := zlibVM(t)
	mustRunZ(t, vm, `
		var checksum = zlib.crc32('hello');
		if (typeof checksum !== 'number') throw new Error('crc32 should return number');
		if (checksum !== 907060870) throw new Error('wrong crc32: ' + checksum);
	`)
}

func TestZlib_CRC32_Buffer(t *testing.T) {
	vm := zlibVM(t)
	mustRunZ(t, vm, `
		var checksum = zlib.crc32(Buffer.from('hello'));
		if (checksum !== 907060870) throw new Error('wrong buffer crc32: ' + checksum);
	`)
}

func TestZlib_CRC32_InitialValue(t *testing.T) {
	vm := zlibVM(t)
	mustRunZ(t, vm, `
		var c1 = zlib.crc32('hello');
		var c2 = zlib.crc32(' world', c1);
		var full = zlib.crc32('hello world');
		if (c2 !== full) throw new Error('incremental crc32 failed: ' + c2 + ' !== ' + full);
	`)
}

// ---------------------------------------------------------------------------
// Callback methods
// ---------------------------------------------------------------------------

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

func TestZlib_GzipCallback(t *testing.T) {
	vm := zlibVM(t)
	mustRunZ(t, vm, `
		var result;
		zlib.gzip('gzip callback', function(err, buf) {
			if (err) throw err;
			result = zlib.gunzipSync(buf).toString();
		});
		if (result !== 'gzip callback') throw new Error('callback gzip failed');
	`)
}

// ---------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------

func TestZlib_Constants(t *testing.T) {
	vm := zlibVM(t)
	mustRunZ(t, vm, `
		if (typeof zlib.constants !== 'object') throw new Error('missing constants');
		if (zlib.constants.Z_DEFAULT_COMPRESSION !== -1) throw new Error('wrong Z_DEFAULT_COMPRESSION');
		if (zlib.constants.Z_BEST_COMPRESSION !== 9) throw new Error('wrong Z_BEST_COMPRESSION');
		if (zlib.constants.Z_NO_COMPRESSION !== 0) throw new Error('wrong Z_NO_COMPRESSION');
	`)
}

// ---------------------------------------------------------------------------
// All sync methods exist
// ---------------------------------------------------------------------------

func TestZlib_AllMethodsExist(t *testing.T) {
	vm := zlibVM(t)
	methods := []string{
		"deflateSync", "inflateSync", "deflateRawSync", "inflateRawSync",
		"gzipSync", "gunzipSync", "unzipSync",
		"brotliCompressSync", "brotliDecompressSync",
		"deflate", "inflate", "deflateRaw", "inflateRaw",
		"gzip", "gunzip", "unzip",
		"brotliCompress", "brotliDecompress",
		"crc32",
	}
	for _, m := range methods {
		t.Run(m, func(t *testing.T) {
			mustRunZ(t, vm, `if (typeof zlib.`+m+` !== 'function') throw new Error('missing: `+m+`')`)
		})
	}
}

// ---------------------------------------------------------------------------
// Binary data roundtrip (all byte values 0-255)
// ---------------------------------------------------------------------------

func TestZlib_BinaryRoundtrip(t *testing.T) {
	vm := zlibVM(t)
	mustRunZ(t, vm, `
		var data = [];
		for (var i = 0; i < 256; i++) data.push(i);
		var buf = Buffer.from(data);
		var compressed = zlib.gzipSync(buf);
		var decompressed = zlib.gunzipSync(compressed);
		if (decompressed.length !== 256) throw new Error('wrong length: ' + decompressed.length);
		for (var j = 0; j < 256; j++) {
			if (decompressed[j] !== j) throw new Error('byte ' + j + ' mismatch: ' + decompressed[j]);
		}
	`)
}
