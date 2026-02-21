// Package modules provides a require()-based module loading system
// for the gode JavaScript runtime. Built-in modules are embedded at
// compile time and evaluated lazily on first require(). User modules
// are resolved following Node.js conventions: relative paths, then
// node_modules directory walking.
package modules

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gode/compat"

	"github.com/robertkrimen/otto"
)

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
	vm       *otto.Otto
	registry map[string]string     // name → JS source (built-ins)
	cache    map[string]otto.Value // resolved path → cached exports
}

// NewLoader creates a Loader, registers all built-in modules, and
// installs require() on vm.
func NewLoader(vm *otto.Otto) *Loader {
	l := &Loader{
		vm:       vm,
		registry: make(map[string]string),
		cache:    make(map[string]otto.Value),
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

func (l *Loader) require(call otto.FunctionCall) otto.Value {
	name := call.Argument(0).String()

	// 1. Built-in modules (check cache, then registry).
	if val, ok := l.cache[name]; ok {
		return val
	}
	if src, ok := l.registry[name]; ok {
		val, err := l.vm.Run(src)
		if err != nil {
			panic(l.vm.MakeCustomError("Error", fmt.Sprintf("Error loading module '%s': %v", name, err)))
		}
		l.cache[name] = val
		return val
	}

	// 2. File-based resolution.
	baseDir := l.callerDir()
	resolved := l.resolveModule(name, baseDir)
	if resolved == "" {
		panic(l.vm.MakeCustomError("Error", fmt.Sprintf("Cannot find module '%s'", name)))
	}

	// Check file cache by absolute path.
	if val, ok := l.cache[resolved]; ok {
		return val
	}

	// 3. Load, wrap, and execute.
	src, err := os.ReadFile(resolved)
	if err != nil {
		panic(l.vm.MakeCustomError("Error", fmt.Sprintf("Cannot read module '%s': %v", resolved, err)))
	}

	dir := filepath.Dir(resolved)

	// Create a per-module require that resolves relative to this module's dir.
	moduleRequire := func(call otto.FunctionCall) otto.Value {
		modName := call.Argument(0).String()

		// Built-in modules first.
		if val, ok := l.cache[modName]; ok {
			return val
		}
		if src, ok := l.registry[modName]; ok {
			if val, ok := l.cache[modName]; ok {
				return val
			}
			val, err := l.vm.Run(src)
			if err != nil {
				panic(l.vm.MakeCustomError("Error", fmt.Sprintf("Error loading module '%s': %v", modName, err)))
			}
			l.cache[modName] = val
			return val
		}

		// File-based resolution from this module's directory.
		resolved := l.resolveModule(modName, dir)
		if resolved == "" {
			panic(l.vm.MakeCustomError("Error", fmt.Sprintf("Cannot find module '%s'", modName)))
		}
		if val, ok := l.cache[resolved]; ok {
			return val
		}
		return l.loadFileModule(modName, resolved)
	}

	// Set up module object.
	moduleObj, _ := l.vm.Object(`({exports: {}})`)
	exportsVal, _ := moduleObj.Get("exports")

	// Save and restore all module-scoped globals.
	oldRequire, _ := l.vm.Get("require")
	oldFilename, _ := l.vm.Get("__filename")
	oldDirname, _ := l.vm.Get("__dirname")
	oldModule, _ := l.vm.Get("module")
	oldExports, _ := l.vm.Get("exports")

	l.vm.Set("require", moduleRequire)
	l.vm.Set("__filename", resolved)
	l.vm.Set("__dirname", dir)
	l.vm.Set("module", moduleObj)
	l.vm.Set("exports", exportsVal)

	_, err = l.vm.Run(compat.Transform(string(src)))

	// Capture module.exports BEFORE restoring globals.
	result, _ := l.vm.Run(`module.exports`)

	// Restore all globals.
	l.vm.Set("require", oldRequire)
	l.vm.Set("__filename", oldFilename)
	l.vm.Set("__dirname", oldDirname)
	l.vm.Set("module", oldModule)
	l.vm.Set("exports", oldExports)

	if err != nil {
		panic(l.vm.MakeCustomError("Error", fmt.Sprintf("Error loading module '%s': %v", name, err)))
	}

	l.cache[resolved] = result
	return result
}

// loadFileModule reads, wraps, and executes a file module.
// This is the shared implementation used by both the top-level and
// per-module require functions.
func (l *Loader) loadFileModule(name, resolved string) otto.Value {
	src, err := os.ReadFile(resolved)
	if err != nil {
		panic(l.vm.MakeCustomError("Error", fmt.Sprintf("Cannot read module '%s': %v", resolved, err)))
	}

	dir := filepath.Dir(resolved)

	// Create a per-module require for this module.
	moduleRequire := func(call otto.FunctionCall) otto.Value {
		modName := call.Argument(0).String()

		if val, ok := l.cache[modName]; ok {
			return val
		}
		if src, ok := l.registry[modName]; ok {
			if val, ok := l.cache[modName]; ok {
				return val
			}
			val, err := l.vm.Run(src)
			if err != nil {
				panic(l.vm.MakeCustomError("Error", fmt.Sprintf("Error loading module '%s': %v", modName, err)))
			}
			l.cache[modName] = val
			return val
		}

		resolved := l.resolveModule(modName, dir)
		if resolved == "" {
			panic(l.vm.MakeCustomError("Error", fmt.Sprintf("Cannot find module '%s'", modName)))
		}
		if val, ok := l.cache[resolved]; ok {
			return val
		}
		return l.loadFileModule(modName, resolved)
	}

	moduleObj, _ := l.vm.Object(`({exports: {}})`)
	exportsVal, _ := moduleObj.Get("exports")

	oldRequire, _ := l.vm.Get("require")
	oldFilename, _ := l.vm.Get("__filename")
	oldDirname, _ := l.vm.Get("__dirname")
	oldModule, _ := l.vm.Get("module")
	oldExports, _ := l.vm.Get("exports")

	l.vm.Set("require", moduleRequire)
	l.vm.Set("__filename", resolved)
	l.vm.Set("__dirname", dir)
	l.vm.Set("module", moduleObj)
	l.vm.Set("exports", exportsVal)

	_, err = l.vm.Run(compat.Transform(string(src)))

	// Capture module.exports BEFORE restoring globals.
	result, _ := l.vm.Run(`module.exports`)

	l.vm.Set("require", oldRequire)
	l.vm.Set("__filename", oldFilename)
	l.vm.Set("__dirname", oldDirname)
	l.vm.Set("module", oldModule)
	l.vm.Set("exports", oldExports)

	if err != nil {
		panic(l.vm.MakeCustomError("Error", fmt.Sprintf("Error loading module '%s': %v", name, err)))
	}

	l.cache[resolved] = result
	return result
}

// callerDir returns the base directory for module resolution.
// Uses __dirname if set (inside a module), otherwise process.cwd().
func (l *Loader) callerDir() string {
	// Check __dirname first (set when inside a loaded module/script).
	if val, err := l.vm.Get("__dirname"); err == nil && val.IsString() {
		s, _ := val.ToString()
		if s != "" && s != "undefined" {
			return s
		}
	}
	// Fall back to process.cwd().
	val, err := l.vm.Run(`process.cwd()`)
	if err != nil {
		cwd, _ := os.Getwd()
		return cwd
	}
	return val.String()
}

// resolveModule implements Node.js-style module resolution.
func (l *Loader) resolveModule(name, baseDir string) string {
	// Relative or absolute paths.
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

	// node_modules resolution: walk up from baseDir.
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
			break // reached filesystem root
		}
		dir = parent
	}
	return ""
}

// resolveFile tries to load path as a file: exact, .js, .json.
func (l *Loader) resolveFile(path string) string {
	candidates := []string{path, path + ".js", path + ".json"}
	for _, c := range candidates {
		if info, err := os.Stat(c); err == nil && !info.IsDir() {
			return c
		}
	}
	return ""
}

// resolveDir tries to load path as a directory: package.json "main", index.js.
func (l *Loader) resolveDir(path string) string {
	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		return ""
	}

	// Try package.json "main" field.
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
			// main might point to a directory itself.
			if r := l.resolveDir(main); r != "" {
				return r
			}
		}
	}

	// Try index.js.
	idx := filepath.Join(path, "index.js")
	if info, err := os.Stat(idx); err == nil && !info.IsDir() {
		return idx
	}

	return ""
}
