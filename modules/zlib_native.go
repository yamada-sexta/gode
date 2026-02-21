package modules

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"compress/zlib"
	"hash/crc32"
	"io"

	"github.com/andybalholm/brotli"
	"github.com/dop251/goja"
)

// setupZlibNative installs a __zlib helper on the VM with Go-backed
// compression functions.
func setupZlibNative(vm *goja.Runtime) {
	obj := vm.NewObject()

	obj.Set("deflateSync", func(call goja.FunctionCall) goja.Value {
		data := latin1ToBytes(call.Argument(0).String())
		level := intArgZ(call, 1, flate.DefaultCompression)
		var buf bytes.Buffer
		w, err := zlib.NewWriterLevel(&buf, level)
		if err != nil {
			panic(vm.ToValue("zlib deflate: " + err.Error()))
		}
		w.Write(data)
		w.Close()
		return vm.ToValue(bytesToLatin1Str(buf.Bytes()))
	})

	obj.Set("inflateSync", func(call goja.FunctionCall) goja.Value {
		data := latin1ToBytes(call.Argument(0).String())
		r, err := zlib.NewReader(bytes.NewReader(data))
		if err != nil {
			panic(vm.ToValue("zlib inflate: " + err.Error()))
		}
		defer r.Close()
		out, err := io.ReadAll(r)
		if err != nil {
			panic(vm.ToValue("zlib inflate: " + err.Error()))
		}
		return vm.ToValue(bytesToLatin1Str(out))
	})

	obj.Set("deflateRawSync", func(call goja.FunctionCall) goja.Value {
		data := latin1ToBytes(call.Argument(0).String())
		level := intArgZ(call, 1, flate.DefaultCompression)
		var buf bytes.Buffer
		w, err := flate.NewWriter(&buf, level)
		if err != nil {
			panic(vm.ToValue("flate deflate: " + err.Error()))
		}
		w.Write(data)
		w.Close()
		return vm.ToValue(bytesToLatin1Str(buf.Bytes()))
	})

	obj.Set("inflateRawSync", func(call goja.FunctionCall) goja.Value {
		data := latin1ToBytes(call.Argument(0).String())
		r := flate.NewReader(bytes.NewReader(data))
		defer r.Close()
		out, err := io.ReadAll(r)
		if err != nil {
			panic(vm.ToValue("flate inflate: " + err.Error()))
		}
		return vm.ToValue(bytesToLatin1Str(out))
	})

	obj.Set("gzipSync", func(call goja.FunctionCall) goja.Value {
		data := latin1ToBytes(call.Argument(0).String())
		level := intArgZ(call, 1, gzip.DefaultCompression)
		var buf bytes.Buffer
		w, err := gzip.NewWriterLevel(&buf, level)
		if err != nil {
			panic(vm.ToValue("gzip: " + err.Error()))
		}
		w.Write(data)
		w.Close()
		return vm.ToValue(bytesToLatin1Str(buf.Bytes()))
	})

	obj.Set("gunzipSync", func(call goja.FunctionCall) goja.Value {
		data := latin1ToBytes(call.Argument(0).String())
		r, err := gzip.NewReader(bytes.NewReader(data))
		if err != nil {
			panic(vm.ToValue("gunzip: " + err.Error()))
		}
		defer r.Close()
		out, err := io.ReadAll(r)
		if err != nil {
			panic(vm.ToValue("gunzip: " + err.Error()))
		}
		return vm.ToValue(bytesToLatin1Str(out))
	})

	obj.Set("unzipSync", func(call goja.FunctionCall) goja.Value {
		data := latin1ToBytes(call.Argument(0).String())
		if len(data) >= 2 && data[0] == 0x1f && data[1] == 0x8b {
			r, err := gzip.NewReader(bytes.NewReader(data))
			if err == nil {
				defer r.Close()
				out, err := io.ReadAll(r)
				if err == nil {
					return vm.ToValue(bytesToLatin1Str(out))
				}
			}
		}
		r, err := zlib.NewReader(bytes.NewReader(data))
		if err != nil {
			panic(vm.ToValue("unzip: " + err.Error()))
		}
		defer r.Close()
		out, err := io.ReadAll(r)
		if err != nil {
			panic(vm.ToValue("unzip: " + err.Error()))
		}
		return vm.ToValue(bytesToLatin1Str(out))
	})

	obj.Set("brotliCompressSync", func(call goja.FunctionCall) goja.Value {
		data := latin1ToBytes(call.Argument(0).String())
		quality := intArgZ(call, 1, brotli.DefaultCompression)
		var buf bytes.Buffer
		w := brotli.NewWriterLevel(&buf, quality)
		w.Write(data)
		w.Close()
		return vm.ToValue(bytesToLatin1Str(buf.Bytes()))
	})

	obj.Set("brotliDecompressSync", func(call goja.FunctionCall) goja.Value {
		data := latin1ToBytes(call.Argument(0).String())
		r := brotli.NewReader(bytes.NewReader(data))
		out, err := io.ReadAll(r)
		if err != nil {
			panic(vm.ToValue("brotli decompress: " + err.Error()))
		}
		return vm.ToValue(bytesToLatin1Str(out))
	})

	obj.Set("crc32", func(call goja.FunctionCall) goja.Value {
		data := latin1ToBytes(call.Argument(0).String())
		init := uint32(0)
		if !goja.IsUndefined(call.Argument(1)) {
			init = uint32(call.Argument(1).ToInteger())
		}
		result := crc32.Update(init, crc32.IEEETable, data)
		return vm.ToValue(result)
	})

	vm.Set("__zlib", obj)
}

// latin1ToBytes decodes a latin1-encoded JS string into raw bytes.
func latin1ToBytes(s string) []byte {
	runes := []rune(s)
	data := make([]byte, len(runes))
	for i, r := range runes {
		data[i] = byte(r)
	}
	return data
}

// bytesToLatin1Str encodes raw bytes as a latin1 string.
func bytesToLatin1Str(data []byte) string {
	runes := make([]rune, len(data))
	for i, b := range data {
		runes[i] = rune(b)
	}
	return string(runes)
}

func intArgZ(call goja.FunctionCall, idx int, def int) int {
	arg := call.Argument(idx)
	if goja.IsUndefined(arg) || goja.IsNull(arg) {
		return def
	}
	return int(arg.ToInteger())
}
