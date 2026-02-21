package main

import (
	"reflect"
	"testing"
)

func TestParseArgs_Version(t *testing.T) {
	for _, flag := range []string{"-v", "--version"} {
		t.Run(flag, func(t *testing.T) {
			opts, err := parseArgs([]string{flag})
			if err != nil {
				t.Fatal(err)
			}
			if !opts.Version {
				t.Fatal("expected Version to be true")
			}
		})
	}
}

func TestParseArgs_Help(t *testing.T) {
	for _, flag := range []string{"-h", "--help"} {
		t.Run(flag, func(t *testing.T) {
			opts, err := parseArgs([]string{flag})
			if err != nil {
				t.Fatal(err)
			}
			if !opts.Help {
				t.Fatal("expected Help to be true")
			}
		})
	}
}

func TestParseArgs_Interactive(t *testing.T) {
	for _, flag := range []string{"-i", "--interactive"} {
		t.Run(flag, func(t *testing.T) {
			opts, err := parseArgs([]string{flag})
			if err != nil {
				t.Fatal(err)
			}
			if !opts.Interactive {
				t.Fatal("expected Interactive to be true")
			}
		})
	}
}

func TestParseArgs_Check(t *testing.T) {
	for _, flag := range []string{"-c", "--check"} {
		t.Run(flag, func(t *testing.T) {
			opts, err := parseArgs([]string{flag, "file.js"})
			if err != nil {
				t.Fatal(err)
			}
			if !opts.Check {
				t.Fatal("expected Check to be true")
			}
			if opts.Script != "file.js" {
				t.Fatalf("expected script 'file.js', got %q", opts.Script)
			}
		})
	}
}

func TestParseArgs_Eval(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"-e", []string{"-e", "console.log(1)"}, "console.log(1)"},
		{"--eval", []string{"--eval", "1+1"}, "1+1"},
		{"--eval=", []string{"--eval=1+1"}, "1+1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts, err := parseArgs(tt.args)
			if err != nil {
				t.Fatal(err)
			}
			if opts.Eval != tt.want {
				t.Fatalf("expected Eval %q, got %q", tt.want, opts.Eval)
			}
		})
	}
}

func TestParseArgs_EvalMissingArg(t *testing.T) {
	_, err := parseArgs([]string{"-e"})
	if err == nil {
		t.Fatal("expected error for missing -e argument")
	}
}

func TestParseArgs_Print(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"-p", []string{"-p", "1+1"}, "1+1"},
		{"--print", []string{"--print", "1+1"}, "1+1"},
		{"--print=", []string{"--print=1+1"}, "1+1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts, err := parseArgs(tt.args)
			if err != nil {
				t.Fatal(err)
			}
			if opts.Print != tt.want {
				t.Fatalf("expected Print %q, got %q", tt.want, opts.Print)
			}
		})
	}
}

func TestParseArgs_PrintMissingArg(t *testing.T) {
	_, err := parseArgs([]string{"-p"})
	if err == nil {
		t.Fatal("expected error for missing -p argument")
	}
}

func TestParseArgs_EnvFile(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want []string
	}{
		{"--env-file", []string{"--env-file", ".env"}, []string{".env"}},
		{"--env-file=", []string{"--env-file=.env"}, []string{".env"}},
		{"multiple", []string{"--env-file", "a.env", "--env-file", "b.env"}, []string{"a.env", "b.env"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts, err := parseArgs(tt.args)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(opts.EnvFiles, tt.want) {
				t.Fatalf("expected EnvFiles %v, got %v", tt.want, opts.EnvFiles)
			}
		})
	}
}

func TestParseArgs_EnvFileIfExists(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want []string
	}{
		{"--env-file-if-exists", []string{"--env-file-if-exists", ".env"}, []string{".env"}},
		{"--env-file-if-exists=", []string{"--env-file-if-exists=.env"}, []string{".env"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts, err := parseArgs(tt.args)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(opts.EnvFilesIfExist, tt.want) {
				t.Fatalf("expected EnvFilesIfExist %v, got %v", tt.want, opts.EnvFilesIfExist)
			}
		})
	}
}

