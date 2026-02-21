package modules

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dop251/goja"
)

func fsVM(t *testing.T) *goja.Runtime {
	t.Helper()
	vm := goja.New()
	NewLoader(vm)
	mustRunF(t, vm, `var fs = require('fs')`)
	return vm
}

func mustRunF(t *testing.T, vm *goja.Runtime, js string) goja.Value {
	t.Helper()
	val, err := vm.RunString(js)
	if err != nil {
		t.Fatalf("unexpected error: %v\nscript: %s", err, js)
	}
	return val
}

// ---------------------------------------------------------------------------
// require
// ---------------------------------------------------------------------------

func TestFSRequire(t *testing.T) {
	vm := goja.New()
	NewLoader(vm)
	mustRunF(t, vm, `var f = require('fs'); if (!f.readFileSync) throw new Error('missing readFileSync')`)
}

func TestFSRequireNodePrefix(t *testing.T) {
	vm := goja.New()
	NewLoader(vm)
	mustRunF(t, vm, `var f = require('node:fs'); if (!f.readFileSync) throw new Error('missing readFileSync')`)
}

// ---------------------------------------------------------------------------
// readFileSync / writeFileSync
// ---------------------------------------------------------------------------

func TestFS_ReadWriteFileSync(t *testing.T) {
	vm := fsVM(t)
	dir := t.TempDir()
	p := filepath.Join(dir, "test.txt")
	vm.Set("__path", p)

	mustRunF(t, vm, `fs.writeFileSync(__path, 'hello world')`)

	val := mustRunF(t, vm, `fs.readFileSync(__path, 'utf8')`)
	if val.String() != "hello world" {
		t.Fatalf("expected 'hello world', got %q", val.String())
	}
}

func TestFS_ReadFileSyncNotFound(t *testing.T) {
	vm := fsVM(t)
	_, err := vm.RunString(`fs.readFileSync('/nonexistent/file.txt')`)
	if err == nil {
		t.Fatal("expected error for missing file")
	}
	s := err.Error()
	if !strings.Contains(s, "ENOENT") {
		t.Fatalf("expected ENOENT error, got: %s", s)
	}
}

func TestFS_ReadFileSyncEncoding(t *testing.T) {
	vm := fsVM(t)
	dir := t.TempDir()
	p := filepath.Join(dir, "enc.txt")
	os.WriteFile(p, []byte("encoded"), 0o644)
	vm.Set("__path", p)

	val := mustRunF(t, vm, `fs.readFileSync(__path, { encoding: 'utf8' })`)
	if val.String() != "encoded" {
		t.Fatalf("expected 'encoded', got %q", val.String())
	}
}

// ---------------------------------------------------------------------------
// appendFileSync
// ---------------------------------------------------------------------------

func TestFS_AppendFileSync(t *testing.T) {
	vm := fsVM(t)
	dir := t.TempDir()
	p := filepath.Join(dir, "append.txt")
	vm.Set("__path", p)

	mustRunF(t, vm, `fs.writeFileSync(__path, 'a')`)
	mustRunF(t, vm, `fs.appendFileSync(__path, 'b')`)

	val := mustRunF(t, vm, `fs.readFileSync(__path, 'utf8')`)
	if val.String() != "ab" {
		t.Fatalf("expected 'ab', got %q", val.String())
	}
}

// ---------------------------------------------------------------------------
// existsSync
// ---------------------------------------------------------------------------

func TestFS_ExistsSync(t *testing.T) {
	vm := fsVM(t)
	dir := t.TempDir()
	p := filepath.Join(dir, "exists.txt")
	os.WriteFile(p, []byte("x"), 0o644)
	vm.Set("__path", p)

	val := mustRunF(t, vm, `fs.existsSync(__path)`)
	if !val.ToBoolean() {
		t.Fatal("expected true for existing file")
	}

	val = mustRunF(t, vm, `fs.existsSync('/nonexistent/path')`)
	if val.ToBoolean() {
		t.Fatal("expected false for nonexistent path")
	}
}

