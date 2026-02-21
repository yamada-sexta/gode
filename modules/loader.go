// Package modules provides a require()-based module loading system
// for the gode JavaScript runtime. Built-in modules are embedded at
// compile time and evaluated lazily on first require(). User modules
// are resolved following Node.js conventions.
package modules

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dop251/goja"
	"github.com/dop251/goja/parser"
)

// sourceMapLoader returns a parser.Option that resolves source map
// paths relative to baseDir. If the .map file does not exist the
// error is silently ignored so that packages shipping without maps
// still load correctly.
func sourceMapLoader(baseDir string) parser.Option {
	return parser.WithSourceMapLoader(func(path string) ([]byte, error) {
		if !filepath.IsAbs(path) {
			path = filepath.Join(baseDir, path)
		}
		data, err := os.ReadFile(path)
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return data, err
	})
}

//go:embed assert.js
var assertSource string

//go:embed buffer.js
var bufferSource string

//go:embed path.js
var pathSource string

//go:embed os.js
var osSource string

//go:embed zlib.js
var zlibSource string

// Loader manages built-in module registration, caching, and the
// require() function installed on the VM.
type Loader struct {
	vm       *goja.Runtime
	registry map[string]string     // name → JS source (built-ins)
	cache    map[string]goja.Value // resolved path → cached exports
}

// NewLoader creates a Loader, registers all built-in modules, and
// installs require() on vm.
func NewLoader(vm *goja.Runtime) *Loader {
	l := &Loader{
		vm:       vm,
		registry: make(map[string]string),
		cache:    make(map[string]goja.Value),
	}

	// Install native helpers needed by JS modules.
	setupOSNative(vm)
	setupZlibNative(vm)

	// Built-in modules.
	l.Register("assert", assertSource)
	l.Register("node:assert", assertSource)
	l.Register("buffer", bufferSource)
	l.Register("node:buffer", bufferSource)
	l.Register("path", pathSource)
	l.Register("node:path", pathSource)
	l.Register("os", osSource)
	l.Register("node:os", osSource)
	l.Register("zlib", zlibSource)
	l.Register("node:zlib", zlibSource)

	// Install require().
	vm.Set("require", l.require)

	return l
}

// Register adds a built-in module.
func (l *Loader) Register(name, source string) {
	l.registry[name] = source
}

func (l *Loader) throwError(msg string) {
	o, _ := l.vm.RunString(fmt.Sprintf(`new Error(%q)`, msg))
	panic(o)
}

func (l *Loader) require(call goja.FunctionCall) goja.Value {
	name := call.Argument(0).String()

	// 1. Built-in modules.
	if val, ok := l.cache[name]; ok {
		return val
	}
	if src, ok := l.registry[name]; ok {
		val, err := l.vm.RunString(src)
		if err != nil {
			l.throwError(fmt.Sprintf("Error loading module '%s': %v", name, err))
		}
		l.cache[name] = val
		return val
	}

	// 2. File-based resolution.
	baseDir := l.callerDir()
	resolved := l.resolveModule(name, baseDir)
	if resolved == "" {
		l.throwError(fmt.Sprintf("Cannot find module '%s'", name))
	}

	if val, ok := l.cache[resolved]; ok {
		return val
	}

	return l.loadFileModule(name, resolved)
}

