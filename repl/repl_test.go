package repl

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/dop251/goja"
)

// ---------------------------------------------------------------------------
// isIncomplete — detects multi-line input
// ---------------------------------------------------------------------------

func TestIsIncomplete_CompleteSingleLine(t *testing.T) {
	tests := []string{
		`var x = 1`,
		`1 + 2`,
		`"hello"`,
		`function foo() {}`,
		`if (true) { 1 }`,
		`[1, 2, 3]`,
		`({a: 1})`,
	}
	for _, src := range tests {
		t.Run(src, func(t *testing.T) {
			if isIncomplete(src) {
				t.Fatalf("expected complete: %q", src)
			}
		})
	}
}

func TestIsIncomplete_IncompleteInput(t *testing.T) {
	tests := []struct {
		name, src string
	}{
		{"open brace", `function foo() {`},
		{"open paren", `foo(`},
		{"open bracket", `[1, 2,`},
		{"if without body", `if (true)`},
		{"object literal", `var x = {`},
		{"template literal", "var x = `hello"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if !isIncomplete(tc.src) {
				t.Fatalf("expected incomplete: %q", tc.src)
			}
		})
	}
}

func TestIsIncomplete_MultiLineComplete(t *testing.T) {
	src := "function foo() {\n  return 1;\n}"
	if isIncomplete(src) {
		t.Fatal("multi-line complete function should not be incomplete")
	}
}

func TestIsIncomplete_EmptyString(t *testing.T) {
	// Empty input: parser succeeds → not incomplete.
	if isIncomplete("") {
		t.Fatal("empty string should not be incomplete")
	}
}

// ---------------------------------------------------------------------------
// parseDotCommand
// ---------------------------------------------------------------------------

func TestParseDotCommand_Simple(t *testing.T) {
	tests := []struct {
		input, cmd, arg string
	}{
		{".exit", ".exit", ""},
		{".help", ".help", ""},
		{".break", ".break", ""},
		{".clear", ".clear", ""},
		{".editor", ".editor", ""},
		{".load foo.js", ".load", "foo.js"},
		{".save output.js", ".save", "output.js"},
		{".load  /path/to/file.js", ".load", "/path/to/file.js"},
	}
	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			cmd, arg := parseDotCommand(tc.input)
			if cmd != tc.cmd {
				t.Fatalf("expected cmd %q, got %q", tc.cmd, cmd)
			}
			if arg != tc.arg {
				t.Fatalf("expected arg %q, got %q", tc.arg, arg)
			}
		})
	}
}

func TestParseDotCommand_WithWhitespace(t *testing.T) {
	cmd, arg := parseDotCommand("  .load   test.js  ")
	if cmd != ".load" {
		t.Fatalf("expected .load, got %q", cmd)
	}
	if arg != "test.js" {
		t.Fatalf("expected 'test.js', got %q", arg)
	}
}

// ---------------------------------------------------------------------------
// loadFile helper
// ---------------------------------------------------------------------------

func TestLoadFile(t *testing.T) {
	vm := goja.New()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.js")
	os.WriteFile(path, []byte(`var loadedVar = 42`), 0644)

	var history []string
	loadFile(vm, path, &history)

	val := vm.Get("loadedVar")
	if val.ToInteger() != 42 {
		t.Fatalf("expected 42, got %v", val)
	}
	if len(history) != 1 {
		t.Fatalf("expected 1 history entry, got %d", len(history))
	}
}

func TestLoadFile_NotFound(t *testing.T) {
	vm := goja.New()
	var history []string
	// Should not panic, just print error.
	loadFile(vm, "/nonexistent/file.js", &history)
	if len(history) != 0 {
		t.Fatal("should not add to history on error")
	}
}

// ---------------------------------------------------------------------------
// saveSession helper
// ---------------------------------------------------------------------------

func TestSaveSession(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "session.js")
	history := []string{"var a = 1", "var b = 2", "a + b"}

	saveSession(path, history)

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read saved session: %v", err)
	}
	content := string(data)
	if content != "var a = 1\nvar b = 2\na + b\n" {
		t.Fatalf("unexpected content: %q", content)
	}
}

// ---------------------------------------------------------------------------
// Console integration (captured output)
// ---------------------------------------------------------------------------

func TestConsoleLogOutput(t *testing.T) {
	// Verify that console.log works by capturing stdout.
	// We redirect os.Stdout temporarily.
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	vm := goja.New()

	// Set up console manually for this isolated test.
	con := vm.NewObject()
	con.Set("log", func(call goja.FunctionCall) goja.Value {
		args := make([]interface{}, len(call.Arguments))
		for i, a := range call.Arguments {
			args[i] = a.Export()
		}
		fmt.Fprintln(os.Stdout, args...)
		return goja.Undefined()
	})
	vm.Set("console", con)

	vm.RunString(`console.log("test output", 42)`)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()
	if output != "test output 42\n" {
		t.Fatalf("expected 'test output 42\\n', got %q", output)
	}
}
