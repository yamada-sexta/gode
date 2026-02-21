package modules

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dop251/goja"
)

// ---------------------------------------------------------------------------
// TransformESM unit tests
// ---------------------------------------------------------------------------

func TestTransformESM_NoESM(t *testing.T) {
	src := `var x = 1; console.log(x);`
	got := TransformESM(src)
	if got != src {
		t.Fatalf("expected unchanged source, got:\n%s", got)
	}
}

func TestTransformESM_NamedImport(t *testing.T) {
	src := `import { foo, bar } from "mod";`
	got := TransformESM(src)
	if !strings.Contains(got, `require("mod")`) {
		t.Fatalf("expected require call, got:\n%s", got)
	}
	if !strings.Contains(got, "var foo") || !strings.Contains(got, "var bar") {
		t.Fatalf("expected named bindings, got:\n%s", got)
	}
}

func TestTransformESM_NamedImportAlias(t *testing.T) {
	src := `import { foo as bar } from "mod";`
	got := TransformESM(src)
	if !strings.Contains(got, "var bar") {
		t.Fatalf("expected alias binding, got:\n%s", got)
	}
	if strings.Contains(got, "var foo") {
		t.Fatalf("should not have original name binding, got:\n%s", got)
	}
}

func TestTransformESM_DefaultImport(t *testing.T) {
	src := `import myMod from "mod";`
	got := TransformESM(src)
	if !strings.Contains(got, `var myMod = require("mod")`) {
		t.Fatalf("expected default import transform, got:\n%s", got)
	}
}

func TestTransformESM_NamespaceImport(t *testing.T) {
	src := `import * as ns from "mod";`
	got := TransformESM(src)
	if !strings.Contains(got, `var ns = require("mod")`) {
		t.Fatalf("expected namespace import transform, got:\n%s", got)
	}
}

func TestTransformESM_BareImport(t *testing.T) {
	src := `import "side-effect";`
	got := TransformESM(src)
	if !strings.Contains(got, `require("side-effect")`) {
		t.Fatalf("expected bare import transform, got:\n%s", got)
	}
}

func TestTransformESM_DefaultAndNamed(t *testing.T) {
	src := `import def, { a, b } from "mod";`
	got := TransformESM(src)
	if !strings.Contains(got, `require("mod")`) {
		t.Fatalf("expected require call, got:\n%s", got)
	}
	if !strings.Contains(got, "var def") {
		t.Fatalf("expected default binding, got:\n%s", got)
	}
	if !strings.Contains(got, "var a") || !strings.Contains(got, "var b") {
		t.Fatalf("expected named bindings, got:\n%s", got)
	}
}

func TestTransformESM_ExportDefault(t *testing.T) {
	src := `export default myFunc;`
	got := TransformESM(src)
	if !strings.Contains(got, "module.exports = myFunc") {
		t.Fatalf("expected export default transform, got:\n%s", got)
	}
}

func TestTransformESM_ExportNamed(t *testing.T) {
	src := `export { foo, bar };`
	got := TransformESM(src)
	if !strings.Contains(got, "module.exports.foo = foo") || !strings.Contains(got, "module.exports.bar = bar") {
		t.Fatalf("expected named export transform, got:\n%s", got)
	}
}

func TestTransformESM_ExportNamedAlias(t *testing.T) {
	src := `export { foo as baz };`
	got := TransformESM(src)
	if !strings.Contains(got, "module.exports.baz = foo") {
		t.Fatalf("expected aliased named export, got:\n%s", got)
	}
}

func TestTransformESM_ExportDecl(t *testing.T) {
	src := `export const x = 42;`
	got := TransformESM(src)
	if !strings.Contains(got, "const x = 42") && !strings.Contains(got, "var x = 42") {
		t.Fatalf("expected declaration, got:\n%s", got)
	}
	if !strings.Contains(got, "module.exports.x = x") {
		t.Fatalf("expected export assignment, got:\n%s", got)
	}
}

func TestTransformESM_ExportFunction(t *testing.T) {
	src := `export function greet() { return "hi"; }`
	got := TransformESM(src)
	if !strings.Contains(got, `function greet()`) {
		t.Fatalf("expected function declaration, got:\n%s", got)
	}
	if !strings.Contains(got, "module.exports.greet = greet") {
		t.Fatalf("expected export assignment, got:\n%s", got)
	}
}

func TestTransformESM_ExportClass(t *testing.T) {
	src := `export class Foo {}`
	got := TransformESM(src)
	if !strings.Contains(got, "class Foo") {
		t.Fatalf("expected class declaration, got:\n%s", got)
	}
	if !strings.Contains(got, "module.exports.Foo = Foo") {
		t.Fatalf("expected export assignment, got:\n%s", got)
	}
}

func TestTransformESM_MixedContent(t *testing.T) {
	src := `import { readFileSync } from "fs";
var data = readFileSync("test.txt", "utf8");
console.log(data);`
	got := TransformESM(src)
	if !strings.Contains(got, `require("fs")`) {
		t.Fatalf("expected require, got:\n%s", got)
	}
	if !strings.Contains(got, "var readFileSync") {
		t.Fatalf("expected named binding, got:\n%s", got)
	}
	// Non-import lines should be unchanged
	if !strings.Contains(got, `var data = readFileSync("test.txt", "utf8");`) {
		t.Fatalf("expected unchanged line, got:\n%s", got)
	}
}

func TestTransformESM_SingleQuotes(t *testing.T) {
	src := `import { foo } from 'mod';`
	got := TransformESM(src)
	if !strings.Contains(got, `require("mod")`) {
		t.Fatalf("expected require with double quotes, got:\n%s", got)
	}
}

// ---------------------------------------------------------------------------
// Integration: require a module that uses ESM import
// ---------------------------------------------------------------------------

func TestRequire_ESMImport(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "node_modules", "esm-pkg"), 0o755)

	// A package that exports using module.exports (CJS)
	pkgJSON := `{"name":"esm-pkg","main":"./index.js"}`
	os.WriteFile(filepath.Join(dir, "node_modules", "esm-pkg", "package.json"), []byte(pkgJSON), 0o644)
	os.WriteFile(filepath.Join(dir, "node_modules", "esm-pkg", "index.js"), []byte(`module.exports.greet = function() { return "hello"; };`), 0o644)

	// Entry file uses ESM import syntax
	entry := filepath.Join(dir, "entry.js")
	os.WriteFile(entry, []byte(`import { greet } from "esm-pkg";
var result = greet();
`), 0o644)

	vm := goja.New()
	NewLoader(vm)
	abs, _ := filepath.Abs(entry)
	vm.Set("__filename", abs)
	vm.Set("__dirname", filepath.Dir(abs))

	// The entry file itself must also be transformed
	src, _ := os.ReadFile(entry)
	_, err := vm.RunString(TransformESM(string(src)))
	if err != nil {
		t.Fatalf("ESM import should work after transform: %v", err)
	}

	v := vm.Get("result")
	if v.String() != "hello" {
		t.Fatalf("expected 'hello', got %q", v.String())
	}
}
