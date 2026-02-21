(function () {
  'use strict';

  // ===========================================================================
  // Encoding helpers
  // ===========================================================================

  var ENCODINGS = {
    'utf8': 'utf8', 'utf-8': 'utf8',
    'ascii': 'ascii',
    'latin1': 'latin1', 'binary': 'latin1',
    'hex': 'hex',
    'base64': 'base64',
    'base64url': 'base64url',
    'ucs2': 'utf16le', 'ucs-2': 'utf16le', 'utf16le': 'utf16le', 'utf-16le': 'utf16le'
  };

  function normalizeEncoding(enc) {
    if (!enc) return 'utf8';
    var lower = enc.toLowerCase();
    var mapped = ENCODINGS[lower];
    if (!mapped) throw new TypeError('Unknown encoding: ' + enc);
    return mapped;
  }

  // --- UTF-8 ---

  function utf8ToBytes(str) {
    var bytes = [];
    for (var i = 0; i < str.length; i++) {
      var c = str.charCodeAt(i);
      if (c < 0x80) {
        bytes.push(c);
      } else if (c < 0x800) {
        bytes.push(0xC0 | (c >> 6), 0x80 | (c & 0x3F));
      } else if (c >= 0xD800 && c <= 0xDBFF) {
        var next = str.charCodeAt(++i);
        if (next >= 0xDC00 && next <= 0xDFFF) {
          var cp = ((c - 0xD800) << 10) + (next - 0xDC00) + 0x10000;
          bytes.push(0xF0 | (cp >> 18), 0x80 | ((cp >> 12) & 0x3F),
            0x80 | ((cp >> 6) & 0x3F), 0x80 | (cp & 0x3F));
        } else {
          bytes.push(0xEF, 0xBF, 0xBD);
          i--;
        }
      } else if (c >= 0xDC00 && c <= 0xDFFF) {
        bytes.push(0xEF, 0xBF, 0xBD);
      } else {
        bytes.push(0xE0 | (c >> 12), 0x80 | ((c >> 6) & 0x3F), 0x80 | (c & 0x3F));
      }
    }
    return bytes;
  }

  function bytesToUtf8(data, start, end) {
    var str = '';
    var i = start;
    while (i < end) {
      var b = data[i];
      if (b < 0x80) {
        str += String.fromCharCode(b);
        i++;
      } else if ((b & 0xE0) === 0xC0 && i + 1 < end) {
        str += String.fromCharCode(((b & 0x1F) << 6) | (data[i + 1] & 0x3F));
        i += 2;
      } else if ((b & 0xF0) === 0xE0 && i + 2 < end) {
        str += String.fromCharCode(((b & 0x0F) << 12) | ((data[i + 1] & 0x3F) << 6) |
          (data[i + 2] & 0x3F));
        i += 3;
      } else if ((b & 0xF8) === 0xF0 && i + 3 < end) {
        var cp = ((b & 0x07) << 18) | ((data[i + 1] & 0x3F) << 12) |
          ((data[i + 2] & 0x3F) << 6) | (data[i + 3] & 0x3F);
        if (cp > 0xFFFF) {
          cp -= 0x10000;
          str += String.fromCharCode((cp >> 10) + 0xD800, (cp & 0x3FF) + 0xDC00);
        } else {
          str += String.fromCharCode(cp);
        }
        i += 4;
      } else {
        str += '\uFFFD';
        i++;
      }
    }
    return str;
  }

  // --- Latin-1 / ASCII ---

  function latin1ToBytes(str) {
    var bytes = [];
    for (var i = 0; i < str.length; i++) bytes.push(str.charCodeAt(i) & 0xFF);
    return bytes;
  }

  function bytesToLatin1(data, start, end) {
    var str = '';
    for (var i = start; i < end; i++) str += String.fromCharCode(data[i]);
    return str;
  }

  function asciiToBytes(str) {
    var bytes = [];
    for (var i = 0; i < str.length; i++) bytes.push(str.charCodeAt(i) & 0x7F);
    return bytes;
  }

  // --- UTF-16LE ---

  function utf16leToBytes(str) {
    var bytes = [];
    for (var i = 0; i < str.length; i++) {
      var c = str.charCodeAt(i);
      bytes.push(c & 0xFF, (c >> 8) & 0xFF);
    }
    return bytes;
  }

  function bytesToUtf16le(data, start, end) {
    var str = '';
    for (var i = start; i + 1 < end; i += 2) {
      str += String.fromCharCode(data[i] | (data[i + 1] << 8));
    }
    return str;
  }

  // --- Hex ---

  var HEX_CHARS = '0123456789abcdef';

  function hexToBytes(str) {
    str = str.replace(/[^0-9a-fA-F]/g, '');
    if (str.length % 2 !== 0) str = str.slice(0, -1);
    var bytes = [];
    for (var i = 0; i < str.length; i += 2) bytes.push(parseInt(str.substr(i, 2), 16));
    return bytes;
  }

  function bytesToHex(data, start, end) {
    var str = '';
    for (var i = start; i < end; i++) {
      str += HEX_CHARS.charAt(data[i] >> 4) + HEX_CHARS.charAt(data[i] & 0x0F);
    }
    return str;
  }

  // --- Base64 ---

  var B64 = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/';
  var B64_LOOKUP = {};
  for (var _i = 0; _i < B64.length; _i++) B64_LOOKUP[B64.charAt(_i)] = _i;
  B64_LOOKUP['-'] = 62;
  B64_LOOKUP['_'] = 63;

  function base64ToBytes(str) {
    str = str.replace(/[\s=]/g, '');
    var bytes = [];
    for (var i = 0; i < str.length; i += 4) {
      var a = B64_LOOKUP[str.charAt(i)] || 0;
      var b = B64_LOOKUP[str.charAt(i + 1)] || 0;
      var c = B64_LOOKUP[str.charAt(i + 2)];
      var d = B64_LOOKUP[str.charAt(i + 3)];
      bytes.push((a << 2) | (b >> 4));
      if (c !== undefined) bytes.push(((b & 0xF) << 4) | (c >> 2));
      if (d !== undefined) bytes.push(((c & 0x3) << 6) | d);
    }
    return bytes;
  }

  function bytesToBase64(data, start, end) {
    var str = '';
    for (var i = start; i < end; i += 3) {
      var a = data[i];
      var b = i + 1 < end ? data[i + 1] : 0;
      var c = i + 2 < end ? data[i + 2] : 0;
      str += B64.charAt(a >> 2);
      str += B64.charAt(((a & 0x3) << 4) | (b >> 4));
      str += (i + 1 < end) ? B64.charAt(((b & 0xF) << 2) | (c >> 6)) : '=';
      str += (i + 2 < end) ? B64.charAt(c & 0x3F) : '=';
    }
    return str;
  }

  function bytesToBase64url(data, start, end) {
    return bytesToBase64(data, start, end)
      .replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');
  }

  // Master encode/decode

  function stringToBytes(str, encoding) {
    switch (normalizeEncoding(encoding)) {
      case 'utf8':     return utf8ToBytes(str);
      case 'ascii':    return asciiToBytes(str);
      case 'latin1':   return latin1ToBytes(str);
      case 'hex':      return hexToBytes(str);
      case 'base64':   return base64ToBytes(str);
      case 'base64url':return base64ToBytes(str);
      case 'utf16le':  return utf16leToBytes(str);
      default:         return utf8ToBytes(str);
    }
  }

  function bytesToString(data, encoding, start, end) {
    switch (normalizeEncoding(encoding)) {
      case 'utf8':      return bytesToUtf8(data, start, end);
      case 'ascii':     return bytesToLatin1(data, start, end);
      case 'latin1':    return bytesToLatin1(data, start, end);
      case 'hex':       return bytesToHex(data, start, end);
      case 'base64':    return bytesToBase64(data, start, end);
      case 'base64url': return bytesToBase64url(data, start, end);
      case 'utf16le':   return bytesToUtf16le(data, start, end);
      default:          return bytesToUtf8(data, start, end);
    }
  }

  // ===========================================================================
  // IEEE 754 helpers (for float/double read/write without DataView)
  // ===========================================================================

  function ieee754Read(data, offset, isLE, mLen, nBytes) {
    var e, m;
    var eLen = nBytes * 8 - mLen - 1;
    var eMax = (1 << eLen) - 1;
    var eBias = eMax >> 1;
    var nBits = -7;
    var i = isLE ? (nBytes - 1) : 0;
    var d = isLE ? -1 : 1;
    var s = data[offset + i];
    i += d;
    e = s & ((1 << (-nBits)) - 1);
    s >>= (-nBits);
    nBits += eLen;
    for (; nBits > 0; e = e * 256 + data[offset + i], i += d, nBits -= 8) {}
    m = e & ((1 << (-nBits)) - 1);
    e >>= (-nBits);
    nBits += mLen;
    for (; nBits > 0; m = m * 256 + data[offset + i], i += d, nBits -= 8) {}
    if (e === 0) {
      e = 1 - eBias;
    } else if (e === eMax) {
      return m ? NaN : ((s ? -1 : 1) * Infinity);
    } else {
      m = m + Math.pow(2, mLen);
      e = e - eBias;
    }
    return (s ? -1 : 1) * m * Math.pow(2, e - mLen);
  }

  function ieee754Write(data, value, offset, isLE, mLen, nBytes) {
    var e, m, c;
    var eLen = nBytes * 8 - mLen - 1;
    var eMax = (1 << eLen) - 1;
    var eBias = eMax >> 1;
    var rt = (mLen === 23 ? Math.pow(2, -24) - Math.pow(2, -77) : 0);
    var i = isLE ? 0 : nBytes - 1;
    var d = isLE ? 1 : -1;
    var s = value < 0 || (value === 0 && 1 / value < 0) ? 1 : 0;
    value = Math.abs(value);
    if (value !== value || value === Infinity) {
      m = value !== value ? 1 : 0;
      e = eMax;
    } else {
      e = Math.floor(Math.log(value) / Math.LN2);
      c = Math.pow(2, -e);
      if (value * c < 1) { e--; c *= 2; }
      if (e + eBias >= 1) {
        value += rt / c;
      } else {
        value += rt * Math.pow(2, 1 - eBias);
      }
      if (value * c >= 2) { e++; c /= 2; }
      if (e + eBias >= eMax) {
        m = 0;
        e = eMax;
      } else if (e + eBias >= 1) {
        m = ((value * c) - 1) * Math.pow(2, mLen);
        e = e + eBias;
      } else {
        m = value * Math.pow(2, eBias - 1) * Math.pow(2, mLen);
        e = 0;
      }
    }
    for (; mLen >= 8; data[offset + i] = m & 0xFF, i += d, m /= 256, mLen -= 8) {}
    e = (e << mLen) | m;
    eLen += mLen;
    for (; eLen > 0; data[offset + i] = e & 0xFF, i += d, e /= 256, eLen -= 8) {}
    data[offset + i - d] |= s * 128;
  }

  // ===========================================================================
  // Buffer class
  // ===========================================================================

  function Buffer(data) {
    if (!(this instanceof Buffer)) return new Buffer(data);
    this._data = data;
    this.length = data.length;
    _defineIndexProps(this);
  }

  // Tag for isBuffer
  Buffer.prototype._isBuffer = true;

  function _defineIndexProps(buf) {
    for (var i = 0; i < buf.length; i++) {
      (function (idx) {
        Object.defineProperty(buf, idx, {
          get: function () { return buf._data[idx]; },
          set: function (v) { buf._data[idx] = v & 0xFF; },
          enumerable: true,
          configurable: false
        });
      })(i);
    }
  }

  // ===========================================================================
  // Static methods
  // ===========================================================================

  Buffer.alloc = function (size, fill, encoding) {
    if (typeof size !== 'number' || size < 0) throw new RangeError('Invalid size');
    var data = [];
    for (var i = 0; i < size; i++) data.push(0);
    var buf = new Buffer(data);
    if (fill !== undefined) buf.fill(fill, 0, size, encoding);
    return buf;
  };

  Buffer.allocUnsafe = function (size) {
    if (typeof size !== 'number' || size < 0) throw new RangeError('Invalid size');
    var data = [];
    for (var i = 0; i < size; i++) data.push(0);
    return new Buffer(data);
  };

  Buffer.allocUnsafeSlow = Buffer.allocUnsafe;

  Buffer.from = function (value, encodingOrOffset, length) {
    if (typeof value === 'string') {
      return new Buffer(stringToBytes(value, encodingOrOffset));
    }
    if (Array.isArray(value)) {
      var data = [];
      for (var i = 0; i < value.length; i++) {
        var v = typeof value[i] === 'number' ? value[i] : parseInt(value[i], 10);
        data.push((v & 0xFF));
      }
      return new Buffer(data);
    }
    if (value && value._isBuffer) {
      var copy = [];
      for (var j = 0; j < value.length; j++) copy.push(value._data[j]);
      return new Buffer(copy);
    }
    throw new TypeError('First argument must be a string, Buffer, or Array');
  };

  Buffer.isBuffer = function (obj) {
    return obj != null && obj._isBuffer === true;
  };

  Buffer.isEncoding = function (encoding) {
    if (typeof encoding !== 'string') return false;
    return ENCODINGS[encoding.toLowerCase()] !== undefined;
  };

  Buffer.byteLength = function (string, encoding) {
    if (Buffer.isBuffer(string)) return string.length;
    if (typeof string !== 'string') throw new TypeError('argument must be a string or Buffer');
    return stringToBytes(string, encoding).length;
  };

  Buffer.compare = function (a, b) {
    if (!Buffer.isBuffer(a) || !Buffer.isBuffer(b))
      throw new TypeError('Arguments must be Buffers');
    if (a === b) return 0;
    var len = Math.min(a.length, b.length);
    for (var i = 0; i < len; i++) {
      if (a._data[i] !== b._data[i]) return a._data[i] < b._data[i] ? -1 : 1;
    }
    if (a.length < b.length) return -1;
    if (a.length > b.length) return 1;
    return 0;
  };

  Buffer.concat = function (list, totalLength) {
    if (!Array.isArray(list)) throw new TypeError('list must be an Array');
    if (list.length === 0) return Buffer.alloc(0);
    if (totalLength === undefined) {
      totalLength = 0;
      for (var i = 0; i < list.length; i++) totalLength += list[i].length;
    }
    var data = [];
    var pos = 0;
    for (var j = 0; j < list.length; j++) {
      var buf = list[j];
      for (var k = 0; k < buf.length && pos < totalLength; k++, pos++) {
        data.push(buf._data[k]);
      }
    }
    while (data.length < totalLength) data.push(0);
    return new Buffer(data);
  };

  Buffer.poolSize = 8192;

  // ===========================================================================
  // Instance methods
  // ===========================================================================

  Buffer.prototype.toString = function (encoding, start, end) {
    start = start || 0;
    end = (end === undefined || end > this.length) ? this.length : end;
    if (start >= end) return '';
    return bytesToString(this._data, encoding, start, end);
  };

  Buffer.prototype.toJSON = function () {
    var arr = [];
    for (var i = 0; i < this.length; i++) arr.push(this._data[i]);
    return { type: 'Buffer', data: arr };
  };

  Buffer.prototype.equals = function (other) {
    if (!Buffer.isBuffer(other)) throw new TypeError('Argument must be a Buffer');
    if (this === other) return true;
    return Buffer.compare(this, other) === 0;
  };

  Buffer.prototype.compare = function (target, targetStart, targetEnd, sourceStart, sourceEnd) {
    if (!Buffer.isBuffer(target)) throw new TypeError('Argument must be a Buffer');
    targetStart = targetStart || 0;
    targetEnd = (targetEnd === undefined) ? target.length : targetEnd;
    sourceStart = sourceStart || 0;
    sourceEnd = (sourceEnd === undefined) ? this.length : sourceEnd;
    var i = sourceStart, j = targetStart;
    while (i < sourceEnd && j < targetEnd) {
      if (this._data[i] !== target._data[j])
        return this._data[i] < target._data[j] ? -1 : 1;
      i++; j++;
    }
    var sLen = sourceEnd - sourceStart;
    var tLen = targetEnd - targetStart;
    if (sLen < tLen) return -1;
    if (sLen > tLen) return 1;
    return 0;
  };

  Buffer.prototype.copy = function (target, targetStart, sourceStart, sourceEnd) {
    targetStart = targetStart || 0;
    sourceStart = sourceStart || 0;
    sourceEnd = (sourceEnd === undefined) ? this.length : sourceEnd;
    var nb = Math.min(sourceEnd - sourceStart, target.length - targetStart);
    for (var i = 0; i < nb; i++) {
      target._data[targetStart + i] = this._data[sourceStart + i];
      // Update indexed property if it exists
      if (targetStart + i < target.length) {
        // Property already defined in constructor
      }
    }
    return nb;
  };

  Buffer.prototype.slice = function (start, end) {
    start = start || 0;
    end = (end === undefined) ? this.length : end;
    if (start < 0) start = Math.max(this.length + start, 0);
    if (end < 0) end = Math.max(this.length + end, 0);
    if (end > this.length) end = this.length;
    if (start >= end) return Buffer.alloc(0);
    var data = [];
    for (var i = start; i < end; i++) data.push(this._data[i]);
    return new Buffer(data);
  };

  Buffer.prototype.subarray = Buffer.prototype.slice;

  Buffer.prototype.fill = function (value, offset, end, encoding) {
    offset = offset || 0;
    end = (end === undefined) ? this.length : end;
    if (typeof value === 'string') {
      if (value.length === 0) { value = 0; }
      else if (value.length === 1) { value = value.charCodeAt(0); }
      else {
        var bytes = stringToBytes(value, encoding);
        for (var j = offset; j < end; j++) this._data[j] = bytes[(j - offset) % bytes.length];
        return this;
      }
    }
    value = value & 0xFF;
    for (var i = offset; i < end; i++) this._data[i] = value;
    return this;
  };

  Buffer.prototype.write = function (string, offset, length, encoding) {
    if (typeof offset === 'string') { encoding = offset; offset = 0; length = this.length; }
    else if (typeof length === 'string') { encoding = length; length = this.length - offset; }
    offset = offset || 0;
    if (length === undefined) length = this.length - offset;
    var bytes = stringToBytes(string, encoding);
    var nb = Math.min(bytes.length, length, this.length - offset);
    for (var i = 0; i < nb; i++) this._data[offset + i] = bytes[i];
    return nb;
  };

  Buffer.prototype.indexOf = function (value, byteOffset, encoding) {
    return _indexOf(this, value, byteOffset, encoding, false);
  };

  Buffer.prototype.lastIndexOf = function (value, byteOffset, encoding) {
    return _indexOf(this, value, byteOffset, encoding, true);
  };

  Buffer.prototype.includes = function (value, byteOffset, encoding) {
    return this.indexOf(value, byteOffset, encoding) !== -1;
  };

  function _indexOf(buf, value, byteOffset, encoding, reverse) {
    if (typeof byteOffset === 'string') { encoding = byteOffset; byteOffset = undefined; }
    if (byteOffset === undefined) byteOffset = reverse ? buf.length - 1 : 0;
    var searchBytes;
    if (typeof value === 'number') {
      searchBytes = [value & 0xFF];
    } else if (typeof value === 'string') {
      searchBytes = stringToBytes(value, encoding);
    } else if (Buffer.isBuffer(value)) {
      searchBytes = value._data;
    } else {
      throw new TypeError('value must be a string, number, or Buffer');
    }
    if (searchBytes.length === 0) return -1;
    if (reverse) {
      if (byteOffset >= buf.length) byteOffset = buf.length - 1;
      for (var i = byteOffset; i >= 0; i--) {
        if (_matchAt(buf._data, i, searchBytes)) return i;
      }
    } else {
      for (var j = byteOffset; j <= buf.length - searchBytes.length; j++) {
        if (_matchAt(buf._data, j, searchBytes)) return j;
      }
    }
    return -1;
  }

  function _matchAt(data, pos, search) {
    for (var i = 0; i < search.length; i++) {
      if (data[pos + i] !== search[i]) return false;
    }
    return true;
  }

  // --- Swap ---

  Buffer.prototype.swap16 = function () {
    if (this.length % 2 !== 0) throw new RangeError('Buffer size must be a multiple of 16-bits');
    for (var i = 0; i < this.length; i += 2) {
      var t = this._data[i]; this._data[i] = this._data[i + 1]; this._data[i + 1] = t;
    }
    return this;
  };

  Buffer.prototype.swap32 = function () {
    if (this.length % 4 !== 0) throw new RangeError('Buffer size must be a multiple of 32-bits');
    for (var i = 0; i < this.length; i += 4) {
      var t0 = this._data[i], t1 = this._data[i + 1];
      this._data[i] = this._data[i + 3]; this._data[i + 1] = this._data[i + 2];
      this._data[i + 2] = t1; this._data[i + 3] = t0;
    }
    return this;
  };

  Buffer.prototype.swap64 = function () {
    if (this.length % 8 !== 0) throw new RangeError('Buffer size must be a multiple of 64-bits');
    for (var i = 0; i < this.length; i += 8) {
      for (var lo = 0; lo < 4; lo++) {
        var t = this._data[i + lo];
        this._data[i + lo] = this._data[i + 7 - lo];
        this._data[i + 7 - lo] = t;
      }
    }
    return this;
  };

  // --- Iteration helpers (return arrays in ES5) ---

  Buffer.prototype.keys = function () {
    var arr = [];
    for (var i = 0; i < this.length; i++) arr.push(i);
    return arr;
  };

  Buffer.prototype.values = function () {
    var arr = [];
    for (var i = 0; i < this.length; i++) arr.push(this._data[i]);
    return arr;
  };

  Buffer.prototype.entries = function () {
    var arr = [];
    for (var i = 0; i < this.length; i++) arr.push([i, this._data[i]]);
    return arr;
  };

  // ===========================================================================
  // Read methods – unsigned integers
  // ===========================================================================

  Buffer.prototype.readUInt8 = function (offset) {
    offset = offset || 0;
    return this._data[offset];
  };

  Buffer.prototype.readUInt16BE = function (offset) {
    offset = offset || 0;
    return (this._data[offset] << 8) | this._data[offset + 1];
  };

  Buffer.prototype.readUInt16LE = function (offset) {
    offset = offset || 0;
    return this._data[offset] | (this._data[offset + 1] << 8);
  };

  Buffer.prototype.readUInt32BE = function (offset) {
    offset = offset || 0;
    return (this._data[offset] * 0x1000000) + ((this._data[offset + 1] << 16) |
      (this._data[offset + 2] << 8) | this._data[offset + 3]);
  };

  Buffer.prototype.readUInt32LE = function (offset) {
    offset = offset || 0;
    return ((this._data[offset + 3] * 0x1000000) + ((this._data[offset + 2] << 16) |
      (this._data[offset + 1] << 8) | this._data[offset]));
  };

  Buffer.prototype.readUIntBE = function (offset, byteLength) {
    var val = 0;
    for (var i = 0; i < byteLength; i++) val = val * 256 + this._data[offset + i];
    return val;
  };

  Buffer.prototype.readUIntLE = function (offset, byteLength) {
    var val = 0;
    var mul = 1;
    for (var i = 0; i < byteLength; i++) { val += this._data[offset + i] * mul; mul *= 256; }
    return val;
  };

  // ===========================================================================
  // Read methods – signed integers
  // ===========================================================================

  Buffer.prototype.readInt8 = function (offset) {
    offset = offset || 0;
    var val = this._data[offset];
    return val >= 0x80 ? val - 0x100 : val;
  };

  Buffer.prototype.readInt16BE = function (offset) {
    offset = offset || 0;
    var val = (this._data[offset] << 8) | this._data[offset + 1];
    return val >= 0x8000 ? val - 0x10000 : val;
  };

  Buffer.prototype.readInt16LE = function (offset) {
    offset = offset || 0;
    var val = this._data[offset] | (this._data[offset + 1] << 8);
    return val >= 0x8000 ? val - 0x10000 : val;
  };

  Buffer.prototype.readInt32BE = function (offset) {
    offset = offset || 0;
    return (this._data[offset] << 24) | (this._data[offset + 1] << 16) |
      (this._data[offset + 2] << 8) | this._data[offset + 3];
  };

  Buffer.prototype.readInt32LE = function (offset) {
    offset = offset || 0;
    return (this._data[offset + 3] << 24) | (this._data[offset + 2] << 16) |
      (this._data[offset + 1] << 8) | this._data[offset];
  };

  Buffer.prototype.readIntBE = function (offset, byteLength) {
    var val = this.readUIntBE(offset, byteLength);
    var limit = Math.pow(2, byteLength * 8 - 1);
    return val >= limit ? val - limit * 2 : val;
  };

  Buffer.prototype.readIntLE = function (offset, byteLength) {
    var val = this.readUIntLE(offset, byteLength);
    var limit = Math.pow(2, byteLength * 8 - 1);
    return val >= limit ? val - limit * 2 : val;
  };

  // ===========================================================================
  // Read methods – floats
  // ===========================================================================

  Buffer.prototype.readFloatBE = function (offset) {
    return ieee754Read(this._data, offset || 0, false, 23, 4);
  };
  Buffer.prototype.readFloatLE = function (offset) {
    return ieee754Read(this._data, offset || 0, true, 23, 4);
  };
  Buffer.prototype.readDoubleBE = function (offset) {
    return ieee754Read(this._data, offset || 0, false, 52, 8);
  };
  Buffer.prototype.readDoubleLE = function (offset) {
    return ieee754Read(this._data, offset || 0, true, 52, 8);
  };

  // ===========================================================================
  // Write methods – unsigned integers
  // ===========================================================================

  Buffer.prototype.writeUInt8 = function (value, offset) {
    offset = offset || 0;
    this._data[offset] = value & 0xFF;
    return offset + 1;
  };

  Buffer.prototype.writeUInt16BE = function (value, offset) {
    offset = offset || 0;
    this._data[offset] = (value >> 8) & 0xFF;
    this._data[offset + 1] = value & 0xFF;
    return offset + 2;
  };

  Buffer.prototype.writeUInt16LE = function (value, offset) {
    offset = offset || 0;
    this._data[offset] = value & 0xFF;
    this._data[offset + 1] = (value >> 8) & 0xFF;
    return offset + 2;
  };

  Buffer.prototype.writeUInt32BE = function (value, offset) {
    offset = offset || 0;
    this._data[offset] = (value >>> 24) & 0xFF;
    this._data[offset + 1] = (value >>> 16) & 0xFF;
    this._data[offset + 2] = (value >>> 8) & 0xFF;
    this._data[offset + 3] = value & 0xFF;
    return offset + 4;
  };

  Buffer.prototype.writeUInt32LE = function (value, offset) {
    offset = offset || 0;
    this._data[offset] = value & 0xFF;
    this._data[offset + 1] = (value >>> 8) & 0xFF;
    this._data[offset + 2] = (value >>> 16) & 0xFF;
    this._data[offset + 3] = (value >>> 24) & 0xFF;
    return offset + 4;
  };

  Buffer.prototype.writeUIntBE = function (value, offset, byteLength) {
    for (var i = byteLength - 1; i >= 0; i--) {
      this._data[offset + i] = value & 0xFF;
      value = Math.floor(value / 256);
    }
    return offset + byteLength;
  };

  Buffer.prototype.writeUIntLE = function (value, offset, byteLength) {
    for (var i = 0; i < byteLength; i++) {
      this._data[offset + i] = value & 0xFF;
      value = Math.floor(value / 256);
    }
    return offset + byteLength;
  };

  // ===========================================================================
  // Write methods – signed integers
  // ===========================================================================

  Buffer.prototype.writeInt8 = function (value, offset) {
    offset = offset || 0;
    if (value < 0) value = 0x100 + value;
    this._data[offset] = value & 0xFF;
    return offset + 1;
  };

  Buffer.prototype.writeInt16BE = function (value, offset) {
    offset = offset || 0;
    if (value < 0) value = 0x10000 + value;
    this._data[offset] = (value >> 8) & 0xFF;
    this._data[offset + 1] = value & 0xFF;
    return offset + 2;
  };

  Buffer.prototype.writeInt16LE = function (value, offset) {
    offset = offset || 0;
    if (value < 0) value = 0x10000 + value;
    this._data[offset] = value & 0xFF;
    this._data[offset + 1] = (value >> 8) & 0xFF;
    return offset + 2;
  };

  Buffer.prototype.writeInt32BE = function (value, offset) {
    offset = offset || 0;
    if (value < 0) value = 0x100000000 + value;
    this._data[offset] = (value >>> 24) & 0xFF;
    this._data[offset + 1] = (value >>> 16) & 0xFF;
    this._data[offset + 2] = (value >>> 8) & 0xFF;
    this._data[offset + 3] = value & 0xFF;
    return offset + 4;
  };

  Buffer.prototype.writeInt32LE = function (value, offset) {
    offset = offset || 0;
    if (value < 0) value = 0x100000000 + value;
    this._data[offset] = value & 0xFF;
    this._data[offset + 1] = (value >>> 8) & 0xFF;
    this._data[offset + 2] = (value >>> 16) & 0xFF;
    this._data[offset + 3] = (value >>> 24) & 0xFF;
    return offset + 4;
  };

  Buffer.prototype.writeIntBE = function (value, offset, byteLength) {
    if (value < 0) value = Math.pow(2, byteLength * 8) + value;
    return this.writeUIntBE(value, offset, byteLength);
  };

  Buffer.prototype.writeIntLE = function (value, offset, byteLength) {
    if (value < 0) value = Math.pow(2, byteLength * 8) + value;
    return this.writeUIntLE(value, offset, byteLength);
  };

  // ===========================================================================
  // Write methods – floats
  // ===========================================================================

  Buffer.prototype.writeFloatBE = function (value, offset) {
    ieee754Write(this._data, value, offset || 0, false, 23, 4);
    return (offset || 0) + 4;
  };
  Buffer.prototype.writeFloatLE = function (value, offset) {
    ieee754Write(this._data, value, offset || 0, true, 23, 4);
    return (offset || 0) + 4;
  };
  Buffer.prototype.writeDoubleBE = function (value, offset) {
    ieee754Write(this._data, value, offset || 0, false, 52, 8);
    return (offset || 0) + 8;
  };
  Buffer.prototype.writeDoubleLE = function (value, offset) {
    ieee754Write(this._data, value, offset || 0, true, 52, 8);
    return (offset || 0) + 8;
  };

  // ===========================================================================
  // Module-level APIs
  // ===========================================================================

  function _atob(data) {
    var bytes = base64ToBytes(data);
    var str = '';
    for (var i = 0; i < bytes.length; i++) str += String.fromCharCode(bytes[i]);
    return str;
  }

  function _btoa(data) {
    var bytes = [];
    for (var i = 0; i < data.length; i++) bytes.push(data.charCodeAt(i) & 0xFF);
    return bytesToBase64(bytes, 0, bytes.length);
  }

  function _isAscii(input) {
    if (Buffer.isBuffer(input)) {
      for (var i = 0; i < input.length; i++) if (input._data[i] > 127) return false;
      return true;
    }
    if (typeof input === 'string') {
      for (var j = 0; j < input.length; j++) if (input.charCodeAt(j) > 127) return false;
      return true;
    }
    return false;
  }

  function _isUtf8(input) {
    var data;
    if (Buffer.isBuffer(input)) { data = input._data; }
    else if (typeof input === 'string') { data = utf8ToBytes(input); }
    else { return false; }
    var i = 0;
    while (i < data.length) {
      var b = data[i];
      if (b < 0x80) { i++; }
      else if ((b & 0xE0) === 0xC0) {
        if (i + 1 >= data.length || (data[i + 1] & 0xC0) !== 0x80) return false;
        i += 2;
      } else if ((b & 0xF0) === 0xE0) {
        if (i + 2 >= data.length || (data[i + 1] & 0xC0) !== 0x80 ||
          (data[i + 2] & 0xC0) !== 0x80) return false;
        i += 3;
      } else if ((b & 0xF8) === 0xF0) {
        if (i + 3 >= data.length || (data[i + 1] & 0xC0) !== 0x80 ||
          (data[i + 2] & 0xC0) !== 0x80 || (data[i + 3] & 0xC0) !== 0x80) return false;
        i += 4;
      } else { return false; }
    }
    return true;
  }

  // ===========================================================================
  // Build the module export
  // ===========================================================================

  var mod = {};
  mod.Buffer = Buffer;
  mod.atob = _atob;
  mod.btoa = _btoa;
  mod.isAscii = _isAscii;
  mod.isUtf8 = _isUtf8;
  mod.INSPECT_MAX_BYTES = 50;
  mod.kMaxLength = 0x7FFFFFFF;
  mod.kStringMaxLength = 0x1FFFFFFF;
  mod.constants = {
    MAX_LENGTH: 0x7FFFFFFF,
    MAX_STRING_LENGTH: 0x1FFFFFFF
  };

  return mod;
})();
