package modules

import (
	"testing"

	"github.com/robertkrimen/otto"
)

func pathVM(t *testing.T) *otto.Otto {
	t.Helper()
	vm := otto.New()
	NewLoader(vm)
	mustRunP(t, vm, `var path = require('path')`)
	return vm
}

func mustRunP(t *testing.T, vm *otto.Otto, js string) otto.Value {
	t.Helper()
	val, err := vm.Run(js)
	if err != nil {
		t.Fatalf("unexpected error: %v\nscript: %s", err, js)
	}
	return val
}

// ---------------------------------------------------------------------------
// require
// ---------------------------------------------------------------------------

func TestPathRequire(t *testing.T) {
	vm := otto.New()
	NewLoader(vm)
	mustRunP(t, vm, `var p = require('path'); if (!p.join) throw new Error('missing join')`)
}

func TestPathRequireNodePrefix(t *testing.T) {
	vm := otto.New()
	NewLoader(vm)
	mustRunP(t, vm, `var p = require('node:path'); if (!p.join) throw new Error('missing join')`)
}

// ---------------------------------------------------------------------------
// sep / delimiter
// ---------------------------------------------------------------------------

func TestPathSep(t *testing.T) {
	vm := pathVM(t)
	mustRunP(t, vm, `if (path.sep !== '/') throw new Error('sep: ' + path.sep)`)
}

func TestPathDelimiter(t *testing.T) {
	vm := pathVM(t)
	mustRunP(t, vm, `if (path.delimiter !== ':') throw new Error('delimiter: ' + path.delimiter)`)
}

// ---------------------------------------------------------------------------
// basename
// ---------------------------------------------------------------------------

func TestPathBasename(t *testing.T) {
	vm := pathVM(t)
	cases := []struct{ js, want string }{
		{`path.basename('/foo/bar/baz/asdf/quux.html')`, "quux.html"},
		{`path.basename('/foo/bar/baz/asdf/quux.html', '.html')`, "quux"},
		{`path.basename('/foo/bar/')`, "bar"},
		{`path.basename('.')`, "."},
		{`path.basename('/')`, ""},
		{`path.basename('/a')`, "a"},
	}
	for _, tc := range cases {
		t.Run(tc.js, func(t *testing.T) {
			val := mustRunP(t, vm, tc.js)
			if val.String() != tc.want {
				t.Fatalf("expected %q, got %q", tc.want, val.String())
			}
		})
	}
}

// ---------------------------------------------------------------------------
// dirname
// ---------------------------------------------------------------------------

func TestPathDirname(t *testing.T) {
	vm := pathVM(t)
	cases := []struct{ js, want string }{
		{`path.dirname('/foo/bar/baz/asdf/quux')`, "/foo/bar/baz/asdf"},
		{`path.dirname('/foo/bar/')`, "/foo"},
		{`path.dirname('/foo')`, "/"},
		{`path.dirname('foo')`, "."},
		{`path.dirname('.')`, "."},
		{`path.dirname('/')`, "/"},
	}
	for _, tc := range cases {
		t.Run(tc.js, func(t *testing.T) {
			val := mustRunP(t, vm, tc.js)
			if val.String() != tc.want {
				t.Fatalf("expected %q, got %q", tc.want, val.String())
			}
		})
	}
}

// ---------------------------------------------------------------------------
// extname
// ---------------------------------------------------------------------------

func TestPathExtname(t *testing.T) {
	vm := pathVM(t)
	cases := []struct{ js, want string }{
		{`path.extname('index.html')`, ".html"},
		{`path.extname('index.coffee.md')`, ".md"},
		{`path.extname('index.')`, "."},
		{`path.extname('index')`, ""},
		{`path.extname('.index')`, ""},
		{`path.extname('.index.md')`, ".md"},
		{`path.extname('..hidden')`, ".hidden"},
	}
	for _, tc := range cases {
		t.Run(tc.js, func(t *testing.T) {
			val := mustRunP(t, vm, tc.js)
			if val.String() != tc.want {
				t.Fatalf("expected %q, got %q", tc.want, val.String())
			}
		})
	}
}

// ---------------------------------------------------------------------------
// join
// ---------------------------------------------------------------------------

func TestPathJoin(t *testing.T) {
	vm := pathVM(t)
	cases := []struct{ js, want string }{
		{`path.join('/foo', 'bar', 'baz/asdf', 'quux', '..')`, "/foo/bar/baz/asdf"},
		{`path.join('foo', 'bar', 'baz')`, "foo/bar/baz"},
		{`path.join('foo', '', 'bar')`, "foo/bar"},
		{`path.join('.')`, "."},
		{`path.join('/', 'foo')`, "/foo"},
		{`path.join('/foo', '/bar')`, "/foo/bar"},
	}
	for _, tc := range cases {
		t.Run(tc.js, func(t *testing.T) {
			val := mustRunP(t, vm, tc.js)
			if val.String() != tc.want {
				t.Fatalf("expected %q, got %q", tc.want, val.String())
			}
		})
	}
}

// ---------------------------------------------------------------------------
// normalize
// ---------------------------------------------------------------------------

