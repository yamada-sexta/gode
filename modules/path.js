(function () {
  'use strict';

  var CHAR_FORWARD_SLASH = 47; // /
  var CHAR_DOT = 46;          // .

  function isPosixSep(code) { return code === CHAR_FORWARD_SLASH; }

  function assertString(value, name) {
    if (typeof value !== 'string')
      throw new TypeError('The "' + name + '" argument must be of type string. Received ' + typeof value);
  }

  // Resolves . and .. elements in a path
  function normalizeString(path, allowAboveRoot) {
    var res = '';
    var lastSegLen = 0;
    var lastSlash = -1;
    var dots = 0;
    var code;
    for (var i = 0; i <= path.length; ++i) {
      if (i < path.length)
        code = path.charCodeAt(i);
      else if (code === CHAR_FORWARD_SLASH)
        break;
      else
        code = CHAR_FORWARD_SLASH;

      if (code === CHAR_FORWARD_SLASH) {
        if (lastSlash === i - 1 || dots === 1) {
          // NOOP
        } else if (dots === 2) {
          if (res.length < 2 || lastSegLen !== 2 ||
              res.charCodeAt(res.length - 1) !== CHAR_DOT ||
              res.charCodeAt(res.length - 2) !== CHAR_DOT) {
            if (res.length > 2) {
              var idx = res.length - lastSegLen - 1;
              if (idx === -1) {
                res = '';
                lastSegLen = 0;
              } else {
                res = res.slice(0, idx);
                lastSegLen = res.length - 1 - res.lastIndexOf('/');
              }
              lastSlash = i;
              dots = 0;
              continue;
            } else if (res.length !== 0) {
              res = '';
              lastSegLen = 0;
              lastSlash = i;
              dots = 0;
              continue;
            }
          }
          if (allowAboveRoot) {
            res += res.length > 0 ? '/..' : '..';
            lastSegLen = 2;
          }
        } else {
          if (res.length > 0)
            res += '/' + path.slice(lastSlash + 1, i);
          else
            res = path.slice(lastSlash + 1, i);
          lastSegLen = i - lastSlash - 1;
        }
        lastSlash = i;
        dots = 0;
      } else if (code === CHAR_DOT && dots !== -1) {
        ++dots;
      } else {
        dots = -1;
      }
    }
    return res;
  }

  // =========================================================================
  // posix path
  // =========================================================================

  var posix = {};

  posix.sep = '/';
  posix.delimiter = ':';

  posix.resolve = function () {
    var resolvedPath = '';
    var resolvedAbsolute = false;
    var args = Array.prototype.slice.call(arguments);

    for (var i = args.length - 1; i >= 0 && !resolvedAbsolute; i--) {
      var path = args[i];
      assertString(path, 'path');
      if (path.length === 0) continue;
      resolvedPath = path + '/' + resolvedPath;
      resolvedAbsolute = path.charCodeAt(0) === CHAR_FORWARD_SLASH;
    }

    if (!resolvedAbsolute) {
      var cwd = (typeof process !== 'undefined' && process.cwd) ? process.cwd() : '/';
      resolvedPath = cwd + '/' + resolvedPath;
      resolvedAbsolute = cwd.charCodeAt(0) === CHAR_FORWARD_SLASH;
    }

    resolvedPath = normalizeString(resolvedPath, !resolvedAbsolute);

    if (resolvedAbsolute) return '/' + resolvedPath;
    return resolvedPath.length > 0 ? resolvedPath : '.';
  };

  posix.normalize = function (path) {
    assertString(path, 'path');
    if (path.length === 0) return '.';

    var isAbsolute = path.charCodeAt(0) === CHAR_FORWARD_SLASH;
    var trailingSep = path.charCodeAt(path.length - 1) === CHAR_FORWARD_SLASH;

    path = normalizeString(path, !isAbsolute);

    if (path.length === 0) {
      if (isAbsolute) return '/';
      return trailingSep ? './' : '.';
    }
    if (trailingSep) path += '/';
    return isAbsolute ? '/' + path : path;
  };

  posix.isAbsolute = function (path) {
    assertString(path, 'path');
    return path.length > 0 && path.charCodeAt(0) === CHAR_FORWARD_SLASH;
  };

  posix.join = function () {
    var args = Array.prototype.slice.call(arguments);
    if (args.length === 0) return '.';
    var parts = [];
    for (var i = 0; i < args.length; i++) {
      assertString(args[i], 'path');
      if (args[i].length > 0) parts.push(args[i]);
    }
    if (parts.length === 0) return '.';
    return posix.normalize(parts.join('/'));
  };

  posix.relative = function (from, to) {
    assertString(from, 'from');
    assertString(to, 'to');
    if (from === to) return '';

    from = posix.resolve(from);
    to = posix.resolve(to);
    if (from === to) return '';

    var fromStart = 1;
    var fromEnd = from.length;
    var fromLen = fromEnd - fromStart;
    var toStart = 1;
    var toLen = to.length - toStart;

    var length = fromLen < toLen ? fromLen : toLen;
    var lastCommonSep = -1;
    var i = 0;
    for (; i < length; i++) {
      var fc = from.charCodeAt(fromStart + i);
      if (fc !== to.charCodeAt(toStart + i)) break;
      else if (fc === CHAR_FORWARD_SLASH) lastCommonSep = i;
    }
    if (i === length) {
      if (toLen > length) {
        if (to.charCodeAt(toStart + i) === CHAR_FORWARD_SLASH)
          return to.slice(toStart + i + 1);
        if (i === 0) return to.slice(toStart + i);
      } else if (fromLen > length) {
        if (from.charCodeAt(fromStart + i) === CHAR_FORWARD_SLASH)
          lastCommonSep = i;
        else if (i === 0) lastCommonSep = 0;
      }
    }

    var out = '';
    for (i = fromStart + lastCommonSep + 1; i <= fromEnd; ++i) {
      if (i === fromEnd || from.charCodeAt(i) === CHAR_FORWARD_SLASH)
        out += out.length === 0 ? '..' : '/..';
    }
    return out + to.slice(toStart + lastCommonSep);
  };

  posix.dirname = function (path) {
    assertString(path, 'path');
    if (path.length === 0) return '.';
    var hasRoot = path.charCodeAt(0) === CHAR_FORWARD_SLASH;
    var end = -1;
    var matchedSlash = true;
    for (var i = path.length - 1; i >= 1; --i) {
      if (path.charCodeAt(i) === CHAR_FORWARD_SLASH) {
        if (!matchedSlash) { end = i; break; }
      } else {
        matchedSlash = false;
      }
    }
    if (end === -1) return hasRoot ? '/' : '.';
    if (hasRoot && end === 1) return '//';
    return path.slice(0, end);
  };

  posix.basename = function (path, suffix) {
    if (suffix !== undefined) assertString(suffix, 'suffix');
    assertString(path, 'path');

    var start = 0;
    var end = -1;
    var matchedSlash = true;

    if (suffix !== undefined && suffix.length > 0 && suffix.length <= path.length) {
      if (suffix === path) return '';
      var extIdx = suffix.length - 1;
      var firstNonSlashEnd = -1;
      for (var i = path.length - 1; i >= 0; --i) {
        var code = path.charCodeAt(i);
        if (code === CHAR_FORWARD_SLASH) {
          if (!matchedSlash) { start = i + 1; break; }
        } else {
          if (firstNonSlashEnd === -1) { matchedSlash = false; firstNonSlashEnd = i + 1; }
          if (extIdx >= 0) {
            if (code === suffix.charCodeAt(extIdx)) {
              if (--extIdx === -1) end = i;
            } else {
              extIdx = -1;
              end = firstNonSlashEnd;
            }
          }
        }
      }
      if (start === end) end = firstNonSlashEnd;
      else if (end === -1) end = path.length;
      return path.slice(start, end);
    }

    for (var j = path.length - 1; j >= 0; --j) {
      if (path.charCodeAt(j) === CHAR_FORWARD_SLASH) {
        if (!matchedSlash) { start = j + 1; break; }
      } else if (end === -1) {
        matchedSlash = false;
        end = j + 1;
      }
    }
    if (end === -1) return '';
    return path.slice(start, end);
  };

  posix.extname = function (path) {
    assertString(path, 'path');
    var startDot = -1;
    var startPart = 0;
    var end = -1;
    var matchedSlash = true;
    var preDotState = 0;

    for (var i = path.length - 1; i >= 0; --i) {
      var code = path.charCodeAt(i);
      if (code === CHAR_FORWARD_SLASH) {
        if (!matchedSlash) { startPart = i + 1; break; }
        continue;
      }
      if (end === -1) { matchedSlash = false; end = i + 1; }
      if (code === CHAR_DOT) {
        if (startDot === -1) startDot = i;
        else if (preDotState !== 1) preDotState = 1;
      } else if (startDot !== -1) {
        preDotState = -1;
      }
    }

    if (startDot === -1 || end === -1 || preDotState === 0 ||
        (preDotState === 1 && startDot === end - 1 && startDot === startPart + 1)) {
      return '';
    }
    return path.slice(startDot, end);
  };

  posix.format = function (pathObject) {
    if (pathObject === null || typeof pathObject !== 'object')
      throw new TypeError('The "pathObject" argument must be of type Object. Received ' + typeof pathObject);

    var dir = pathObject.dir || pathObject.root;
    var base = pathObject.base || ((pathObject.name || '') + (pathObject.ext ? (pathObject.ext[0] === '.' ? '' : '.') + pathObject.ext : ''));
    if (!dir) return base;
    return dir === pathObject.root ? dir + base : dir + '/' + base;
  };

  posix.parse = function (path) {
    assertString(path, 'path');
    var ret = { root: '', dir: '', base: '', ext: '', name: '' };
    if (path.length === 0) return ret;

    var isAbsolute = path.charCodeAt(0) === CHAR_FORWARD_SLASH;
    var start;
    if (isAbsolute) { ret.root = '/'; start = 1; } else { start = 0; }

    var startDot = -1;
    var startPart = 0;
    var end = -1;
    var matchedSlash = true;
    var preDotState = 0;

    for (var i = path.length - 1; i >= start; --i) {
      var code = path.charCodeAt(i);
      if (code === CHAR_FORWARD_SLASH) {
        if (!matchedSlash) { startPart = i + 1; break; }
        continue;
      }
      if (end === -1) { matchedSlash = false; end = i + 1; }
      if (code === CHAR_DOT) {
        if (startDot === -1) startDot = i;
        else if (preDotState !== 1) preDotState = 1;
      } else if (startDot !== -1) {
        preDotState = -1;
      }
    }

    if (end !== -1) {
      var s = (startPart === 0 && isAbsolute) ? 1 : startPart;
      if (startDot === -1 || preDotState === 0 ||
          (preDotState === 1 && startDot === end - 1 && startDot === startPart + 1)) {
        ret.base = ret.name = path.slice(s, end);
      } else {
        ret.name = path.slice(s, startDot);
        ret.base = path.slice(s, end);
        ret.ext = path.slice(startDot, end);
      }
    }

    if (startPart > 0) ret.dir = path.slice(0, startPart - 1);
    else if (isAbsolute) ret.dir = '/';

    return ret;
  };

  posix.toNamespacedPath = function (path) { return path; };

  // Self-references (Node.js compat).
  posix.posix = posix;
  posix.win32 = null;

  return posix;
})();