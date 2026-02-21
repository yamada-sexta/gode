(function () {
  'use strict';

  var _os = __os;

  var mod = {};

  mod.hostname = function () { return _os.hostname(); };
  mod.homedir = function () { return _os.homedir(); };
  mod.tmpdir = function () { return _os.tmpdir(); };
  mod.platform = function () { return _os.platform(); };
  mod.arch = function () { return _os.arch(); };
  mod.type = function () { return _os.type(); };
  mod.release = function () { return _os.release(); };
  mod.version = function () { return _os.version(); };
  mod.machine = function () { return _os.machine(); };
  mod.endianness = function () { return _os.endianness(); };
  mod.uptime = function () { return _os.uptime(); };
  mod.freemem = function () { return _os.freemem(); };
  mod.totalmem = function () { return _os.totalmem(); };
  mod.availableParallelism = function () { return _os.availableParallelism(); };
  mod.loadavg = function () { return _os.loadavg(); };
  mod.cpus = function () { return _os.cpus(); };
  mod.networkInterfaces = function () { return _os.networkInterfaces(); };
  mod.userInfo = function () { return _os.userInfo(); };

  mod.EOL = _os.eol;
  mod.devNull = _os.devNull;

  mod.constants = {
    signals: {},
    errno: {},
    priority: {
      PRIORITY_LOW: 19,
      PRIORITY_BELOW_NORMAL: 10,
      PRIORITY_NORMAL: 0,
      PRIORITY_ABOVE_NORMAL: -7,
      PRIORITY_HIGH: -14,
      PRIORITY_HIGHEST: -20
    }
  };

  return mod;
})();