func TestPathNormalize(t *testing.T) {
	vm := pathVM(t)
	cases := []struct{ js, want string }{
		{`path.normalize('/foo/bar//baz/asdf/quux/..')`, "/foo/bar/baz/asdf"},
		{`path.normalize('.')`, "."},
		{`path.normalize('./')`, "./"},
		{`path.normalize('/')`, "/"},
		{`path.normalize('/foo/../bar')`, "/bar"},
		{`path.normalize('foo/bar/../baz')`, "foo/baz"},
	}
	for _, tc := range cases {
		t.Run(tc.js, func(t *testing.T) {
			val := mustRunP(t, vm, tc.js)
			if val.String() != tc.want {
				t.Fatalf("expected %q, got %q", tc.want, val.String())
			}
		})
	}
}

// ---------------------------------------------------------------------------
// isAbsolute
// ---------------------------------------------------------------------------

func TestPathIsAbsolute(t *testing.T) {
	vm := pathVM(t)
	mustRunP(t, vm, `
		if (!path.isAbsolute('/foo/bar')) throw new Error('should be absolute');
		if (!path.isAbsolute('/')) throw new Error('root should be absolute');
		if (path.isAbsolute('foo/bar')) throw new Error('relative should not be absolute');
		if (path.isAbsolute('.')) throw new Error('dot not absolute');
	`)
}

// ---------------------------------------------------------------------------
// resolve
// ---------------------------------------------------------------------------

func TestPathResolve(t *testing.T) {
	vm := pathVM(t)
	mustRunP(t, vm, `
		var r = path.resolve('/foo/bar', './baz');
		if (r !== '/foo/bar/baz') throw new Error('resolve 1: ' + r);
		r = path.resolve('/foo/bar', '/tmp/file');
		if (r !== '/tmp/file') throw new Error('resolve 2: ' + r);
	`)
}

// ---------------------------------------------------------------------------
// relative
// ---------------------------------------------------------------------------

func TestPathRelative(t *testing.T) {
	vm := pathVM(t)
	mustRunP(t, vm, `
		var r = path.relative('/data/orandea/test/aaa', '/data/orandea/impl/bbb');
		if (r !== '../../impl/bbb') throw new Error('relative: ' + r);
	`)
}

func TestPathRelative_Same(t *testing.T) {
	vm := pathVM(t)
	mustRunP(t, vm, `
		if (path.relative('/foo', '/foo') !== '') throw new Error('same paths should be empty');
	`)
}

// ---------------------------------------------------------------------------
// parse
// ---------------------------------------------------------------------------

func TestPathParse(t *testing.T) {
	vm := pathVM(t)
	mustRunP(t, vm, `
		var p = path.parse('/home/user/dir/file.txt');
		if (p.root !== '/') throw new Error('root: ' + p.root);
		if (p.dir !== '/home/user/dir') throw new Error('dir: ' + p.dir);
		if (p.base !== 'file.txt') throw new Error('base: ' + p.base);
		if (p.ext !== '.txt') throw new Error('ext: ' + p.ext);
		if (p.name !== 'file') throw new Error('name: ' + p.name);
	`)
}

func TestPathParse_Root(t *testing.T) {
	vm := pathVM(t)
	mustRunP(t, vm, `
		var p = path.parse('/');
		if (p.root !== '/') throw new Error('root: ' + p.root);
		if (p.dir !== '/') throw new Error('dir: ' + p.dir);
		if (p.base !== '') throw new Error('base: ' + p.base);
	`)
}

func TestPathParse_Relative(t *testing.T) {
	vm := pathVM(t)
	mustRunP(t, vm, `
		var p = path.parse('foo.txt');
		if (p.root !== '') throw new Error('root: ' + p.root);
		if (p.base !== 'foo.txt') throw new Error('base: ' + p.base);
		if (p.name !== 'foo') throw new Error('name: ' + p.name);
		if (p.ext !== '.txt') throw new Error('ext: ' + p.ext);
	`)
}

// ---------------------------------------------------------------------------
// format
// ---------------------------------------------------------------------------

func TestPathFormat(t *testing.T) {
	vm := pathVM(t)
	mustRunP(t, vm, `
		var f = path.format({ root: '/', dir: '/home/user/dir', base: 'file.txt' });
		if (f !== '/home/user/dir/file.txt') throw new Error('format: ' + f);
	`)
}

func TestPathFormat_NameExt(t *testing.T) {
	vm := pathVM(t)
	mustRunP(t, vm, `
		var f = path.format({ root: '/', name: 'file', ext: '.txt' });
		if (f !== '/file.txt') throw new Error('format: ' + f);
	`)
}

// ---------------------------------------------------------------------------
// toNamespacedPath
// ---------------------------------------------------------------------------

func TestPathToNamespacedPath(t *testing.T) {
	vm := pathVM(t)
	mustRunP(t, vm, `
		if (path.toNamespacedPath('/foo/bar') !== '/foo/bar') throw new Error('should be identity');
	`)
}

// ---------------------------------------------------------------------------
// posix self-reference
// ---------------------------------------------------------------------------

func TestPathPosixRef(t *testing.T) {
	vm := pathVM(t)
	mustRunP(t, vm, `
		if (path.posix !== path) throw new Error('path.posix should reference self');
	`)
}

// ---------------------------------------------------------------------------
// Edge cases
// ---------------------------------------------------------------------------

func TestPathJoinEmpty(t *testing.T) {
	vm := pathVM(t)
	mustRunP(t, vm, `
		if (path.join() !== '.') throw new Error('empty join');
	`)
}

func TestPathNormalizeDoubleDot(t *testing.T) {
	vm := pathVM(t)
	mustRunP(t, vm, `
		if (path.normalize('foo/../../bar') !== '../bar') throw new Error('double dot: ' + path.normalize('foo/../../bar'));
	`)
}
