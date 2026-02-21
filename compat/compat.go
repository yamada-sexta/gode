// Package compat provides source-level transformations that bridge
// common ES6+ syntax to ES5, which is what the otto engine supports.
package compat

import "regexp"

// constLetRe matches `const` or `let` used as declaration keywords.
// Uses word boundaries to avoid replacing inside identifiers.
var constLetRe = regexp.MustCompile(`\b(const|let)\s+`)

// Transform applies ES6→ES5 compatibility transformations to src.
// Currently handles:
//   - const/let → var
func Transform(src string) string {
	return constLetRe.ReplaceAllString(src, "var ")
}
