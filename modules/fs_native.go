package modules

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"syscall"

	"github.com/dop251/goja"
)

// setupFSNative installs a __fs helper object on the VM with Go-backed
// synchronous filesystem functions that the fs.js wrapper calls.
func setupFSNative(vm *goja.Runtime) {
	fsObj := vm.NewObject()

	// ── Constants ──────────────────────────────────────────────────────

	consts := vm.NewObject()
	consts.Set("F_OK", 0)
	consts.Set("R_OK", 4)
	consts.Set("W_OK", 2)
	consts.Set("X_OK", 1)
	consts.Set("COPYFILE_EXCL", 1)
	consts.Set("COPYFILE_FICLONE", 2)
	consts.Set("COPYFILE_FICLONE_FORCE", 4)
	// Open flags
	consts.Set("O_RDONLY", syscall.O_RDONLY)
	consts.Set("O_WRONLY", syscall.O_WRONLY)
	consts.Set("O_RDWR", syscall.O_RDWR)
	consts.Set("O_CREAT", syscall.O_CREAT)
	consts.Set("O_EXCL", syscall.O_EXCL)
	consts.Set("O_TRUNC", syscall.O_TRUNC)
	consts.Set("O_APPEND", syscall.O_APPEND)
	consts.Set("O_SYNC", syscall.O_SYNC)
	fsObj.Set("constants", consts)

	// ── Helper: build a Stats-like JS object ──────────────────────────

	buildStats := func(info os.FileInfo) goja.Value {
		s := vm.NewObject()
		mode := info.Mode()

		s.Set("dev", 0)
		s.Set("ino", 0)
		s.Set("mode", int(mode.Perm()))
		s.Set("nlink", 1)
		s.Set("uid", 0)
		s.Set("gid", 0)
		s.Set("rdev", 0)
		s.Set("size", info.Size())
		s.Set("blksize", 4096)
		s.Set("blocks", (info.Size()+511)/512)

		mt := info.ModTime()
		ms := float64(mt.UnixNano()) / 1e6
		s.Set("atimeMs", ms)
		s.Set("mtimeMs", ms)
		s.Set("ctimeMs", ms)
		s.Set("birthtimeMs", ms)
		s.Set("atime", mt.Format("2006-01-02T15:04:05.000Z"))
		s.Set("mtime", mt.Format("2006-01-02T15:04:05.000Z"))
		s.Set("ctime", mt.Format("2006-01-02T15:04:05.000Z"))
		s.Set("birthtime", mt.Format("2006-01-02T15:04:05.000Z"))

		// Try to get uid/gid/ino/dev from syscall
		if sys, ok := info.Sys().(*syscall.Stat_t); ok {
			s.Set("dev", sys.Dev)
			s.Set("ino", sys.Ino)
			s.Set("nlink", sys.Nlink)
			s.Set("uid", sys.Uid)
			s.Set("gid", sys.Gid)
			s.Set("rdev", sys.Rdev)
			s.Set("blksize", sys.Blksize)
			s.Set("blocks", sys.Blocks)
		}

		isDir := mode.IsDir()
		isFile := mode.IsRegular()
		isSymlink := mode&os.ModeSymlink != 0
		isSocket := mode&os.ModeSocket != 0
		isFIFO := mode&os.ModeNamedPipe != 0
		isBlockDev := mode&os.ModeDevice != 0 && mode&os.ModeCharDevice == 0
		isCharDev := mode&os.ModeCharDevice != 0

		s.Set("isFile", func(call goja.FunctionCall) goja.Value {
			return vm.ToValue(isFile)
		})
		s.Set("isDirectory", func(call goja.FunctionCall) goja.Value {
			return vm.ToValue(isDir)
		})
		s.Set("isSymbolicLink", func(call goja.FunctionCall) goja.Value {
			return vm.ToValue(isSymlink)
		})
		s.Set("isBlockDevice", func(call goja.FunctionCall) goja.Value {
			return vm.ToValue(isBlockDev)
		})
		s.Set("isCharacterDevice", func(call goja.FunctionCall) goja.Value {
			return vm.ToValue(isCharDev)
		})
		s.Set("isFIFO", func(call goja.FunctionCall) goja.Value {
			return vm.ToValue(isFIFO)
		})
		s.Set("isSocket", func(call goja.FunctionCall) goja.Value {
			return vm.ToValue(isSocket)
		})

		return s
	}

	// ── Helper: throw an Error with code ──────────────────────────────

	throwFSError := func(err error, syscallName, path string) {
		code := "ERR"
		errno := -1
		msg := err.Error()
		if pe, ok := err.(*os.PathError); ok {
			msg = pe.Err.Error()
			if se, ok := pe.Err.(syscall.Errno); ok {
				errno = int(se)
				switch se {
				case syscall.ENOENT:
					code = "ENOENT"
				case syscall.EEXIST:
					code = "EEXIST"
				case syscall.EACCES:
					code = "EACCES"
				case syscall.EPERM:
					code = "EPERM"
				case syscall.ENOTDIR:
					code = "ENOTDIR"
				case syscall.EISDIR:
					code = "EISDIR"
				case syscall.ENOTEMPTY:
					code = "ENOTEMPTY"
				default:
					code = se.Error()
				}
			}
		} else if le, ok := err.(*os.LinkError); ok {
			msg = le.Err.Error()
			if se, ok := le.Err.(syscall.Errno); ok {
				errno = int(se)
				switch se {
				case syscall.ENOENT:
					code = "ENOENT"
				case syscall.EEXIST:
					code = "EEXIST"
				default:
					code = se.Error()
				}
			}
		}

		errStr := fmt.Sprintf("%s: %s, %s '%s'", code, msg, syscallName, path)
		errObj, _ := vm.RunString(fmt.Sprintf(`
			(function() {
				var e = new Error(%q);
				e.code = %q;
				e.errno = %d;
				e.syscall = %q;
				e.path = %q;
				return e;
			})()
		`, errStr, code, errno, syscallName, path))
		panic(errObj)
	}

	// ── readFile ──────────────────────────────────────────────────────
	fsObj.Set("readFile", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()

		var callback goja.Callable
		var options goja.Value
		var hasOptions bool

		if len(call.Arguments) > 1 {
			arg1 := call.Argument(1)
			if cb, ok := goja.AssertFunction(arg1); ok {
				callback = cb
			} else {
				options = arg1
				hasOptions = true
				if len(call.Arguments) > 2 {
					if cb, ok := goja.AssertFunction(call.Argument(2)); ok {
						callback = cb
					}
				}
			}
		}

		if callback == nil {
			panic(vm.ToValue("TypeError [ERR_INVALID_CALLBACK]: Callback must be a function"))
		}

		data, err := os.ReadFile(path)
		if err != nil {
			code := "ERR"
			if pe, ok := err.(*os.PathError); ok {
				if se, ok := pe.Err.(syscall.Errno); ok {
					switch se {
					case syscall.ENOENT:
						code = "ENOENT"
					default:
						code = se.Error()
					}
				}
			}

			errObj, _ := vm.RunString(fmt.Sprintf(`
				(function() {
					var e = new Error(%q);
					e.code = %q;
					e.path = %q;
					return e;
				})()
			`, err.Error(), code, path))
			callback(goja.Undefined(), errObj)
			return goja.Undefined()
		}

		var result goja.Value
		enc := ""
		if hasOptions {
			if s, ok := options.Export().(string); ok {
				enc = s
			} else if obj, ok := options.Export().(map[string]interface{}); ok {
				if e, ok := obj["encoding"]; ok {
					enc, _ = e.(string)
				}
			}
		}

		if enc == "utf8" || enc == "utf-8" || enc == "ascii" || enc == "latin1" {
			result = vm.ToValue(string(data))
		} else {
			// Return string by default (Buffer not yet fully supported)
			result = vm.ToValue(string(data))
		}

		callback(goja.Undefined(), goja.Null(), result)
		return goja.Undefined()
	})

	// ── readFileSync ──────────────────────────────────────────────────

	fsObj.Set("readFileSync", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		data, err := os.ReadFile(path)
		if err != nil {
			throwFSError(err, "open", path)
		}
		// Check for encoding option
		if len(call.Arguments) > 1 {
			arg := call.Argument(1)
			enc := ""
			if s, ok := arg.Export().(string); ok {
				enc = s
			} else if obj, ok := arg.Export().(map[string]interface{}); ok {
				if e, ok := obj["encoding"]; ok {
					enc, _ = e.(string)
				}
			}
			if enc == "utf8" || enc == "utf-8" || enc == "ascii" || enc == "latin1" {
				return vm.ToValue(string(data))
			}
		}
		// Return string by default (Buffer not yet fully supported)
		return vm.ToValue(string(data))
	})

	// ── writeFileSync ─────────────────────────────────────────────────

	fsObj.Set("writeFileSync", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		data := call.Argument(1).String()
		mode := os.FileMode(0o666)
		flag := os.O_WRONLY | os.O_CREATE | os.O_TRUNC

		if len(call.Arguments) > 2 {
			if obj, ok := call.Argument(2).Export().(map[string]interface{}); ok {
				if m, ok := obj["mode"]; ok {
					if mi, ok := m.(int64); ok {
						mode = os.FileMode(mi)
					}
				}
				if f, ok := obj["flag"]; ok {
					if fs, ok := f.(string); ok {
						switch fs {
						case "a":
							flag = os.O_WRONLY | os.O_CREATE | os.O_APPEND
						case "ax":
							flag = os.O_WRONLY | os.O_CREATE | os.O_EXCL
						case "w":
							flag = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
						case "wx":
							flag = os.O_WRONLY | os.O_CREATE | os.O_EXCL
						}
					}
				}
			}
		}

		f, err := os.OpenFile(path, flag, mode)
		if err != nil {
			throwFSError(err, "open", path)
		}
		defer f.Close()
		_, err = f.WriteString(data)
		if err != nil {
			throwFSError(err, "write", path)
		}
		return goja.Undefined()
	})

	// ── appendFileSync ────────────────────────────────────────────────

	fsObj.Set("appendFileSync", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		data := call.Argument(1).String()
		f, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0o666)
		if err != nil {
			throwFSError(err, "open", path)
		}
		defer f.Close()
		_, err = f.WriteString(data)
		if err != nil {
			throwFSError(err, "write", path)
		}
		return goja.Undefined()
	})

	// ── existsSync ────────────────────────────────────────────────────

	fsObj.Set("existsSync", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		_, err := os.Stat(path)
		return vm.ToValue(err == nil)
	})

	// ── accessSync ────────────────────────────────────────────────────

	fsObj.Set("accessSync", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		mode := 0 // F_OK
		if len(call.Arguments) > 1 {
			mode = int(call.Argument(1).ToInteger())
		}

		info, err := os.Stat(path)
		if err != nil {
			throwFSError(err, "access", path)
		}

		perm := info.Mode().Perm()
		if mode&4 != 0 && perm&0o444 == 0 { // R_OK
			throwFSError(&os.PathError{Op: "access", Path: path, Err: syscall.EACCES}, "access", path)
		}
		if mode&2 != 0 && perm&0o222 == 0 { // W_OK
			throwFSError(&os.PathError{Op: "access", Path: path, Err: syscall.EACCES}, "access", path)
		}
		if mode&1 != 0 && perm&0o111 == 0 { // X_OK
			throwFSError(&os.PathError{Op: "access", Path: path, Err: syscall.EACCES}, "access", path)
		}
		return goja.Undefined()
	})

	// ── statSync ──────────────────────────────────────────────────────

	fsObj.Set("statSync", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		throwIfNoEntry := true
		if len(call.Arguments) > 1 {
			if obj, ok := call.Argument(1).Export().(map[string]interface{}); ok {
				if v, ok := obj["throwIfNoEntry"]; ok {
					if b, ok := v.(bool); ok {
						throwIfNoEntry = b
					}
				}
			}
		}
		info, err := os.Stat(path)
		if err != nil {
			if !throwIfNoEntry && os.IsNotExist(err) {
				return goja.Undefined()
			}
			throwFSError(err, "stat", path)
		}
		return buildStats(info)
	})

	// ── lstatSync ─────────────────────────────────────────────────────

	fsObj.Set("lstatSync", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		throwIfNoEntry := true
		if len(call.Arguments) > 1 {
			if obj, ok := call.Argument(1).Export().(map[string]interface{}); ok {
				if v, ok := obj["throwIfNoEntry"]; ok {
					if b, ok := v.(bool); ok {
						throwIfNoEntry = b
					}
				}
			}
		}
		info, err := os.Lstat(path)
		if err != nil {
			if !throwIfNoEntry && os.IsNotExist(err) {
				return goja.Undefined()
			}
			throwFSError(err, "lstat", path)
		}
		return buildStats(info)
	})

	// ── readdirSync ───────────────────────────────────────────────────

	fsObj.Set("readdirSync", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		withFileTypes := false
		if len(call.Arguments) > 1 {
			if obj, ok := call.Argument(1).Export().(map[string]interface{}); ok {
				if v, ok := obj["withFileTypes"]; ok {
					if b, ok := v.(bool); ok {
						withFileTypes = b
					}
				}
			}
		}

		entries, err := os.ReadDir(path)
		if err != nil {
			throwFSError(err, "scandir", path)
		}

		if withFileTypes {
			result := make([]interface{}, len(entries))
			for i, e := range entries {
				d := vm.NewObject()
				d.Set("name", e.Name())
				isDir := e.IsDir()
				isFile := e.Type().IsRegular()
				isSymlink := e.Type()&fs.ModeSymlink != 0
				d.Set("isFile", func(call goja.FunctionCall) goja.Value {
					return vm.ToValue(isFile)
				})
				d.Set("isDirectory", func(call goja.FunctionCall) goja.Value {
					return vm.ToValue(isDir)
				})
				d.Set("isSymbolicLink", func(call goja.FunctionCall) goja.Value {
					return vm.ToValue(isSymlink)
				})
				d.Set("isBlockDevice", func(call goja.FunctionCall) goja.Value {
					return vm.ToValue(false)
				})
				d.Set("isCharacterDevice", func(call goja.FunctionCall) goja.Value {
					return vm.ToValue(false)
				})
				d.Set("isFIFO", func(call goja.FunctionCall) goja.Value {
					return vm.ToValue(false)
				})
				d.Set("isSocket", func(call goja.FunctionCall) goja.Value {
					return vm.ToValue(false)
				})
				result[i] = d
			}
			return vm.ToValue(result)
		}

		names := make([]interface{}, len(entries))
		for i, e := range entries {
			names[i] = e.Name()
		}
		return vm.ToValue(names)
	})

	// ── mkdirSync ─────────────────────────────────────────────────────

	fsObj.Set("mkdirSync", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		mode := os.FileMode(0o777)
		recursive := false

		if len(call.Arguments) > 1 {
			if obj, ok := call.Argument(1).Export().(map[string]interface{}); ok {
				if m, ok := obj["mode"]; ok {
					if mi, ok := m.(int64); ok {
						mode = os.FileMode(mi)
					}
				}
				if r, ok := obj["recursive"]; ok {
					if rb, ok := r.(bool); ok {
						recursive = rb
					}
				}
			}
		}

		var err error
		if recursive {
			err = os.MkdirAll(path, mode)
		} else {
			err = os.Mkdir(path, mode)
		}
		if err != nil {
			throwFSError(err, "mkdir", path)
		}

		if recursive {
			return vm.ToValue(path)
		}
		return goja.Undefined()
	})

	// ── rmdirSync ─────────────────────────────────────────────────────

	fsObj.Set("rmdirSync", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		err := os.Remove(path)
		if err != nil {
			throwFSError(err, "rmdir", path)
		}
		return goja.Undefined()
	})

	// ── rmSync ────────────────────────────────────────────────────────

	fsObj.Set("rmSync", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		recursive := false
		force := false

		if len(call.Arguments) > 1 {
			if obj, ok := call.Argument(1).Export().(map[string]interface{}); ok {
				if r, ok := obj["recursive"]; ok {
					if rb, ok := r.(bool); ok {
						recursive = rb
					}
				}
				if f, ok := obj["force"]; ok {
					if fb, ok := f.(bool); ok {
						force = fb
					}
				}
			}
		}

		var err error
		if recursive {
			err = os.RemoveAll(path)
		} else {
			err = os.Remove(path)
		}
		if err != nil && !force {
			throwFSError(err, "rm", path)
		}
		return goja.Undefined()
	})

	// ── unlinkSync ────────────────────────────────────────────────────

	fsObj.Set("unlinkSync", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		err := os.Remove(path)
		if err != nil {
			throwFSError(err, "unlink", path)
		}
		return goja.Undefined()
	})

	// ── renameSync ────────────────────────────────────────────────────

	fsObj.Set("renameSync", func(call goja.FunctionCall) goja.Value {
		oldPath := call.Argument(0).String()
		newPath := call.Argument(1).String()
		err := os.Rename(oldPath, newPath)
		if err != nil {
			throwFSError(err, "rename", oldPath)
		}
		return goja.Undefined()
	})

	// ── copyFileSync ──────────────────────────────────────────────────

	fsObj.Set("copyFileSync", func(call goja.FunctionCall) goja.Value {
		src := call.Argument(0).String()
		dest := call.Argument(1).String()
		mode := 0
		if len(call.Arguments) > 2 {
			mode = int(call.Argument(2).ToInteger())
		}

		// COPYFILE_EXCL: fail if dest exists
		if mode&1 != 0 {
			if _, err := os.Stat(dest); err == nil {
				throwFSError(&os.PathError{Op: "copyfile", Path: dest, Err: syscall.EEXIST}, "copyfile", dest)
			}
		}

		data, err := os.ReadFile(src)
		if err != nil {
			throwFSError(err, "open", src)
		}
		srcInfo, _ := os.Stat(src)
		perm := os.FileMode(0o666)
		if srcInfo != nil {
			perm = srcInfo.Mode().Perm()
		}
		err = os.WriteFile(dest, data, perm)
		if err != nil {
			throwFSError(err, "open", dest)
		}
		return goja.Undefined()
	})

	// ── chmodSync ─────────────────────────────────────────────────────

	fsObj.Set("chmodSync", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		mode := os.FileMode(call.Argument(1).ToInteger())
		err := os.Chmod(path, mode)
		if err != nil {
			throwFSError(err, "chmod", path)
		}
		return goja.Undefined()
	})

	// ── chownSync ─────────────────────────────────────────────────────

	fsObj.Set("chownSync", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		uid := int(call.Argument(1).ToInteger())
		gid := int(call.Argument(2).ToInteger())
		err := os.Chown(path, uid, gid)
		if err != nil {
			throwFSError(err, "chown", path)
		}
		return goja.Undefined()
	})

	// ── truncateSync ──────────────────────────────────────────────────

	fsObj.Set("truncateSync", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		length := int64(0)
		if len(call.Arguments) > 1 {
			length = call.Argument(1).ToInteger()
		}
		err := os.Truncate(path, length)
		if err != nil {
			throwFSError(err, "truncate", path)
		}
		return goja.Undefined()
	})

	// ── mkdtempSync ───────────────────────────────────────────────────

	fsObj.Set("mkdtempSync", func(call goja.FunctionCall) goja.Value {
		prefix := call.Argument(0).String()
		dir := filepath.Dir(prefix)
		pattern := filepath.Base(prefix) + "*"
		tmpDir, err := os.MkdirTemp(dir, pattern)
		if err != nil {
			throwFSError(err, "mkdtemp", prefix)
		}
		return vm.ToValue(tmpDir)
	})

	// ── realpathSync ──────────────────────────────────────────────────

	fsObj.Set("realpathSync", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		resolved, err := filepath.EvalSymlinks(path)
		if err != nil {
			throwFSError(err, "realpath", path)
		}
		abs, err := filepath.Abs(resolved)
		if err != nil {
			throwFSError(err, "realpath", path)
		}
		return vm.ToValue(abs)
	})

	// ── readlinkSync ──────────────────────────────────────────────────

	fsObj.Set("readlinkSync", func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		target, err := os.Readlink(path)
		if err != nil {
			throwFSError(err, "readlink", path)
		}
		return vm.ToValue(target)
	})

	// ── symlinkSync ───────────────────────────────────────────────────

	fsObj.Set("symlinkSync", func(call goja.FunctionCall) goja.Value {
		target := call.Argument(0).String()
		path := call.Argument(1).String()
		err := os.Symlink(target, path)
		if err != nil {
			throwFSError(err, "symlink", path)
		}
		return goja.Undefined()
	})

	// ── linkSync ──────────────────────────────────────────────────────

	fsObj.Set("linkSync", func(call goja.FunctionCall) goja.Value {
		existingPath := call.Argument(0).String()
		newPath := call.Argument(1).String()
		err := os.Link(existingPath, newPath)
		if err != nil {
			throwFSError(err, "link", newPath)
		}
		return goja.Undefined()
	})

	vm.Set("__fs", fsObj)
}
