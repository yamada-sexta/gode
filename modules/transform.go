package modules

import (
	"strings"

	"github.com/evanw/esbuild/pkg/api"
)

// TransformESM uses esbuild to transpile modern JavaScript (ESM
// import/export, for-await-of, top-level await, etc.) into CommonJS
// that goja can execute.
//
// If the source contains no ESM syntax, it is still run through
// esbuild to down-level any unsupported syntax (e.g. for-await-of).
func TransformESM(src string) string {
	result := api.Transform(src, api.TransformOptions{
		Format:        api.FormatCommonJS,
		Loader:        api.LoaderJS,
		Target:        api.ES2020,
		Platform:      api.PlatformNode,
		Sourcemap:     api.SourceMapNone,
		LogLevel:      api.LogLevelSilent,
		Charset:       api.CharsetUTF8,
		LegalComments: api.LegalCommentsNone,
	})

	if len(result.Errors) > 0 {
		// If esbuild fails to parse, return original source and let
		// goja report the error with proper line numbers.
		return src
	}

	out := string(result.Code)
	// esbuild wraps the output in its own CJS shim. Trim any
	// trailing newline for cleaner output.
	out = strings.TrimRight(out, "\n")
	return out
}
