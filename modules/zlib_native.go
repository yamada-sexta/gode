package modules

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"compress/zlib"
	"hash/crc32"
	"io"

	"github.com/andybalholm/brotli"
	"github.com/robertkrimen/otto"
)

// setupZlibNative installs a __zlib helper on the VM with Go-backed
// compression functions. Data is passed as latin1-encoded strings for
// efficient transfer between JS and Go.
func setupZlibNative(vm *otto.Otto) {
	obj, _ := vm.Object(`({})`)

	// --- deflate (zlib header, RFC 1950) ---

	obj.Set("deflateSync", func(call otto.FunctionCall) otto.Value {
		data := latin1ToBytes(call.Argument(0))
		level := intArg(call, 1, flate.DefaultCompression)
		var buf bytes.Buffer
		w, err := zlib.NewWriterLevel(&buf, level)
		if err != nil {
			panic(vm.MakeCustomError("Error", "zlib deflate: "+err.Error()))
		}
		w.Write(data)
		w.Close()
		v, _ := otto.ToValue(bytesToLatin1(buf.Bytes()))
		return v
	})

	obj.Set("inflateSync", func(call otto.FunctionCall) otto.Value {
		data := latin1ToBytes(call.Argument(0))
		r, err := zlib.NewReader(bytes.NewReader(data))
		if err != nil {
			panic(vm.MakeCustomError("Error", "zlib inflate: "+err.Error()))
		}
		defer r.Close()
		out, err := io.ReadAll(r)
		if err != nil {
			panic(vm.MakeCustomError("Error", "zlib inflate: "+err.Error()))
		}
		v, _ := otto.ToValue(bytesToLatin1(out))
		return v
	})

	// --- deflateRaw (no header, RFC 1951) ---

	obj.Set("deflateRawSync", func(call otto.FunctionCall) otto.Value {
		data := latin1ToBytes(call.Argument(0))
		level := intArg(call, 1, flate.DefaultCompression)
		var buf bytes.Buffer
		w, err := flate.NewWriter(&buf, level)
		if err != nil {
			panic(vm.MakeCustomError("Error", "flate deflate: "+err.Error()))
		}
		w.Write(data)
		w.Close()
		v, _ := otto.ToValue(bytesToLatin1(buf.Bytes()))
		return v
	})

	obj.Set("inflateRawSync", func(call otto.FunctionCall) otto.Value {
		data := latin1ToBytes(call.Argument(0))
		r := flate.NewReader(bytes.NewReader(data))
		defer r.Close()
		out, err := io.ReadAll(r)
		if err != nil {
			panic(vm.MakeCustomError("Error", "flate inflate: "+err.Error()))
		}
		v, _ := otto.ToValue(bytesToLatin1(out))
		return v
	})

	// --- gzip (RFC 1952) ---

	obj.Set("gzipSync", func(call otto.FunctionCall) otto.Value {
		data := latin1ToBytes(call.Argument(0))
		level := intArg(call, 1, gzip.DefaultCompression)
		var buf bytes.Buffer
		w, err := gzip.NewWriterLevel(&buf, level)
		if err != nil {
			panic(vm.MakeCustomError("Error", "gzip: "+err.Error()))
		}
		w.Write(data)
		w.Close()
		v, _ := otto.ToValue(bytesToLatin1(buf.Bytes()))
		return v
	})

	obj.Set("gunzipSync", func(call otto.FunctionCall) otto.Value {
		data := latin1ToBytes(call.Argument(0))
		r, err := gzip.NewReader(bytes.NewReader(data))
		if err != nil {
			panic(vm.MakeCustomError("Error", "gunzip: "+err.Error()))
		}
		defer r.Close()
		out, err := io.ReadAll(r)
		if err != nil {
			panic(vm.MakeCustomError("Error", "gunzip: "+err.Error()))
		}
		v, _ := otto.ToValue(bytesToLatin1(out))
		return v
	})

	// --- unzip (auto-detect gzip or zlib) ---

	obj.Set("unzipSync", func(call otto.FunctionCall) otto.Value {
		data := latin1ToBytes(call.Argument(0))
		// Try gzip first (magic: 0x1f 0x8b)
		if len(data) >= 2 && data[0] == 0x1f && data[1] == 0x8b {
			r, err := gzip.NewReader(bytes.NewReader(data))
			if err == nil {
				defer r.Close()
				out, err := io.ReadAll(r)
				if err == nil {
					v, _ := otto.ToValue(bytesToLatin1(out))
					return v
				}
			}
		}
		// Try zlib
		r, err := zlib.NewReader(bytes.NewReader(data))
		if err != nil {
			panic(vm.MakeCustomError("Error", "unzip: "+err.Error()))
		}
		defer r.Close()
		out, err := io.ReadAll(r)
		if err != nil {
			panic(vm.MakeCustomError("Error", "unzip: "+err.Error()))
		}
		v, _ := otto.ToValue(bytesToLatin1(out))
		return v
	})

	// --- brotli ---

	obj.Set("brotliCompressSync", func(call otto.FunctionCall) otto.Value {
		data := latin1ToBytes(call.Argument(0))
		quality := intArg(call, 1, brotli.DefaultCompression)
		var buf bytes.Buffer
		w := brotli.NewWriterLevel(&buf, quality)
		w.Write(data)
		w.Close()
		v, _ := otto.ToValue(bytesToLatin1(buf.Bytes()))
		return v
	})

	obj.Set("brotliDecompressSync", func(call otto.FunctionCall) otto.Value {
		data := latin1ToBytes(call.Argument(0))
		r := brotli.NewReader(bytes.NewReader(data))
		out, err := io.ReadAll(r)
		if err != nil {
			panic(vm.MakeCustomError("Error", "brotli decompress: "+err.Error()))
		}
		v, _ := otto.ToValue(bytesToLatin1(out))
		return v
	})

	// --- crc32 ---

	obj.Set("crc32", func(call otto.FunctionCall) otto.Value {
		data := latin1ToBytes(call.Argument(0))
		init := uint32(0)
		if !call.Argument(1).IsUndefined() {
			n, _ := call.Argument(1).ToInteger()
			init = uint32(n)
		}
		result := crc32.Update(init, crc32.IEEETable, data)
		v, _ := otto.ToValue(result)
		return v
	})

	vm.Set("__zlib", obj)
}

// latin1ToBytes decodes a latin1-encoded JS string into raw bytes.
// Each JS character (Unicode 0x00–0xFF) maps to one byte.
func latin1ToBytes(val otto.Value) []byte {
	s, _ := val.ToString()
	runes := []rune(s)
	data := make([]byte, len(runes))
	for i, r := range runes {
		data[i] = byte(r)
	}
	return data
}

// bytesToLatin1 encodes raw bytes as a latin1 string that JS can decode.
func bytesToLatin1(data []byte) string {
	runes := make([]rune, len(data))
	for i, b := range data {
		runes[i] = rune(b)
	}
	return string(runes)
}

func intArg(call otto.FunctionCall, idx int, def int) int {
	arg := call.Argument(idx)
	if arg.IsUndefined() || arg.IsNull() {
		return def
	}
	n, err := arg.ToInteger()
	if err != nil {
		return def
	}
	return int(n)
}
