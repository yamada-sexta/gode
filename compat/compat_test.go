package compat

import (
	"testing"

	"github.com/robertkrimen/otto"
)

func TestTransform_ConstToVar(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"basic const", "const x = 1", "var x = 1"},
		{"basic let", "let x = 1", "var x = 1"},
		{"multiple", "const a = 1; let b = 2;", "var a = 1; var b = 2;"},
		{"indented", "  const x = 1", "  var x = 1"},
		{"in for loop", "for (let i = 0; i < 10; i++) {}", "for (var i = 0; i < 10; i++) {}"},
		{"var unchanged", "var x = 1", "var x = 1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Transform(tt.in)
			if got != tt.want {
				t.Fatalf("Transform(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestTransform_NoFalsePositives(t *testing.T) {
	tests := []struct {
		name string
		in   string
	}{
		{"constant identifier", "var constantly = 1"},
		{"letter identifier", "var letter = 'a'"},
		{"inside method name", "obj.constructor()"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Transform(tt.in)
			if got != tt.in {
				t.Fatalf("Transform(%q) should not change, got %q", tt.in, got)
			}
		})
	}
}

func TestTransform_RunsInOtto(t *testing.T) {
	vm := otto.New()

	tests := []struct {
		name   string
		code   string
		check  string
		expect string
	}{
		{"const declaration", "const x = 42;", "x", "42"},
		{"let declaration", "let y = 'hello';", "y", "hello"},
		{"const in block", "{ const z = true; }", "z", "true"},
		{"let in for", "var sum = 0; for (let i = 1; i <= 3; i++) { sum += i; }", "sum", "6"},
		{"const object", "const obj = {a: 1, b: 2};", "JSON.stringify(obj)", `{"a":1,"b":2}`},
		{"const array", "const arr = [1, 2, 3];", "arr.length", "3"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := vm.Run(Transform(tt.code))
			if err != nil {
				t.Fatalf("Run error: %v\ncode: %s\ntransformed: %s", err, tt.code, Transform(tt.code))
			}
			val, err := vm.Run(tt.check)
			if err != nil {
				t.Fatalf("Check error: %v", err)
			}
			if val.String() != tt.expect {
				t.Fatalf("expected %s = %q, got %q", tt.check, tt.expect, val.String())
			}
		})
	}
}