// ---------------------------------------------------------------------------
// statSync / lstatSync
// ---------------------------------------------------------------------------

func TestFS_StatSync(t *testing.T) {
	vm := fsVM(t)
	dir := t.TempDir()
	p := filepath.Join(dir, "stat.txt")
	os.WriteFile(p, []byte("hello"), 0o644)
	vm.Set("__path", p)

	mustRunF(t, vm, `
		var s = fs.statSync(__path);
		if (!s.isFile()) throw new Error('expected isFile true');
		if (s.isDirectory()) throw new Error('expected isDirectory false');
		if (s.size !== 5) throw new Error('expected size 5, got ' + s.size);
	`)
}

func TestFS_StatSyncDir(t *testing.T) {
	vm := fsVM(t)
	dir := t.TempDir()
	vm.Set("__path", dir)

	mustRunF(t, vm, `
		var s = fs.statSync(__path);
		if (s.isFile()) throw new Error('expected isFile false');
		if (!s.isDirectory()) throw new Error('expected isDirectory true');
	`)
}

func TestFS_StatSyncNotFound(t *testing.T) {
	vm := fsVM(t)
	_, err := vm.RunString(`fs.statSync('/nonexistent')`)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestFS_StatSyncThrowIfNoEntry(t *testing.T) {
	vm := fsVM(t)
	val := mustRunF(t, vm, `fs.statSync('/nonexistent', { throwIfNoEntry: false })`)
	if !goja.IsUndefined(val) {
		t.Fatalf("expected undefined, got %v", val)
	}
}

func TestFS_LstatSync(t *testing.T) {
	vm := fsVM(t)
	dir := t.TempDir()
	target := filepath.Join(dir, "target.txt")
	link := filepath.Join(dir, "link.txt")
	os.WriteFile(target, []byte("x"), 0o644)
	os.Symlink(target, link)
	vm.Set("__link", link)

	mustRunF(t, vm, `
		var s = fs.lstatSync(__link);
		if (!s.isSymbolicLink()) throw new Error('expected isSymbolicLink true');
	`)
}

// ---------------------------------------------------------------------------
// readdirSync
// ---------------------------------------------------------------------------

func TestFS_ReaddirSync(t *testing.T) {
	vm := fsVM(t)
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.txt"), []byte("a"), 0o644)
	os.WriteFile(filepath.Join(dir, "b.txt"), []byte("b"), 0o644)
	os.Mkdir(filepath.Join(dir, "sub"), 0o755)
	vm.Set("__dir", dir)

	val := mustRunF(t, vm, `JSON.stringify(fs.readdirSync(__dir).sort())`)
	s := val.String()
	if !strings.Contains(s, "a.txt") || !strings.Contains(s, "b.txt") || !strings.Contains(s, "sub") {
		t.Fatalf("unexpected readdir result: %s", s)
	}
}

func TestFS_ReaddirSyncWithFileTypes(t *testing.T) {
	vm := fsVM(t)
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "file.txt"), []byte("x"), 0o644)
	os.Mkdir(filepath.Join(dir, "dir"), 0o755)
	vm.Set("__dir", dir)

	mustRunF(t, vm, `
		var entries = fs.readdirSync(__dir, { withFileTypes: true });
		var file = entries.find(function(e) { return e.name === 'file.txt'; });
		var d = entries.find(function(e) { return e.name === 'dir'; });
		if (!file.isFile()) throw new Error('expected file.isFile()');
		if (!d.isDirectory()) throw new Error('expected dir.isDirectory()');
	`)
}

// ---------------------------------------------------------------------------
// mkdirSync / rmdirSync
// ---------------------------------------------------------------------------

func TestFS_MkdirSync(t *testing.T) {
	vm := fsVM(t)
	dir := t.TempDir()
	p := filepath.Join(dir, "newdir")
	vm.Set("__path", p)

	mustRunF(t, vm, `fs.mkdirSync(__path)`)

	info, err := os.Stat(p)
	if err != nil || !info.IsDir() {
		t.Fatal("directory not created")
	}
}

