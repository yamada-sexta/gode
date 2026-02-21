package modules

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dop251/goja"
)

// ---------------------------------------------------------------------------
// TransformESM unit tests — verify esbuild output is valid CJS
// ---------------------------------------------------------------------------

func TestTransformESM_NoESM(t *testing.T) {
	src := `var x = 1; console.log(x);`
	got := TransformESM(src)
	// Should still produce valid output (esbuild may reformat slightly)
	if !strings.Contains(got, "var x = 1") {
		t.Fatalf("expected source preserved, got:\n%s", got)
	}
}

func TestTransformESM_NamedImport(t *testing.T) {
	src := `import { foo, bar } from "mod";`
	got := TransformESM(src)
	if !strings.Contains(got, `require("mod")`) {
		t.Fatalf("expected require call, got:\n%s", got)
	}
}

func TestTransformESM_DefaultImport(t *testing.T) {
	src := `import myMod from "mod";`
	got := TransformESM(src)
	if !strings.Contains(got, `require("mod")`) {
		t.Fatalf("expected require call, got:\n%s", got)
	}
}

func TestTransformESM_NamespaceImport(t *testing.T) {
	src := `import * as ns from "mod";`
	got := TransformESM(src)
	if !strings.Contains(got, `require("mod")`) {
		t.Fatalf("expected require call, got:\n%s", got)
	}
}

func TestTransformESM_BareImport(t *testing.T) {
	src := `import "side-effect";`
	got := TransformESM(src)
	if !strings.Contains(got, `require("side-effect")`) {
		t.Fatalf("expected require call, got:\n%s", got)
	}
}

func TestTransformESM_ExportDefault(t *testing.T) {
	src := `export default 42;`
	got := TransformESM(src)
	if !strings.Contains(got, "module.exports") || !strings.Contains(got, "exports") {
		t.Fatalf("expected module.exports, got:\n%s", got)
	}
}

func TestTransformESM_ExportNamed(t *testing.T) {
	src := `var foo = 1; var bar = 2; export { foo, bar };`
	got := TransformESM(src)
	if !strings.Contains(got, "exports") {
		t.Fatalf("expected exports, got:\n%s", got)
	}
}

func TestTransformESM_ExportDecl(t *testing.T) {
	src := `export const x = 42;`
	got := TransformESM(src)
	if !strings.Contains(got, "42") || !strings.Contains(got, "exports") {
		t.Fatalf("expected declaration + export, got:\n%s", got)
	}
}

func TestTransformESM_ForAwaitOf(t *testing.T) {
	// for-await-of should be transpiled so goja can parse it
	src := `async function test() { for await (const x of [1,2]) { console.log(x); } }`
	got := TransformESM(src)
	// esbuild should transform for-await-of into something goja can handle
	vm := goja.New()
	_, err := vm.RunString(got)
	if err != nil {
		t.Fatalf("for-await-of should be transpiled: %v\ntransformed:\n%s", err, got)
	}
}

func TestTransformESM_ParseError(t *testing.T) {
	// Invalid JS: esbuild should fail, and we return original source
	src := `this is not valid javascript @@@@`
	got := TransformESM(src)
	if got != src {
		t.Fatalf("on parse error, should return original source")
	}
}

// ---------------------------------------------------------------------------
// Integration: actual execution in goja
// ---------------------------------------------------------------------------

func TestTransformESM_NamedImportExecution(t *testing.T) {
	// Verify that esbuild's CJS output for named imports actually runs in goja
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "node_modules", "esm-pkg"), 0o755)

	pkgJSON := `{"name":"esm-pkg","main":"./index.js"}`
	os.WriteFile(filepath.Join(dir, "node_modules", "esm-pkg", "package.json"), []byte(pkgJSON), 0o644)
	os.WriteFile(filepath.Join(dir, "node_modules", "esm-pkg", "index.js"),
		[]byte(`module.exports.greet = function() { return "hello"; };`), 0o644)

	vm := goja.New()
	NewLoader(vm)
	abs, _ := filepath.Abs(filepath.Join(dir, "entry.js"))
	vm.Set("__filename", abs)
	vm.Set("__dirname", filepath.Dir(abs))

	src := `import { greet } from "esm-pkg";
var result = greet();`
	_, err := vm.RunString(TransformESM(src))
	if err != nil {
		t.Fatalf("ESM import should work: %v", err)
	}

	v := vm.Get("result")
	if v == nil || v.String() != "hello" {
		t.Fatalf("expected 'hello', got %v", v)
	}
}

func TestTransformESM_DefaultImportExecution(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "node_modules", "def-pkg"), 0o755)

	pkgJSON := `{"name":"def-pkg","main":"./index.js"}`
	os.WriteFile(filepath.Join(dir, "node_modules", "def-pkg", "package.json"), []byte(pkgJSON), 0o644)
	os.WriteFile(filepath.Join(dir, "node_modules", "def-pkg", "index.js"),
		[]byte(`module.exports = function() { return "default"; };`), 0o644)

	vm := goja.New()
	NewLoader(vm)
	abs, _ := filepath.Abs(filepath.Join(dir, "entry.js"))
	vm.Set("__filename", abs)
	vm.Set("__dirname", filepath.Dir(abs))

	src := `import myFunc from "def-pkg";
var result = myFunc();`
	_, err := vm.RunString(TransformESM(src))
	if err != nil {
		t.Fatalf("default import should work: %v", err)
	}

	v := vm.Get("result")
	if v == nil || v.String() != "default" {
		t.Fatalf("expected 'default', got %v", v)
	}
}

func TestTransformESM_ExportExecution(t *testing.T) {
	// Verify export const works with module.exports set up
	vm := goja.New()
	moduleObj := vm.NewObject()
	exportsObj := vm.NewObject()
	moduleObj.Set("exports", exportsObj)
	vm.Set("module", moduleObj)
	vm.Set("exports", exportsObj)

	src := `export const answer = 42;`
	_, err := vm.RunString(TransformESM(src))
	if err != nil {
		t.Fatalf("export const should work: %v", err)
	}

	result, _ := vm.RunString(`module.exports.answer`)
	if result == nil || result.ToInteger() != 42 {
		t.Fatalf("expected 42, got %v", result)
	}
}

func TestTransformESM_TopLevelAwaitExecution(t *testing.T) {
	vm := goja.New()
	src := `var x = await Promise.resolve(42);`
	transformed := TransformESM(src)
	_, err := vm.RunString(transformed)
	if err != nil {
		t.Fatalf("top-level await should execute: %v\ntransformed:\n%s", err, transformed)
	}
}
