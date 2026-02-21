package modules

import (
	"regexp"
	"strings"
)

// TransformESM converts ES module import/export syntax to CommonJS
// require()/module.exports so that goja (which only supports ES5.1
// scripts) can execute the code.
//
// If the source contains no import/export statements the original
// string is returned unchanged.
func TransformESM(src string) string {
	// Quick check: skip transformation if no ESM keywords present.
	if !hasESMSyntax(src) {
		return src
	}

	lines := strings.Split(src, "\n")
	out := make([]string, 0, len(lines))
	// Counter for generating unique temp variable names for named imports.
	counter := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// ── import handling ───────────────────────────────────────
		if strings.HasPrefix(trimmed, "import ") || strings.HasPrefix(trimmed, "import{") {
			if converted, ok := convertImport(trimmed, &counter); ok {
				out = append(out, converted)
				continue
			}
		}

		// ── export handling ───────────────────────────────────────
		if strings.HasPrefix(trimmed, "export ") {
			if converted, ok := convertExport(trimmed); ok {
				out = append(out, converted)
				continue
			}
		}

		out = append(out, line)
	}

	return strings.Join(out, "\n")
}

// hasESMSyntax does a quick scan for import/export at the start of
// a line (ignoring leading whitespace).
func hasESMSyntax(src string) bool {
	for _, line := range strings.Split(src, "\n") {
		t := strings.TrimSpace(line)
		if strings.HasPrefix(t, "import ") || strings.HasPrefix(t, "import{") ||
			strings.HasPrefix(t, "export ") {
			return true
		}
	}
	return false
}

// ── Import patterns ──────────────────────────────────────────────────

var (
	// import { a, b as c } from "mod"
	reNamedImport = regexp.MustCompile(`^import\s*\{([^}]+)\}\s*from\s*["']([^"']+)["']\s*;?\s*$`)

	// import x from "mod"
	reDefaultImport = regexp.MustCompile(`^import\s+(\w+)\s+from\s*["']([^"']+)["']\s*;?\s*$`)

	// import * as x from "mod"
	reNamespaceImport = regexp.MustCompile(`^import\s+\*\s+as\s+(\w+)\s+from\s*["']([^"']+)["']\s*;?\s*$`)

	// import "mod"  (side-effect only)
	reBareImport = regexp.MustCompile(`^import\s+["']([^"']+)["']\s*;?\s*$`)

	// import x, { a, b } from "mod"  (default + named)
	reDefaultAndNamedImport = regexp.MustCompile(`^import\s+(\w+)\s*,\s*\{([^}]+)\}\s*from\s*["']([^"']+)["']\s*;?\s*$`)
)

func convertImport(line string, counter *int) (string, bool) {
	// import x, { a, b } from "mod"
	if m := reDefaultAndNamedImport.FindStringSubmatch(line); m != nil {
		def := m[1]
		specs := m[2]
		mod := m[3]
		*counter++
		tmp := tempVar(*counter)
		parts := []string{
			"var " + tmp + " = require(\"" + mod + "\");",
			"var " + def + " = " + tmp + ".default !== undefined ? " + tmp + ".default : " + tmp + ";",
		}
		parts = append(parts, namedBindings(specs, tmp)...)
		return strings.Join(parts, " "), true
	}

	// import { a, b as c } from "mod"
	if m := reNamedImport.FindStringSubmatch(line); m != nil {
		specs := m[1]
		mod := m[2]
		*counter++
		tmp := tempVar(*counter)
		parts := []string{"var " + tmp + " = require(\"" + mod + "\");"}
		parts = append(parts, namedBindings(specs, tmp)...)
		return strings.Join(parts, " "), true
	}

	// import * as x from "mod"
	if m := reNamespaceImport.FindStringSubmatch(line); m != nil {
		name := m[1]
		mod := m[2]
		return "var " + name + " = require(\"" + mod + "\");", true
	}

	// import x from "mod"
	if m := reDefaultImport.FindStringSubmatch(line); m != nil {
		name := m[1]
		mod := m[2]
		return "var " + name + " = require(\"" + mod + "\");", true
	}

	// import "mod"
	if m := reBareImport.FindStringSubmatch(line); m != nil {
		mod := m[1]
		return "require(\"" + mod + "\");", true
	}

	return "", false
}

// namedBindings converts "a, b as c, d" into individual var assignments
// from a temp variable.
func namedBindings(specs, tmp string) []string {
	var result []string
	for _, spec := range strings.Split(specs, ",") {
		spec = strings.TrimSpace(spec)
		if spec == "" {
			continue
		}
		parts := strings.Fields(spec)
		if len(parts) == 3 && parts[1] == "as" {
			// { original as alias }
			result = append(result, "var "+parts[2]+" = "+tmp+"."+parts[0]+";")
		} else if len(parts) == 1 {
			// { name }
			result = append(result, "var "+parts[0]+" = "+tmp+"."+parts[0]+";")
		}
	}
	return result
}

func tempVar(n int) string {
	return "_esm$" + itoa(n)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	s := ""
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	return s
}

// ── Export patterns ──────────────────────────────────────────────────

var (
	// export default <expr>
	reExportDefault = regexp.MustCompile(`^export\s+default\s+(.+)$`)

	// export { a, b, c as d }
	reExportNamed = regexp.MustCompile(`^export\s*\{([^}]+)\}\s*;?\s*$`)

	// export const/let/var name = ...
	reExportDecl = regexp.MustCompile(`^export\s+(const|let|var)\s+(\w+)\s*=\s*(.+)$`)

	// export function name(...)  { ... }
	reExportFunc = regexp.MustCompile(`^export\s+(function\s+(\w+).*)$`)

	// export class Name { ... }
	reExportClass = regexp.MustCompile(`^export\s+(class\s+(\w+).*)$`)
)

func convertExport(line string) (string, bool) {
	// export default ...
	if m := reExportDefault.FindStringSubmatch(line); m != nil {
		expr := m[1]
		// Remove trailing semicolons
		expr = strings.TrimRight(expr, "; ")
		return "module.exports = " + expr + ";", true
	}

	// export { a, b, c as d }
	if m := reExportNamed.FindStringSubmatch(line); m != nil {
		specs := m[1]
		var parts []string
		for _, spec := range strings.Split(specs, ",") {
			spec = strings.TrimSpace(spec)
			if spec == "" {
				continue
			}
			fields := strings.Fields(spec)
			if len(fields) == 3 && fields[1] == "as" {
				parts = append(parts, "module.exports."+fields[2]+" = "+fields[0]+";")
			} else if len(fields) == 1 {
				parts = append(parts, "module.exports."+fields[0]+" = "+fields[0]+";")
			}
		}
		return strings.Join(parts, " "), true
	}

	// export const/let/var name = ...
	if m := reExportDecl.FindStringSubmatch(line); m != nil {
		kind := m[1]
		name := m[2]
		value := m[3]
		return kind + " " + name + " = " + value + " module.exports." + name + " = " + name + ";", true
	}

	// export function name(...) { ... }
	if m := reExportFunc.FindStringSubmatch(line); m != nil {
		funcDecl := m[1]
		name := m[2]
		return funcDecl + " module.exports." + name + " = " + name + ";", true
	}

	// export class Name { ... }
	if m := reExportClass.FindStringSubmatch(line); m != nil {
		classDecl := m[1]
		name := m[2]
		return classDecl + " module.exports." + name + " = " + name + ";", true
	}

	return "", false
}
