(function () {
  'use strict';

  var _fs = __fs;

  var mod = {};

  // ── Sync API ──────────────────────────────────────────────────────

  mod.readFileSync = function (path, options) { return _fs.readFileSync(path, options); };
  mod.writeFileSync = function (path, data, options) { return _fs.writeFileSync(path, data, options); };
  mod.appendFileSync = function (path, data, options) { return _fs.appendFileSync(path, data, options); };
  mod.existsSync = function (path) { return _fs.existsSync(path); };
  mod.accessSync = function (path, mode) { return _fs.accessSync(path, mode); };
  mod.statSync = function (path, options) { return _fs.statSync(path, options); };
  mod.lstatSync = function (path, options) { return _fs.lstatSync(path, options); };
  mod.readdirSync = function (path, options) { return _fs.readdirSync(path, options); };
  mod.mkdirSync = function (path, options) { return _fs.mkdirSync(path, options); };
  mod.rmdirSync = function (path, options) { return _fs.rmdirSync(path, options); };
  mod.rmSync = function (path, options) { return _fs.rmSync(path, options); };
  mod.unlinkSync = function (path) { return _fs.unlinkSync(path); };
  mod.renameSync = function (oldPath, newPath) { return _fs.renameSync(oldPath, newPath); };
  mod.copyFileSync = function (src, dest, mode) { return _fs.copyFileSync(src, dest, mode); };
  mod.chmodSync = function (path, mode) { return _fs.chmodSync(path, mode); };
  mod.chownSync = function (path, uid, gid) { return _fs.chownSync(path, uid, gid); };
  mod.truncateSync = function (path, len) { return _fs.truncateSync(path, len); };
  mod.mkdtempSync = function (prefix, options) { return _fs.mkdtempSync(prefix, options); };
  mod.realpathSync = function (path, options) { return _fs.realpathSync(path, options); };
  mod.readlinkSync = function (path, options) { return _fs.readlinkSync(path, options); };
  mod.symlinkSync = function (target, path, type) { return _fs.symlinkSync(target, path, type); };
  mod.linkSync = function (existingPath, newPath) { return _fs.linkSync(existingPath, newPath); };

  // ── Constants ─────────────────────────────────────────────────────

  mod.constants = _fs.constants;

  // Re-export commonly used constants at top level
  mod.F_OK = _fs.constants.F_OK;
  mod.R_OK = _fs.constants.R_OK;
  mod.W_OK = _fs.constants.W_OK;
  mod.X_OK = _fs.constants.X_OK;

  return mod;
})();