func TestParseArgs_Stdin(t *testing.T) {
	opts, err := parseArgs([]string{"-"})
	if err != nil {
		t.Fatal(err)
	}
	if !opts.ReadStdin {
		t.Fatal("expected ReadStdin to be true")
	}
}

func TestParseArgs_StdinWithArgs(t *testing.T) {
	opts, err := parseArgs([]string{"-", "arg1", "arg2"})
	if err != nil {
		t.Fatal(err)
	}
	if !opts.ReadStdin {
		t.Fatal("expected ReadStdin to be true")
	}
	want := []string{"arg1", "arg2"}
	if !reflect.DeepEqual(opts.ScriptArgs, want) {
		t.Fatalf("expected ScriptArgs %v, got %v", want, opts.ScriptArgs)
	}
}

func TestParseArgs_Script(t *testing.T) {
	opts, err := parseArgs([]string{"app.js"})
	if err != nil {
		t.Fatal(err)
	}
	if opts.Script != "app.js" {
		t.Fatalf("expected script 'app.js', got %q", opts.Script)
	}
}

func TestParseArgs_ScriptWithArgs(t *testing.T) {
	opts, err := parseArgs([]string{"app.js", "arg1", "--flag", "arg2"})
	if err != nil {
		t.Fatal(err)
	}
	if opts.Script != "app.js" {
		t.Fatalf("expected script 'app.js', got %q", opts.Script)
	}
	want := []string{"arg1", "--flag", "arg2"}
	if !reflect.DeepEqual(opts.ScriptArgs, want) {
		t.Fatalf("expected ScriptArgs %v, got %v", want, opts.ScriptArgs)
	}
}

func TestParseArgs_DoubleDash(t *testing.T) {
	opts, err := parseArgs([]string{"--", "script.js", "a", "b"})
	if err != nil {
		t.Fatal(err)
	}
	if opts.Script != "script.js" {
		t.Fatalf("expected script 'script.js', got %q", opts.Script)
	}
	want := []string{"a", "b"}
	if !reflect.DeepEqual(opts.ScriptArgs, want) {
		t.Fatalf("expected ScriptArgs %v, got %v", want, opts.ScriptArgs)
	}
}

func TestParseArgs_FlagsBeforeScript(t *testing.T) {
	opts, err := parseArgs([]string{"-i", "--env-file", ".env", "app.js", "arg1"})
	if err != nil {
		t.Fatal(err)
	}
	if !opts.Interactive {
		t.Fatal("expected Interactive")
	}
	if !reflect.DeepEqual(opts.EnvFiles, []string{".env"}) {
		t.Fatalf("expected EnvFiles ['.env'], got %v", opts.EnvFiles)
	}
	if opts.Script != "app.js" {
		t.Fatalf("expected script 'app.js', got %q", opts.Script)
	}
	if !reflect.DeepEqual(opts.ScriptArgs, []string{"arg1"}) {
		t.Fatalf("expected ScriptArgs ['arg1'], got %v", opts.ScriptArgs)
	}
}

func TestParseArgs_BadOption(t *testing.T) {
	_, err := parseArgs([]string{"--unknown-flag"})
	if err == nil {
		t.Fatal("expected error for unknown flag")
	}
}

func TestParseArgs_Empty(t *testing.T) {
	opts, err := parseArgs([]string{})
	if err != nil {
		t.Fatal(err)
	}
	// No flags set, no script — should default to REPL.
	if opts.Script != "" || opts.Eval != "" || opts.Version || opts.Help {
		t.Fatal("expected empty options for no args")
	}
}

func TestGetVersion(t *testing.T) {
	v := getVersion()
	if v == "" {
		t.Fatal("version should not be empty")
	}
}