func TestFS_MkdirSyncRecursive(t *testing.T) {
	vm := fsVM(t)
	dir := t.TempDir()
	p := filepath.Join(dir, "a", "b", "c")
	vm.Set("__path", p)

	mustRunF(t, vm, `fs.mkdirSync(__path, { recursive: true })`)

	info, err := os.Stat(p)
	if err != nil || !info.IsDir() {
		t.Fatal("recursive directory not created")
	}
}

func TestFS_RmdirSync(t *testing.T) {
	vm := fsVM(t)
	dir := t.TempDir()
	p := filepath.Join(dir, "rmme")
	os.Mkdir(p, 0o755)
	vm.Set("__path", p)

	mustRunF(t, vm, `fs.rmdirSync(__path)`)

	if _, err := os.Stat(p); !os.IsNotExist(err) {
		t.Fatal("directory not removed")
	}
}

// ---------------------------------------------------------------------------
// rmSync
// ---------------------------------------------------------------------------

func TestFS_RmSync(t *testing.T) {
	vm := fsVM(t)
	dir := t.TempDir()
	p := filepath.Join(dir, "rmme")
	os.Mkdir(p, 0o755)
	os.WriteFile(filepath.Join(p, "child.txt"), []byte("x"), 0o644)
	vm.Set("__path", p)

	mustRunF(t, vm, `fs.rmSync(__path, { recursive: true })`)

	if _, err := os.Stat(p); !os.IsNotExist(err) {
		t.Fatal("directory not removed recursively")
	}
}

func TestFS_RmSyncForce(t *testing.T) {
	vm := fsVM(t)
	// force: true should not throw on nonexistent path
	mustRunF(t, vm, `fs.rmSync('/nonexistent/file', { force: true })`)
}

// ---------------------------------------------------------------------------
// unlinkSync
// ---------------------------------------------------------------------------

func TestFS_UnlinkSync(t *testing.T) {
	vm := fsVM(t)
	dir := t.TempDir()
	p := filepath.Join(dir, "unlink.txt")
	os.WriteFile(p, []byte("x"), 0o644)
	vm.Set("__path", p)

	mustRunF(t, vm, `fs.unlinkSync(__path)`)

	if _, err := os.Stat(p); !os.IsNotExist(err) {
		t.Fatal("file not removed")
	}
}

// ---------------------------------------------------------------------------
// renameSync
// ---------------------------------------------------------------------------

func TestFS_RenameSync(t *testing.T) {
	vm := fsVM(t)
	dir := t.TempDir()
	old := filepath.Join(dir, "old.txt")
	new := filepath.Join(dir, "new.txt")
	os.WriteFile(old, []byte("renamed"), 0o644)
	vm.Set("__old", old)
	vm.Set("__new", new)

	mustRunF(t, vm, `fs.renameSync(__old, __new)`)

	data, err := os.ReadFile(new)
	if err != nil || string(data) != "renamed" {
		t.Fatal("rename failed")
	}
	if _, err := os.Stat(old); !os.IsNotExist(err) {
		t.Fatal("old file still exists")
	}
}

// ---------------------------------------------------------------------------
// copyFileSync
// ---------------------------------------------------------------------------

func TestFS_CopyFileSync(t *testing.T) {
	vm := fsVM(t)
	dir := t.TempDir()
	src := filepath.Join(dir, "src.txt")
	dest := filepath.Join(dir, "dest.txt")
	os.WriteFile(src, []byte("copied"), 0o644)
	vm.Set("__src", src)
	vm.Set("__dest", dest)

	mustRunF(t, vm, `fs.copyFileSync(__src, __dest)`)

	data, err := os.ReadFile(dest)
	if err != nil || string(data) != "copied" {
		t.Fatal("copy failed")
	}
}