// loadFileModule reads, wraps, and executes a file module.
func (l *Loader) loadFileModule(name, resolved string) goja.Value {
	src, err := os.ReadFile(resolved)
	if err != nil {
		l.throwError(fmt.Sprintf("Cannot read module '%s': %v", resolved, err))
	}

	dir := filepath.Dir(resolved)

	// Create a per-module require that resolves relative to this dir.
	moduleRequire := func(call goja.FunctionCall) goja.Value {
		modName := call.Argument(0).String()

		if val, ok := l.cache[modName]; ok {
			return val
		}
		if src, ok := l.registry[modName]; ok {
			val, err := l.vm.RunString(src)
			if err != nil {
				l.throwError(fmt.Sprintf("Error loading module '%s': %v", modName, err))
			}
			l.cache[modName] = val
			return val
		}

		resolved := l.resolveModule(modName, dir)
		if resolved == "" {
			l.throwError(fmt.Sprintf("Cannot find module '%s'", modName))
		}
		if val, ok := l.cache[resolved]; ok {
			return val
		}
		return l.loadFileModule(modName, resolved)
	}

	// Save current module-scoped globals.
	oldRequire := l.vm.Get("require")
	oldFilename := l.vm.Get("__filename")
	oldDirname := l.vm.Get("__dirname")
	oldModule := l.vm.Get("module")
	oldExports := l.vm.Get("exports")

	// Set up module context.
	moduleObj := l.vm.NewObject()
	exportsObj := l.vm.NewObject()
	moduleObj.Set("exports", exportsObj)

	l.vm.Set("require", moduleRequire)
	l.vm.Set("__filename", resolved)
	l.vm.Set("__dirname", dir)
	l.vm.Set("module", moduleObj)
	l.vm.Set("exports", exportsObj)

	ast, parseErr := goja.Parse(resolved, string(src), sourceMapLoader(dir))
	if parseErr != nil {
		err = parseErr
	} else {
		prg, compileErr := goja.CompileAST(ast, false)
		if compileErr != nil {
			err = compileErr
		} else {
			_, err = l.vm.RunProgram(prg)
		}
	}

	// Capture module.exports BEFORE restoring globals.
	result, _ := l.vm.RunString(`module.exports`)

	// Restore globals.
	l.vm.Set("require", oldRequire)
	l.vm.Set("__filename", oldFilename)
	l.vm.Set("__dirname", oldDirname)
	l.vm.Set("module", oldModule)
	l.vm.Set("exports", oldExports)

	if err != nil {
		l.throwError(fmt.Sprintf("Error loading module '%s': %v", name, err))
	}

	l.cache[resolved] = result
	return result
}

// callerDir returns the base directory for module resolution.
func (l *Loader) callerDir() string {
	v := l.vm.Get("__dirname")
	if v != nil && !goja.IsUndefined(v) && !goja.IsNull(v) {
		s := v.String()
		if s != "" && s != "undefined" {
			return s
		}
	}
	val, err := l.vm.RunString(`process.cwd()`)
	if err != nil {
		cwd, _ := os.Getwd()
		return cwd
	}
	return val.String()
}

// resolveModule implements Node.js-style module resolution.
func (l *Loader) resolveModule(name, baseDir string) string {
	if strings.HasPrefix(name, "./") || strings.HasPrefix(name, "../") || strings.HasPrefix(name, "/") {
		var abs string
		if filepath.IsAbs(name) {
			abs = name
		} else {
			abs = filepath.Join(baseDir, name)
		}
		if r := l.resolveFile(abs); r != "" {
			return r
		}
		if r := l.resolveDir(abs); r != "" {
			return r
		}
		return ""
	}

	dir := baseDir
	for {
		candidate := filepath.Join(dir, "node_modules", name)
		if r := l.resolveFile(candidate); r != "" {
			return r
		}
		if r := l.resolveDir(candidate); r != "" {
			return r
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

func (l *Loader) resolveFile(path string) string {
	candidates := []string{path, path + ".js", path + ".json"}
	for _, c := range candidates {
		if info, err := os.Stat(c); err == nil && !info.IsDir() {
			return c
		}
	}
	return ""
}

func (l *Loader) resolveDir(path string) string {
	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		return ""
	}

	pkgPath := filepath.Join(path, "package.json")
	if data, err := os.ReadFile(pkgPath); err == nil {
		var pkg struct {
			Main string `json:"main"`
		}
		if json.Unmarshal(data, &pkg) == nil && pkg.Main != "" {
			main := filepath.Join(path, pkg.Main)
			if r := l.resolveFile(main); r != "" {
				return r
			}
			if r := l.resolveDir(main); r != "" {
				return r
			}
		}
	}

	idx := filepath.Join(path, "index.js")
	if info, err := os.Stat(idx); err == nil && !info.IsDir() {
		return idx
	}
	return ""
}
