(function () {
  'use strict';

  var Buffer = require('buffer').Buffer;
  var _z = __zlib;

  function toBuffer(input) {
    if (typeof input === 'string') return Buffer.from(input);
    if (input && input._isBuffer) return input;
    if (Array.isArray(input)) return Buffer.from(input);
    throw new TypeError('buffer must be a string, Buffer, or Array');
  }

  function getLevel(opts) {
    if (opts && typeof opts.level === 'number') return opts.level;
    return -1; // default
  }

  function getQuality(opts) {
    if (opts && typeof opts.params === 'object') {
      var q = opts.params[mod.constants.BROTLI_PARAM_QUALITY];
      if (typeof q === 'number') return q;
    }
    return -1; // default
  }

  // --- Sync methods ---

  function deflateSync(buffer, opts) {
    var buf = toBuffer(buffer);
    return Buffer.from(_z.deflateSync(buf.toString('latin1'), getLevel(opts)), 'latin1');
  }

  function inflateSync(buffer, opts) {
    var buf = toBuffer(buffer);
    return Buffer.from(_z.inflateSync(buf.toString('latin1')), 'latin1');
  }

  function deflateRawSync(buffer, opts) {
    var buf = toBuffer(buffer);
    return Buffer.from(_z.deflateRawSync(buf.toString('latin1'), getLevel(opts)), 'latin1');
  }

  function inflateRawSync(buffer, opts) {
    var buf = toBuffer(buffer);
    return Buffer.from(_z.inflateRawSync(buf.toString('latin1')), 'latin1');
  }

  function gzipSync(buffer, opts) {
    var buf = toBuffer(buffer);
    return Buffer.from(_z.gzipSync(buf.toString('latin1'), getLevel(opts)), 'latin1');
  }

  function gunzipSync(buffer, opts) {
    var buf = toBuffer(buffer);
    return Buffer.from(_z.gunzipSync(buf.toString('latin1')), 'latin1');
  }

  function unzipSync(buffer, opts) {
    var buf = toBuffer(buffer);
    return Buffer.from(_z.unzipSync(buf.toString('latin1')), 'latin1');
  }

  function brotliCompressSync(buffer, opts) {
    var buf = toBuffer(buffer);
    return Buffer.from(_z.brotliCompressSync(buf.toString('latin1'), getQuality(opts)), 'latin1');
  }

  function brotliDecompressSync(buffer, opts) {
    var buf = toBuffer(buffer);
    return Buffer.from(_z.brotliDecompressSync(buf.toString('latin1')), 'latin1');
  }

  // --- Callback wrappers (synchronous under the hood) ---

  function wrapAsync(syncFn) {
    return function (buffer, opts, callback) {
      if (typeof opts === 'function') { callback = opts; opts = {}; }
      try {
        var result = syncFn(buffer, opts);
        if (typeof callback === 'function') callback(null, result);
      } catch (e) {
        if (typeof callback === 'function') callback(e);
        else throw e;
      }
    };
  }

  // --- crc32 ---

  function crc32(data, value) {
    var buf = toBuffer(data);
    return _z.crc32(buf.toString('latin1'), value || 0);
  }

  // --- Constants ---

  var constants = {
    Z_NO_FLUSH: 0,
    Z_PARTIAL_FLUSH: 1,
    Z_SYNC_FLUSH: 2,
    Z_FULL_FLUSH: 3,
    Z_FINISH: 4,
    Z_BLOCK: 5,
    Z_TREES: 6,

    Z_OK: 0,
    Z_STREAM_END: 1,
    Z_NEED_DICT: 2,
    Z_ERRNO: -1,
    Z_STREAM_ERROR: -2,
    Z_DATA_ERROR: -3,
    Z_MEM_ERROR: -4,
    Z_BUF_ERROR: -5,
    Z_VERSION_ERROR: -6,

    Z_NO_COMPRESSION: 0,
    Z_BEST_SPEED: 1,
    Z_BEST_COMPRESSION: 9,
    Z_DEFAULT_COMPRESSION: -1,

    Z_FILTERED: 1,
    Z_HUFFMAN_ONLY: 2,
    Z_RLE: 3,
    Z_FIXED: 4,
    Z_DEFAULT_STRATEGY: 0,

    Z_DEFAULT_WINDOWBITS: 15,
    Z_MIN_WINDOWBITS: 8,
    Z_MAX_WINDOWBITS: 15,
    Z_MIN_CHUNK: 64,
    Z_MAX_CHUNK: Infinity,
    Z_DEFAULT_CHUNK: 16384,
    Z_MIN_MEMLEVEL: 1,
    Z_MAX_MEMLEVEL: 9,
    Z_DEFAULT_MEMLEVEL: 8,
    Z_MIN_LEVEL: -1,
    Z_MAX_LEVEL: 9,

    BROTLI_PARAM_QUALITY: 1,
    BROTLI_MIN_QUALITY: 0,
    BROTLI_MAX_QUALITY: 11,
    BROTLI_DEFAULT_QUALITY: 11,
    BROTLI_PARAM_LGWIN: 2,
    BROTLI_PARAM_LGBLOCK: 3,
    BROTLI_PARAM_MODE: 0,
    BROTLI_MODE_GENERIC: 0,
    BROTLI_MODE_TEXT: 1,
    BROTLI_MODE_FONT: 2,

    BROTLI_OPERATION_PROCESS: 0,
    BROTLI_OPERATION_FLUSH: 1,
    BROTLI_OPERATION_FINISH: 2,

    DEFLATE: 1,
    INFLATE: 2,
    GZIP: 3,
    GUNZIP: 4,
    DEFLATERAW: 5,
    INFLATERAW: 6,
    UNZIP: 7,
    BROTLI_DECODE: 8,
    BROTLI_ENCODE: 9
  };

  // --- Module export ---

  var mod = {
    // Sync methods
    deflateSync: deflateSync,
    inflateSync: inflateSync,
    deflateRawSync: deflateRawSync,
    inflateRawSync: inflateRawSync,
    gzipSync: gzipSync,
    gunzipSync: gunzipSync,
    unzipSync: unzipSync,
    brotliCompressSync: brotliCompressSync,
    brotliDecompressSync: brotliDecompressSync,

    // Callback methods
    deflate: wrapAsync(deflateSync),
    inflate: wrapAsync(inflateSync),
    deflateRaw: wrapAsync(deflateRawSync),
    inflateRaw: wrapAsync(inflateRawSync),
    gzip: wrapAsync(gzipSync),
    gunzip: wrapAsync(gunzipSync),
    unzip: wrapAsync(unzipSync),
    brotliCompress: wrapAsync(brotliCompressSync),
    brotliDecompress: wrapAsync(brotliDecompressSync),

    crc32: crc32,
    constants: constants
  };

  return mod;
})();