func TestFS_CopyFileSyncExcl(t *testing.T) {
	vm := fsVM(t)
	dir := t.TempDir()
	src := filepath.Join(dir, "src.txt")
	dest := filepath.Join(dir, "dest.txt")
	os.WriteFile(src, []byte("x"), 0o644)
	os.WriteFile(dest, []byte("y"), 0o644)
	vm.Set("__src", src)
	vm.Set("__dest", dest)

	_, err := vm.RunString(`fs.copyFileSync(__src, __dest, fs.constants.COPYFILE_EXCL)`)
	if err == nil {
		t.Fatal("expected EEXIST error")
	}
}

// ---------------------------------------------------------------------------
// chmodSync
// ---------------------------------------------------------------------------

func TestFS_ChmodSync(t *testing.T) {
	vm := fsVM(t)
	dir := t.TempDir()
	p := filepath.Join(dir, "chmod.txt")
	os.WriteFile(p, []byte("x"), 0o644)
	vm.Set("__path", p)

	mustRunF(t, vm, `fs.chmodSync(__path, 0o755)`)

	info, _ := os.Stat(p)
	if info.Mode().Perm() != 0o755 {
		t.Fatalf("expected 0755, got %o", info.Mode().Perm())
	}
}

// ---------------------------------------------------------------------------
// truncateSync
// ---------------------------------------------------------------------------

func TestFS_TruncateSync(t *testing.T) {
	vm := fsVM(t)
	dir := t.TempDir()
	p := filepath.Join(dir, "trunc.txt")
	os.WriteFile(p, []byte("hello world"), 0o644)
	vm.Set("__path", p)

	mustRunF(t, vm, `fs.truncateSync(__path, 5)`)

	data, _ := os.ReadFile(p)
	if string(data) != "hello" {
		t.Fatalf("expected 'hello', got %q", string(data))
	}
}

// ---------------------------------------------------------------------------
// mkdtempSync
// ---------------------------------------------------------------------------

func TestFS_MkdtempSync(t *testing.T) {
	vm := fsVM(t)
	dir := t.TempDir()
	prefix := filepath.Join(dir, "prefix-")
	vm.Set("__prefix", prefix)

	val := mustRunF(t, vm, `fs.mkdtempSync(__prefix)`)
	created := val.String()

	if !strings.HasPrefix(created, filepath.Join(dir, "prefix-")) {
		t.Fatalf("unexpected prefix: %s", created)
	}
	info, err := os.Stat(created)
	if err != nil || !info.IsDir() {
		t.Fatal("temp dir not created")
	}
}

// ---------------------------------------------------------------------------
// realpathSync
// ---------------------------------------------------------------------------

func TestFS_RealpathSync(t *testing.T) {
	vm := fsVM(t)
	dir := t.TempDir()
	p := filepath.Join(dir, "real.txt")
	os.WriteFile(p, []byte("x"), 0o644)
	vm.Set("__path", p)

	val := mustRunF(t, vm, `fs.realpathSync(__path)`)
	abs, _ := filepath.Abs(p)
	resolved, _ := filepath.EvalSymlinks(abs)
	if val.String() != resolved {
		t.Fatalf("expected %q, got %q", resolved, val.String())
	}
}

// ---------------------------------------------------------------------------
// symlinkSync / readlinkSync
// ---------------------------------------------------------------------------

func TestFS_SymlinkReadlinkSync(t *testing.T) {
	vm := fsVM(t)
	dir := t.TempDir()
	target := filepath.Join(dir, "target.txt")
	link := filepath.Join(dir, "link.txt")
	os.WriteFile(target, []byte("sym"), 0o644)
	vm.Set("__target", target)
	vm.Set("__link", link)

	mustRunF(t, vm, `fs.symlinkSync(__target, __link)`)

	val := mustRunF(t, vm, `fs.readlinkSync(__link)`)
	if val.String() != target {
		t.Fatalf("expected %q, got %q", target, val.String())
	}
}

