package runner

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dop251/goja"
)

func newVM(t *testing.T) *goja.Runtime {
	t.Helper()
	return goja.New()
}

// ---------------------------------------------------------------------------
// ExecEval
// ---------------------------------------------------------------------------

func TestExecEval_ValidScript(t *testing.T) {
	vm := newVM(t)
	err := ExecEval(vm, `var x = 1 + 2`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	val := vm.Get("x")
	if val.ToInteger() != 3 {
		t.Fatalf("expected 3, got %v", val)
	}
}

func TestExecEval_ES6Features(t *testing.T) {
	vm := newVM(t)
	err := ExecEval(vm, `
		const add = (a, b) => a + b;
		let result = add(2, 3);
	`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExecEval_SyntaxError(t *testing.T) {
	vm := newVM(t)
	err := ExecEval(vm, `function(`)
	if err == nil {
		t.Fatal("expected syntax error")
	}
}

func TestExecEval_RuntimeError(t *testing.T) {
	vm := newVM(t)
	err := ExecEval(vm, `undefinedVar.property`)
	if err == nil {
		t.Fatal("expected runtime error")
	}
}

func TestExecEval_SetsGlobals(t *testing.T) {
	vm := newVM(t)
	ExecEval(vm, `var greeting = "hello"`)
	val := vm.Get("greeting")
	if val.String() != "hello" {
		t.Fatalf("expected 'hello', got %q", val.String())
	}
}

// ---------------------------------------------------------------------------
// ExecPrint
// ---------------------------------------------------------------------------

func TestExecPrint_ReturnsValue(t *testing.T) {
	vm := newVM(t)
	val, err := ExecPrint(vm, `1 + 2`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.ToInteger() != 3 {
		t.Fatalf("expected 3, got %v", val)
	}
}

func TestExecPrint_ReturnsString(t *testing.T) {
	vm := newVM(t)
	val, err := ExecPrint(vm, `"hello" + " " + "world"`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.String() != "hello world" {
		t.Fatalf("expected 'hello world', got %q", val.String())
	}
}

func TestExecPrint_Error(t *testing.T) {
	vm := newVM(t)
	_, err := ExecPrint(vm, `throw new Error("boom")`)
	if err == nil {
		t.Fatal("expected error")
	}
}

// ---------------------------------------------------------------------------
// ExecFile
// ---------------------------------------------------------------------------

func TestExecFile_ValidScript(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.js")
	os.WriteFile(path, []byte(`var fileResult = 42`), 0644)

	vm := newVM(t)
	err := ExecFile(vm, path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	val := vm.Get("fileResult")
	if val.ToInteger() != 42 {
		t.Fatalf("expected 42, got %v", val)
	}
}

func TestExecFile_SetsFilenameAndDirname(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.js")
	os.WriteFile(path, []byte(`/* empty */`), 0644)

	vm := newVM(t)
	ExecFile(vm, path)

	abs, _ := filepath.Abs(path)
	fn := vm.Get("__filename")
	if fn.String() != abs {
		t.Fatalf("expected __filename=%q, got %q", abs, fn.String())
	}
	dn := vm.Get("__dirname")
	if dn.String() != filepath.Dir(abs) {
		t.Fatalf("expected __dirname=%q, got %q", filepath.Dir(abs), dn.String())
	}
}

func TestExecFile_ES6Syntax(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "es6.js")
	os.WriteFile(path, []byte(`const x = [1, ...[2,3]]; var es6len = x.length;`), 0644)

	vm := newVM(t)
	err := ExecFile(vm, path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	val := vm.Get("es6len")
	if val.ToInteger() != 3 {
		t.Fatalf("expected 3, got %v", val)
	}
}

func TestExecFile_FileNotFound(t *testing.T) {
	vm := newVM(t)
	err := ExecFile(vm, "/nonexistent/file.js")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestExecFile_SyntaxError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.js")
	os.WriteFile(path, []byte(`function(`), 0644)

	vm := newVM(t)
	err := ExecFile(vm, path)
	if err == nil {
		t.Fatal("expected syntax error")
	}
}

// ---------------------------------------------------------------------------
// ExecStdin
// ---------------------------------------------------------------------------

func TestExecStdin_ValidScript(t *testing.T) {
	vm := newVM(t)
	r := strings.NewReader(`var stdinResult = "from stdin"`)
	err := ExecStdin(vm, r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	val := vm.Get("stdinResult")
	if val.String() != "from stdin" {
		t.Fatalf("expected 'from stdin', got %q", val.String())
	}
}

func TestExecStdin_SyntaxError(t *testing.T) {
	vm := newVM(t)
	r := strings.NewReader(`function(`)
	err := ExecStdin(vm, r)
	if err == nil {
		t.Fatal("expected syntax error")
	}
}

func TestExecStdin_Empty(t *testing.T) {
	vm := newVM(t)
	r := strings.NewReader(``)
	err := ExecStdin(vm, r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// ValidateSyntax
// ---------------------------------------------------------------------------

func TestValidateSyntax_ValidFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "valid.js")
	os.WriteFile(path, []byte(`const x = 1; function foo() { return x; }`), 0644)

	err := ValidateSyntax(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateSyntax_InvalidFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "invalid.js")
	os.WriteFile(path, []byte(`function(`), 0644)

	err := ValidateSyntax(path)
	if err == nil {
		t.Fatal("expected syntax error")
	}
}

func TestValidateSyntax_FileNotFound(t *testing.T) {
	err := ValidateSyntax("/nonexistent/file.js")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestValidateSyntax_ES6Syntax(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "es6.js")
	os.WriteFile(path, []byte(`
		const x = 1;
		let y = 2;
		const add = (a, b) => a + b;
		const tpl = `+"`hello ${x}`"+`;
		class Foo { constructor() {} }
	`), 0644)

	err := ValidateSyntax(path)
	if err != nil {
		t.Fatalf("ES6 syntax should be valid: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Source Map Loading
// ---------------------------------------------------------------------------

func TestExecFile_SourceMapMissing(t *testing.T) {
	// A JS file with a sourceMappingURL pointing to a nonexistent .map
	// file should execute without error (map loading is best-effort).
	dir := t.TempDir()
	path := filepath.Join(dir, "with_map.js")
	os.WriteFile(path, []byte(`var smResult = 42;
//# sourceMappingURL=with_map.js.map
`), 0644)

	vm := newVM(t)
	err := ExecFile(vm, path)
	if err != nil {
		t.Fatalf("sourceMappingURL with missing .map should not error: %v", err)
	}
	val := vm.Get("smResult")
	if val.ToInteger() != 42 {
		t.Fatalf("expected 42, got %v", val)
	}
}

func TestExecFile_SourceMapPresent(t *testing.T) {
	// When the .map file exists, it should be loaded without error.
	dir := t.TempDir()
	path := filepath.Join(dir, "mapped.js")
	os.WriteFile(path, []byte(`var mapped = "ok";
//# sourceMappingURL=mapped.js.map
`), 0644)

	// Write a minimal valid source map
	mapContent := `{"version":3,"sources":["mapped.ts"],"names":[],"mappings":"AAAA","file":"mapped.js"}`
	os.WriteFile(filepath.Join(dir, "mapped.js.map"), []byte(mapContent), 0644)

	vm := newVM(t)
	err := ExecFile(vm, path)
	if err != nil {
		t.Fatalf("sourceMappingURL with present .map should not error: %v", err)
	}
	val := vm.Get("mapped")
	if val.String() != "ok" {
		t.Fatalf("expected 'ok', got %q", val.String())
	}
}
