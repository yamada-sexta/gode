package envfile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_BasicKeyValue(t *testing.T) {
	f := writeTemp(t, "KEY=value\n")
	if err := Load(f); err != nil {
		t.Fatal(err)
	}
	assertEnv(t, "KEY", "value")
}

func TestLoad_QuotedValues(t *testing.T) {
	f := writeTemp(t, `
DOUBLE="hello world"
SINGLE='foo bar'
`)
	if err := Load(f); err != nil {
		t.Fatal(err)
	}
	assertEnv(t, "DOUBLE", "hello world")
	assertEnv(t, "SINGLE", "foo bar")
}

func TestLoad_Comments(t *testing.T) {
	f := writeTemp(t, `
# This is a comment
KEY=value
# Another comment
`)
	if err := Load(f); err != nil {
		t.Fatal(err)
	}
	assertEnv(t, "KEY", "value")
}

func TestLoad_EmptyLines(t *testing.T) {
	f := writeTemp(t, `
A=1

B=2

`)
	if err := Load(f); err != nil {
		t.Fatal(err)
	}
	assertEnv(t, "A", "1")
	assertEnv(t, "B", "2")
}

func TestLoad_ValueWithEquals(t *testing.T) {
	f := writeTemp(t, "URL=https://example.com?a=1&b=2\n")
	if err := Load(f); err != nil {
		t.Fatal(err)
	}
	assertEnv(t, "URL", "https://example.com?a=1&b=2")
}

func TestLoad_TrimWhitespace(t *testing.T) {
	f := writeTemp(t, "  KEY  =  value  \n")
	if err := Load(f); err != nil {
		t.Fatal(err)
	}
	assertEnv(t, "KEY", "value")
}

func TestLoad_MissingFile(t *testing.T) {
	err := Load("/tmp/nonexistent_env_file_12345")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoad_InvalidLine(t *testing.T) {
	f := writeTemp(t, "NOEQUALS\n")
	err := Load(f)
	if err == nil {
		t.Fatal("expected error for invalid line")
	}
}

func TestLoadIfExists_FileExists(t *testing.T) {
	f := writeTemp(t, "X=42\n")
	if err := LoadIfExists(f); err != nil {
		t.Fatal(err)
	}
	assertEnv(t, "X", "42")
}

func TestLoadIfExists_FileMissing(t *testing.T) {
	err := LoadIfExists("/tmp/nonexistent_env_file_67890")
	if err != nil {
		t.Fatalf("expected no error for missing file, got: %v", err)
	}
}

// --- Helpers ---

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func assertEnv(t *testing.T, key, want string) {
	t.Helper()
	got := os.Getenv(key)
	if got != want {
		t.Fatalf("env %s: expected %q, got %q", key, want, got)
	}
}