// ---------------------------------------------------------------------------
// linkSync
// ---------------------------------------------------------------------------

func TestFS_LinkSync(t *testing.T) {
	vm := fsVM(t)
	dir := t.TempDir()
	src := filepath.Join(dir, "src.txt")
	dst := filepath.Join(dir, "hard.txt")
	os.WriteFile(src, []byte("hardlink"), 0o644)
	vm.Set("__src", src)
	vm.Set("__dst", dst)

	mustRunF(t, vm, `fs.linkSync(__src, __dst)`)

	data, err := os.ReadFile(dst)
	if err != nil || string(data) != "hardlink" {
		t.Fatal("hard link not created")
	}
}

// ---------------------------------------------------------------------------
// accessSync
// ---------------------------------------------------------------------------

func TestFS_AccessSync(t *testing.T) {
	vm := fsVM(t)
	dir := t.TempDir()
	p := filepath.Join(dir, "access.txt")
	os.WriteFile(p, []byte("x"), 0o644)
	vm.Set("__path", p)

	// F_OK should not throw
	mustRunF(t, vm, `fs.accessSync(__path)`)
	mustRunF(t, vm, `fs.accessSync(__path, fs.constants.R_OK)`)
}

func TestFS_AccessSyncNotFound(t *testing.T) {
	vm := fsVM(t)
	_, err := vm.RunString(`fs.accessSync('/nonexistent')`)
	if err == nil {
		t.Fatal("expected error")
	}
}

// ---------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------

func TestFS_Constants(t *testing.T) {
	vm := fsVM(t)
	mustRunF(t, vm, `
		if (typeof fs.constants !== 'object') throw new Error('missing constants');
		if (fs.constants.F_OK !== 0) throw new Error('wrong F_OK');
		if (fs.constants.R_OK !== 4) throw new Error('wrong R_OK');
		if (fs.constants.W_OK !== 2) throw new Error('wrong W_OK');
		if (fs.constants.X_OK !== 1) throw new Error('wrong X_OK');
	`)
}

func TestFS_TopLevelConstants(t *testing.T) {
	vm := fsVM(t)
	mustRunF(t, vm, `
		if (fs.F_OK !== 0) throw new Error('wrong fs.F_OK');
		if (fs.R_OK !== 4) throw new Error('wrong fs.R_OK');
	`)
}

// ---------------------------------------------------------------------------
// All functions exist
// ---------------------------------------------------------------------------

func TestFS_AllFunctionsExist(t *testing.T) {
	vm := fsVM(t)
	fns := []string{
		"readFileSync", "writeFileSync", "appendFileSync",
		"existsSync", "accessSync",
		"statSync", "lstatSync",
		"readdirSync",
		"mkdirSync", "rmdirSync", "rmSync",
		"unlinkSync", "renameSync",
		"copyFileSync", "chmodSync", "chownSync",
		"truncateSync", "mkdtempSync",
		"realpathSync", "readlinkSync",
		"symlinkSync", "linkSync",
	}
	for _, fn := range fns {
		t.Run(fn, func(t *testing.T) {
			mustRunF(t, vm, `if (typeof fs.`+fn+` !== 'function') throw new Error('missing: `+fn+`')`)
		})
	}
}

// ---------------------------------------------------------------------------
// Error properties
// ---------------------------------------------------------------------------

func TestFS_ErrorProperties(t *testing.T) {
	vm := fsVM(t)
	mustRunF(t, vm, `
		try {
			fs.readFileSync('/nonexistent/xyz');
			throw new Error('should have thrown');
		} catch(e) {
			if (e.code !== 'ENOENT') throw new Error('expected code ENOENT, got ' + e.code);
			if (typeof e.errno !== 'number') throw new Error('expected errno to be number');
			if (e.syscall !== 'open') throw new Error('expected syscall open, got ' + e.syscall);
			if (e.path !== '/nonexistent/xyz') throw new Error('expected path, got ' + e.path);
		}
	`)
}
