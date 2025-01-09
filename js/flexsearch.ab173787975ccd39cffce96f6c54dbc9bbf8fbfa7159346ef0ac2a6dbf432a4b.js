(() => {
  var __getOwnPropNames = Object.getOwnPropertyNames;
  var __commonJS = (cb, mod) => function __require() {
    return mod || (0, cb[__getOwnPropNames(cb)[0]])((mod = { exports: {} }).exports, mod), mod.exports;
  };

  // (disabled):worker_threads
  var require_worker_threads = __commonJS({
    "(disabled):worker_threads"() {
    }
  });

  // node_modules/flexsearch/dist/flexsearch.bundle.module.min.js
  var t;
  function u(a) {
    return "undefined" !== typeof a ? a : true;
  }
  function v(a) {
    const b = Array(a);
    for (let c = 0; c < a; c++) b[c] = x();
    return b;
  }
  function x() {
    return /* @__PURE__ */ Object.create(null);
  }
  function aa(a, b) {
    return b.length - a.length;
  }
  function C(a) {
    return "string" === typeof a;
  }
  function D(a) {
    return "object" === typeof a;
  }
  function E(a) {
    return "function" === typeof a;
  }
  function F(a, b) {
    var c = ba;
    if (a && (b && (a = G(a, b)), this.H && (a = G(a, this.H)), this.J && 1 < a.length && (a = G(a, this.J)), c || "" === c)) {
      b = a.split(c);
      if (this.filter) {
        a = this.filter;
        c = b.length;
        const d = [];
        for (let e = 0, f = 0; e < c; e++) {
          const h = b[e];
          h && !a[h] && (d[f++] = h);
        }
        a = d;
      } else a = b;
      return a;
    }
    return a;
  }
  var ba = /[\p{Z}\p{S}\p{P}\p{C}]+/u;
  var ca = /[\u0300-\u036f]/g;
  function I(a, b) {
    const c = Object.keys(a), d = c.length, e = [];
    let f = "", h = 0;
    for (let g = 0, k, m; g < d; g++) k = c[g], (m = a[k]) ? (e[h++] = J(b ? "(?!\\b)" + k + "(\\b|_)" : k), e[h++] = m) : f += (f ? "|" : "") + k;
    f && (e[h++] = J(b ? "(?!\\b)(" + f + ")(\\b|_)" : "(" + f + ")"), e[h] = "");
    return e;
  }
  function G(a, b) {
    for (let c = 0, d = b.length; c < d && (a = a.replace(b[c], b[c + 1]), a); c += 2) ;
    return a;
  }
  function J(a) {
    return new RegExp(a, "g");
  }
  function da(a) {
    let b = "", c = "";
    for (let d = 0, e = a.length, f; d < e; d++) (f = a[d]) !== c && (b += c = f);
    return b;
  }
  var fa = { encode: ea, F: false, G: "" };
  function ea(a) {
    return F.call(this, ("" + a).toLowerCase(), false);
  }
  var ha = {};
  var K = {};
  function ia(a) {
    L(a, "add");
    L(a, "append");
    L(a, "search");
    L(a, "update");
    L(a, "remove");
  }
  function L(a, b) {
    a[b + "Async"] = function() {
      const c = this, d = arguments;
      var e = d[d.length - 1];
      let f;
      E(e) && (f = e, delete d[d.length - 1]);
      e = new Promise(function(h) {
        setTimeout(function() {
          c.async = true;
          const g = c[b].apply(c, d);
          c.async = false;
          h(g);
        });
      });
      return f ? (e.then(f), this) : e;
    };
  }
  function ja(a, b, c, d) {
    const e = a.length;
    let f = [], h, g, k = 0;
    d && (d = []);
    for (let m = e - 1; 0 <= m; m--) {
      const n = a[m], w = n.length, q = x();
      let r = !h;
      for (let l = 0; l < w; l++) {
        const p = n[l], A = p.length;
        if (A) for (let B = 0, z, y; B < A; B++) if (y = p[B], h) {
          if (h[y]) {
            if (!m) {
              if (c) c--;
              else if (f[k++] = y, k === b) return f;
            }
            if (m || d) q[y] = 1;
            r = true;
          }
          if (d && (z = (g[y] || 0) + 1, g[y] = z, z < e)) {
            const H = d[z - 2] || (d[z - 2] = []);
            H[H.length] = y;
          }
        } else q[y] = 1;
      }
      if (d) h || (g = q);
      else if (!r) return [];
      h = q;
    }
    if (d) for (let m = d.length - 1, n, w; 0 <= m; m--) {
      n = d[m];
      w = n.length;
      for (let q = 0, r; q < w; q++) if (r = n[q], !h[r]) {
        if (c) c--;
        else if (f[k++] = r, k === b) return f;
        h[r] = 1;
      }
    }
    return f;
  }
  function ka(a, b) {
    const c = x(), d = x(), e = [];
    for (let f = 0; f < a.length; f++) c[a[f]] = 1;
    for (let f = 0, h; f < b.length; f++) {
      h = b[f];
      for (let g = 0, k; g < h.length; g++) k = h[g], c[k] && !d[k] && (d[k] = 1, e[e.length] = k);
    }
    return e;
  }
  function M(a) {
    this.l = true !== a && a;
    this.cache = x();
    this.h = [];
  }
  function la(a, b, c) {
    D(a) && (a = a.query);
    let d = this.cache.get(a);
    d || (d = this.search(a, b, c), this.cache.set(a, d));
    return d;
  }
  M.prototype.set = function(a, b) {
    if (!this.cache[a]) {
      var c = this.h.length;
      c === this.l ? delete this.cache[this.h[c - 1]] : c++;
      for (--c; 0 < c; c--) this.h[c] = this.h[c - 1];
      this.h[0] = a;
    }
    this.cache[a] = b;
  };
  M.prototype.get = function(a) {
    const b = this.cache[a];
    if (this.l && b && (a = this.h.indexOf(a))) {
      const c = this.h[a - 1];
      this.h[a - 1] = this.h[a];
      this.h[a] = c;
    }
    return b;
  };
  var na = { memory: { charset: "latin:extra", D: 3, B: 4, m: false }, performance: { D: 3, B: 3, s: false, context: { depth: 2, D: 1 } }, match: { charset: "latin:extra", G: "reverse" }, score: { charset: "latin:advanced", D: 20, B: 3, context: { depth: 3, D: 9 } }, "default": {} };
  function oa(a, b, c, d, e, f, h, g) {
    setTimeout(function() {
      const k = a(c ? c + "." + d : d, JSON.stringify(h));
      k && k.then ? k.then(function() {
        b.export(a, b, c, e, f + 1, g);
      }) : b.export(a, b, c, e, f + 1, g);
    });
  }
  function N(a, b) {
    if (!(this instanceof N)) return new N(a);
    var c;
    if (a) {
      C(a) ? a = na[a] : (c = a.preset) && (a = Object.assign({}, c[c], a));
      c = a.charset;
      var d = a.lang;
      C(c) && (-1 === c.indexOf(":") && (c += ":default"), c = K[c]);
      C(d) && (d = ha[d]);
    } else a = {};
    let e, f, h = a.context || {};
    this.encode = a.encode || c && c.encode || ea;
    this.register = b || x();
    this.D = e = a.resolution || 9;
    this.G = b = c && c.G || a.tokenize || "strict";
    this.depth = "strict" === b && h.depth;
    this.l = u(h.bidirectional);
    this.s = f = u(a.optimize);
    this.m = u(a.fastupdate);
    this.B = a.minlength || 1;
    this.C = a.boost;
    this.map = f ? v(e) : x();
    this.A = e = h.resolution || 1;
    this.h = f ? v(e) : x();
    this.F = c && c.F || a.rtl;
    this.H = (b = a.matcher || d && d.H) && I(b, false);
    this.J = (b = a.stemmer || d && d.J) && I(b, true);
    if (c = b = a.filter || d && d.filter) {
      c = b;
      d = x();
      for (let g = 0, k = c.length; g < k; g++) d[c[g]] = 1;
      c = d;
    }
    this.filter = c;
    this.cache = (b = a.cache) && new M(b);
  }
  t = N.prototype;
  t.append = function(a, b) {
    return this.add(a, b, true);
  };
  t.add = function(a, b, c, d) {
    if (b && (a || 0 === a)) {
      if (!d && !c && this.register[a]) return this.update(a, b);
      b = this.encode(b);
      if (d = b.length) {
        const m = x(), n = x(), w = this.depth, q = this.D;
        for (let r = 0; r < d; r++) {
          let l = b[this.F ? d - 1 - r : r];
          var e = l.length;
          if (l && e >= this.B && (w || !n[l])) {
            var f = O(q, d, r), h = "";
            switch (this.G) {
              case "full":
                if (2 < e) {
                  for (f = 0; f < e; f++) for (var g = e; g > f; g--) if (g - f >= this.B) {
                    var k = O(q, d, r, e, f);
                    h = l.substring(f, g);
                    P(this, n, h, k, a, c);
                  }
                  break;
                }
              case "reverse":
                if (1 < e) {
                  for (g = e - 1; 0 < g; g--) h = l[g] + h, h.length >= this.B && P(
                    this,
                    n,
                    h,
                    O(q, d, r, e, g),
                    a,
                    c
                  );
                  h = "";
                }
              case "forward":
                if (1 < e) {
                  for (g = 0; g < e; g++) h += l[g], h.length >= this.B && P(this, n, h, f, a, c);
                  break;
                }
              default:
                if (this.C && (f = Math.min(f / this.C(b, l, r) | 0, q - 1)), P(this, n, l, f, a, c), w && 1 < d && r < d - 1) {
                  for (e = x(), h = this.A, f = l, g = Math.min(w + 1, d - r), e[f] = 1, k = 1; k < g; k++) if ((l = b[this.F ? d - 1 - r - k : r + k]) && l.length >= this.B && !e[l]) {
                    e[l] = 1;
                    const p = this.l && l > f;
                    P(this, m, p ? f : l, O(h + (d / 2 > h ? 0 : 1), d, r, g - 1, k - 1), a, c, p ? l : f);
                  }
                }
            }
          }
        }
        this.m || (this.register[a] = 1);
      }
    }
    return this;
  };
  function O(a, b, c, d, e) {
    return c && 1 < a ? b + (d || 0) <= a ? c + (e || 0) : (a - 1) / (b + (d || 0)) * (c + (e || 0)) + 1 | 0 : 0;
  }
  function P(a, b, c, d, e, f, h) {
    let g = h ? a.h : a.map;
    if (!b[c] || h && !b[c][h]) a.s && (g = g[d]), h ? (b = b[c] || (b[c] = x()), b[h] = 1, g = g[h] || (g[h] = x())) : b[c] = 1, g = g[c] || (g[c] = []), a.s || (g = g[d] || (g[d] = [])), f && g.includes(e) || (g[g.length] = e, a.m && (a = a.register[e] || (a.register[e] = []), a[a.length] = g));
  }
  t.search = function(a, b, c) {
    c || (!b && D(a) ? (c = a, a = c.query) : D(b) && (c = b));
    let d = [], e;
    let f, h = 0;
    if (c) {
      a = c.query || a;
      b = c.limit;
      h = c.offset || 0;
      var g = c.context;
      f = c.suggest;
    }
    if (a && (a = this.encode("" + a), e = a.length, 1 < e)) {
      c = x();
      var k = [];
      for (let n = 0, w = 0, q; n < e; n++) if ((q = a[n]) && q.length >= this.B && !c[q]) if (this.s || f || this.map[q]) k[w++] = q, c[q] = 1;
      else return d;
      a = k;
      e = a.length;
    }
    if (!e) return d;
    b || (b = 100);
    g = this.depth && 1 < e && false !== g;
    c = 0;
    let m;
    g ? (m = a[0], c = 1) : 1 < e && a.sort(aa);
    for (let n, w; c < e; c++) {
      w = a[c];
      g ? (n = pa(
        this,
        d,
        f,
        b,
        h,
        2 === e,
        w,
        m
      ), f && false === n && d.length || (m = w)) : n = pa(this, d, f, b, h, 1 === e, w);
      if (n) return n;
      if (f && c === e - 1) {
        k = d.length;
        if (!k) {
          if (g) {
            g = 0;
            c = -1;
            continue;
          }
          return d;
        }
        if (1 === k) return qa(d[0], b, h);
      }
    }
    return ja(d, b, h, f);
  };
  function pa(a, b, c, d, e, f, h, g) {
    let k = [], m = g ? a.h : a.map;
    a.s || (m = ra(m, h, g, a.l));
    if (m) {
      let n = 0;
      const w = Math.min(m.length, g ? a.A : a.D);
      for (let q = 0, r = 0, l, p; q < w; q++) if (l = m[q]) {
        if (a.s && (l = ra(l, h, g, a.l)), e && l && f && (p = l.length, p <= e ? (e -= p, l = null) : (l = l.slice(e), e = 0)), l && (k[n++] = l, f && (r += l.length, r >= d))) break;
      }
      if (n) {
        if (f) return qa(k, d, 0);
        b[b.length] = k;
        return;
      }
    }
    return !c && k;
  }
  function qa(a, b, c) {
    a = 1 === a.length ? a[0] : [].concat.apply([], a);
    return c || a.length > b ? a.slice(c, c + b) : a;
  }
  function ra(a, b, c, d) {
    c ? (d = d && b > c, a = (a = a[d ? b : c]) && a[d ? c : b]) : a = a[b];
    return a;
  }
  t.contain = function(a) {
    return !!this.register[a];
  };
  t.update = function(a, b) {
    return this.remove(a).add(a, b);
  };
  t.remove = function(a, b) {
    const c = this.register[a];
    if (c) {
      if (this.m) for (let d = 0, e; d < c.length; d++) e = c[d], e.splice(e.indexOf(a), 1);
      else Q(this.map, a, this.D, this.s), this.depth && Q(this.h, a, this.A, this.s);
      b || delete this.register[a];
      if (this.cache) {
        b = this.cache;
        for (let d = 0, e, f; d < b.h.length; d++) f = b.h[d], e = b.cache[f], e.includes(a) && (b.h.splice(d--, 1), delete b.cache[f]);
      }
    }
    return this;
  };
  function Q(a, b, c, d, e) {
    let f = 0;
    if (a.constructor === Array) if (e) b = a.indexOf(b), -1 !== b ? 1 < a.length && (a.splice(b, 1), f++) : f++;
    else {
      e = Math.min(a.length, c);
      for (let h = 0, g; h < e; h++) if (g = a[h]) f = Q(g, b, c, d, e), d || f || delete a[h];
    }
    else for (let h in a) (f = Q(a[h], b, c, d, e)) || delete a[h];
    return f;
  }
  t.searchCache = la;
  t.export = function(a, b, c, d, e, f) {
    let h = true;
    "undefined" === typeof f && (h = new Promise((m) => {
      f = m;
    }));
    let g, k;
    switch (e || (e = 0)) {
      case 0:
        g = "reg";
        if (this.m) {
          k = x();
          for (let m in this.register) k[m] = 1;
        } else k = this.register;
        break;
      case 1:
        g = "cfg";
        k = { doc: 0, opt: this.s ? 1 : 0 };
        break;
      case 2:
        g = "map";
        k = this.map;
        break;
      case 3:
        g = "ctx";
        k = this.h;
        break;
      default:
        "undefined" === typeof c && f && f();
        return;
    }
    oa(a, b || this, c, g, d, e, k, f);
    return h;
  };
  t.import = function(a, b) {
    if (b) switch (C(b) && (b = JSON.parse(b)), a) {
      case "cfg":
        this.s = !!b.opt;
        break;
      case "reg":
        this.m = false;
        this.register = b;
        break;
      case "map":
        this.map = b;
        break;
      case "ctx":
        this.h = b;
    }
  };
  ia(N.prototype);
  function sa(a) {
    a = a.data;
    var b = self._index;
    const c = a.args;
    var d = a.task;
    switch (d) {
      case "init":
        d = a.options || {};
        a = a.factory;
        b = d.encode;
        d.cache = false;
        b && 0 === b.indexOf("function") && (d.encode = Function("return " + b)());
        a ? (Function("return " + a)()(self), self._index = new self.FlexSearch.Index(d), delete self.FlexSearch) : self._index = new N(d);
        break;
      default:
        a = a.id, b = b[d].apply(b, c), postMessage("search" === d ? { id: a, msg: b } : { id: a });
    }
  }
  var ta = 0;
  function S(a) {
    if (!(this instanceof S)) return new S(a);
    var b;
    a ? E(b = a.encode) && (a.encode = b.toString()) : a = {};
    (b = (self || window)._factory) && (b = b.toString());
    const c = "undefined" === typeof window && self.exports, d = this;
    this.o = ua(b, c, a.worker);
    this.h = x();
    if (this.o) {
      if (c) this.o.on("message", function(e) {
        d.h[e.id](e.msg);
        delete d.h[e.id];
      });
      else this.o.onmessage = function(e) {
        e = e.data;
        d.h[e.id](e.msg);
        delete d.h[e.id];
      };
      this.o.postMessage({ task: "init", factory: b, options: a });
    }
  }
  T("add");
  T("append");
  T("search");
  T("update");
  T("remove");
  function T(a) {
    S.prototype[a] = S.prototype[a + "Async"] = function() {
      const b = this, c = [].slice.call(arguments);
      var d = c[c.length - 1];
      let e;
      E(d) && (e = d, c.splice(c.length - 1, 1));
      d = new Promise(function(f) {
        setTimeout(function() {
          b.h[++ta] = f;
          b.o.postMessage({ task: a, id: ta, args: c });
        });
      });
      return e ? (d.then(e), this) : d;
    };
  }
  function ua(a, b, c) {
    let d;
    try {
      d = b ? new (require_worker_threads())["Worker"](__dirname + "/node/node.js") : a ? new Worker(URL.createObjectURL(new Blob(["onmessage=" + sa.toString()], { type: "text/javascript" }))) : new Worker(C(c) ? c : "worker/worker.js", { type: "module" });
    } catch (e) {
    }
    return d;
  }
  function U(a) {
    if (!(this instanceof U)) return new U(a);
    var b = a.document || a.doc || a, c;
    this.K = [];
    this.h = [];
    this.A = [];
    this.register = x();
    this.key = (c = b.key || b.id) && V(c, this.A) || "id";
    this.m = u(a.fastupdate);
    this.C = (c = b.store) && true !== c && [];
    this.store = c && x();
    this.I = (c = b.tag) && V(c, this.A);
    this.l = c && x();
    this.cache = (c = a.cache) && new M(c);
    a.cache = false;
    this.o = a.worker;
    this.async = false;
    c = x();
    let d = b.index || b.field || b;
    C(d) && (d = [d]);
    for (let e = 0, f, h; e < d.length; e++) f = d[e], C(f) || (h = f, f = f.field), h = D(h) ? Object.assign({}, a, h) : a, this.o && (c[f] = new S(h), c[f].o || (this.o = false)), this.o || (c[f] = new N(h, this.register)), this.K[e] = V(f, this.A), this.h[e] = f;
    if (this.C) for (a = b.store, C(a) && (a = [a]), b = 0; b < a.length; b++) this.C[b] = V(a[b], this.A);
    this.index = c;
  }
  function V(a, b) {
    const c = a.split(":");
    let d = 0;
    for (let e = 0; e < c.length; e++) a = c[e], 0 <= a.indexOf("[]") && (a = a.substring(0, a.length - 2)) && (b[d] = true), a && (c[d++] = a);
    d < c.length && (c.length = d);
    return 1 < d ? c : c[0];
  }
  function X(a, b) {
    if (C(b)) a = a[b];
    else for (let c = 0; a && c < b.length; c++) a = a[b[c]];
    return a;
  }
  function Y(a, b, c, d, e) {
    a = a[e];
    if (d === c.length - 1) b[e] = a;
    else if (a) if (a.constructor === Array) for (b = b[e] = Array(a.length), e = 0; e < a.length; e++) Y(a, b, c, d, e);
    else b = b[e] || (b[e] = x()), e = c[++d], Y(a, b, c, d, e);
  }
  function Z(a, b, c, d, e, f, h, g) {
    if (a = a[h]) if (d === b.length - 1) {
      if (a.constructor === Array) {
        if (c[d]) {
          for (b = 0; b < a.length; b++) e.add(f, a[b], true, true);
          return;
        }
        a = a.join(" ");
      }
      e.add(f, a, g, true);
    } else if (a.constructor === Array) for (h = 0; h < a.length; h++) Z(a, b, c, d, e, f, h, g);
    else h = b[++d], Z(a, b, c, d, e, f, h, g);
  }
  t = U.prototype;
  t.add = function(a, b, c) {
    D(a) && (b = a, a = X(b, this.key));
    if (b && (a || 0 === a)) {
      if (!c && this.register[a]) return this.update(a, b);
      for (let d = 0, e, f; d < this.h.length; d++) f = this.h[d], e = this.K[d], C(e) && (e = [e]), Z(b, e, this.A, 0, this.index[f], a, e[0], c);
      if (this.I) {
        let d = X(b, this.I), e = x();
        C(d) && (d = [d]);
        for (let f = 0, h, g; f < d.length; f++) if (h = d[f], !e[h] && (e[h] = 1, g = this.l[h] || (this.l[h] = []), !c || !g.includes(a))) {
          if (g[g.length] = a, this.m) {
            const k = this.register[a] || (this.register[a] = []);
            k[k.length] = g;
          }
        }
      }
      if (this.store && (!c || !this.store[a])) {
        let d;
        if (this.C) {
          d = x();
          for (let e = 0, f; e < this.C.length; e++) f = this.C[e], C(f) ? d[f] = b[f] : Y(b, d, f, 0, f[0]);
        }
        this.store[a] = d || b;
      }
    }
    return this;
  };
  t.append = function(a, b) {
    return this.add(a, b, true);
  };
  t.update = function(a, b) {
    return this.remove(a).add(a, b);
  };
  t.remove = function(a) {
    D(a) && (a = X(a, this.key));
    if (this.register[a]) {
      for (var b = 0; b < this.h.length && (this.index[this.h[b]].remove(a, !this.o), !this.m); b++) ;
      if (this.I && !this.m) for (let c in this.l) {
        b = this.l[c];
        const d = b.indexOf(a);
        -1 !== d && (1 < b.length ? b.splice(d, 1) : delete this.l[c]);
      }
      this.store && delete this.store[a];
      delete this.register[a];
    }
    return this;
  };
  t.search = function(a, b, c, d) {
    c || (!b && D(a) ? (c = a, a = "") : D(b) && (c = b, b = 0));
    let e = [], f = [], h, g, k, m, n, w, q = 0;
    if (c) if (c.constructor === Array) k = c, c = null;
    else {
      a = c.query || a;
      k = (h = c.pluck) || c.index || c.field;
      m = c.tag;
      g = this.store && c.enrich;
      n = "and" === c.bool;
      b = c.limit || b || 100;
      w = c.offset || 0;
      if (m && (C(m) && (m = [m]), !a)) {
        for (let l = 0, p; l < m.length; l++) if (p = va.call(this, m[l], b, w, g)) e[e.length] = p, q++;
        return q ? e : [];
      }
      C(k) && (k = [k]);
    }
    k || (k = this.h);
    n = n && (1 < k.length || m && 1 < m.length);
    const r = !d && (this.o || this.async) && [];
    for (let l = 0, p, A, B; l < k.length; l++) {
      let z;
      A = k[l];
      C(A) || (z = A, A = z.field, a = z.query || a, b = z.limit || b, g = z.enrich || g);
      if (r) r[l] = this.index[A].searchAsync(a, b, z || c);
      else {
        d ? p = d[l] : p = this.index[A].search(a, b, z || c);
        B = p && p.length;
        if (m && B) {
          const y = [];
          let H = 0;
          n && (y[0] = [p]);
          for (let W = 0, ma, R; W < m.length; W++) if (ma = m[W], B = (R = this.l[ma]) && R.length) H++, y[y.length] = n ? [R] : R;
          H && (p = n ? ja(y, b || 100, w || 0) : ka(p, y), B = p.length);
        }
        if (B) f[q] = A, e[q++] = p;
        else if (n) return [];
      }
    }
    if (r) {
      const l = this;
      return new Promise(function(p) {
        Promise.all(r).then(function(A) {
          p(l.search(
            a,
            b,
            c,
            A
          ));
        });
      });
    }
    if (!q) return [];
    if (h && (!g || !this.store)) return e[0];
    for (let l = 0, p; l < f.length; l++) {
      p = e[l];
      p.length && g && (p = wa.call(this, p));
      if (h) return p;
      e[l] = { field: f[l], result: p };
    }
    return e;
  };
  function va(a, b, c, d) {
    let e = this.l[a], f = e && e.length - c;
    if (f && 0 < f) {
      if (f > b || c) e = e.slice(c, c + b);
      d && (e = wa.call(this, e));
      return { tag: a, result: e };
    }
  }
  function wa(a) {
    const b = Array(a.length);
    for (let c = 0, d; c < a.length; c++) d = a[c], b[c] = { id: d, doc: this.store[d] };
    return b;
  }
  t.contain = function(a) {
    return !!this.register[a];
  };
  t.get = function(a) {
    return this.store[a];
  };
  t.set = function(a, b) {
    this.store[a] = b;
    return this;
  };
  t.searchCache = la;
  t.export = function(a, b, c, d, e, f) {
    let h;
    "undefined" === typeof f && (h = new Promise((g) => {
      f = g;
    }));
    e || (e = 0);
    d || (d = 0);
    if (d < this.h.length) {
      const g = this.h[d], k = this.index[g];
      b = this;
      setTimeout(function() {
        k.export(a, b, e ? g : "", d, e++, f) || (d++, e = 1, b.export(a, b, g, d, e, f));
      });
    } else {
      let g, k;
      switch (e) {
        case 1:
          g = "tag";
          k = this.l;
          c = null;
          break;
        case 2:
          g = "store";
          k = this.store;
          c = null;
          break;
        default:
          f();
          return;
      }
      oa(a, this, c, g, d, e, k, f);
    }
    return h;
  };
  t.import = function(a, b) {
    if (b) switch (C(b) && (b = JSON.parse(b)), a) {
      case "tag":
        this.l = b;
        break;
      case "reg":
        this.m = false;
        this.register = b;
        for (let d = 0, e; d < this.h.length; d++) e = this.index[this.h[d]], e.register = b, e.m = false;
        break;
      case "store":
        this.store = b;
        break;
      default:
        a = a.split(".");
        const c = a[0];
        a = a[1];
        c && a && this.index[c].import(a, b);
    }
  };
  ia(U.prototype);
  var ya = { encode: xa, F: false, G: "" };
  var za = [J("[\xE0\xE1\xE2\xE3\xE4\xE5]"), "a", J("[\xE8\xE9\xEA\xEB]"), "e", J("[\xEC\xED\xEE\xEF]"), "i", J("[\xF2\xF3\xF4\xF5\xF6\u0151]"), "o", J("[\xF9\xFA\xFB\xFC\u0171]"), "u", J("[\xFD\u0177\xFF]"), "y", J("\xF1"), "n", J("[\xE7c]"), "k", J("\xDF"), "s", J(" & "), " and "];
  function xa(a) {
    var b = a = "" + a;
    b.normalize && (b = b.normalize("NFD").replace(ca, ""));
    return F.call(this, b.toLowerCase(), !a.normalize && za);
  }
  var Ba = { encode: Aa, F: false, G: "strict" };
  var Ca = /[^a-z0-9]+/;
  var Da = { b: "p", v: "f", w: "f", z: "s", x: "s", "\xDF": "s", d: "t", n: "m", c: "k", g: "k", j: "k", q: "k", i: "e", y: "e", u: "o" };
  function Aa(a) {
    a = xa.call(this, a).join(" ");
    const b = [];
    if (a) {
      const c = a.split(Ca), d = c.length;
      for (let e = 0, f, h = 0; e < d; e++) if ((a = c[e]) && (!this.filter || !this.filter[a])) {
        f = a[0];
        let g = Da[f] || f, k = g;
        for (let m = 1; m < a.length; m++) {
          f = a[m];
          const n = Da[f] || f;
          n && n !== k && (g += n, k = n);
        }
        b[h++] = g;
      }
    }
    return b;
  }
  var Fa = { encode: Ea, F: false, G: "" };
  var Ga = [J("ae"), "a", J("oe"), "o", J("sh"), "s", J("th"), "t", J("ph"), "f", J("pf"), "f", J("(?![aeo])h(?![aeo])"), "", J("(?!^[aeo])h(?!^[aeo])"), ""];
  function Ea(a, b) {
    a && (a = Aa.call(this, a).join(" "), 2 < a.length && (a = G(a, Ga)), b || (1 < a.length && (a = da(a)), a && (a = a.split(" "))));
    return a || [];
  }
  var Ia = { encode: Ha, F: false, G: "" };
  var Ja = J("(?!\\b)[aeo]");
  function Ha(a) {
    a && (a = Ea.call(this, a, true), 1 < a.length && (a = a.replace(Ja, "")), 1 < a.length && (a = da(a)), a && (a = a.split(" ")));
    return a || [];
  }
  K["latin:default"] = fa;
  K["latin:simple"] = ya;
  K["latin:balance"] = Ba;
  K["latin:advanced"] = Fa;
  K["latin:extra"] = Ia;
  var flexsearch_bundle_module_min_default = { Index: N, Document: U, Worker: S, registerCharset: function(a, b) {
    K[a] = b;
  }, registerLanguage: function(a, b) {
    ha[a] = b;
  } };

  // <stdin>
  (function() {
    "use strict";
    const index = new flexsearch_bundle_module_min_default.Document({
      tokenize: "forward",
      document: {
        id: "id",
        index: [
          {
            field: "title"
          },
          {
            field: "tags"
          },
          {
            field: "content"
          },
          {
            field: "date",
            tokenize: "strict",
            encode: false
          }
        ],
        store: ["title", "summary", "date", "permalink"]
      }
    });
    function showResults(items) {
      const template = document.querySelector("template").content;
      const fragment = document.createDocumentFragment();
      const results = document.querySelector(".search-results");
      results.textContent = "";
      const itemsLength = Object.keys(items).length;
      if (itemsLength === 0 && query.value === "") {
        document.querySelector(".search-no-results").classList.add("d-none");
        document.querySelector(".search-no-recent").classList.remove("d-none");
      } else if (itemsLength === 0 && query.value !== "") {
        document.querySelector(".search-no-recent").classList.add("d-none");
        const queryNoResults = document.querySelector(".query-no-results");
        queryNoResults.innerText = query.value;
        document.querySelector(".search-no-results").classList.remove("d-none");
      } else {
        document.querySelector(".search-no-recent").classList.add("d-none");
        document.querySelector(".search-no-results").classList.add("d-none");
      }
      for (const id in items) {
        const item = items[id];
        const result = template.cloneNode(true);
        const a = result.querySelector("a");
        const time = result.querySelector("time");
        const content = result.querySelector(".content");
        a.innerHTML = item.title;
        a.href = item.permalink;
        time.innerText = item.date;
        content.innerHTML = item.summary;
        fragment.appendChild(result);
      }
      results.appendChild(fragment);
    }
    function doSearch() {
      const query2 = document.querySelector(".search-text").value.trim();
      const limit = 99;
      const results = index.search({
        query: query2,
        enrich: true,
        limit
      });
      const items = {};
      results.forEach(function(result) {
        result.result.forEach(function(r) {
          items[r.id] = r.doc;
        });
      });
      showResults(items);
    }
    function enableUI() {
      const searchform = document.querySelector(".search-form");
      searchform.addEventListener("submit", function(e) {
        e.preventDefault();
        doSearch();
      });
      searchform.addEventListener("input", function() {
        doSearch();
      });
      document.querySelector(".search-loading").classList.add("d-none");
      document.querySelector(".search-input").classList.remove("d-none");
      document.querySelector(".search-text").focus();
    }
    function buildIndex() {
      document.querySelector(".search-loading").classList.remove("d-none");
      fetch("/gdg/search-index.json").then(function(response) {
        return response.json();
      }).then(function(data) {
        data.forEach(function(item) {
          index.add(item);
        });
      });
    }
    buildIndex();
    enableUI();
  })();
})();
/*!
 * FlexSearch for Bootstrap based Thulite sites
 * Copyright 2021-2024 Thulite
 * Licensed under the MIT License
 * Based on https://github.com/frjo/hugo-theme-zen/blob/main/assets/js/search.js
 */
//# sourceMappingURL=data:application/json;base64,ewogICJ2ZXJzaW9uIjogMywKICAic291cmNlcyI6IFsibm9kZV9tb2R1bGVzL2ZsZXhzZWFyY2gvZGlzdC9mbGV4c2VhcmNoLmJ1bmRsZS5tb2R1bGUubWluLmpzIiwgIjxzdGRpbj4iXSwKICAic291cmNlc0NvbnRlbnQiOiBbIi8qKiFcclxuICogRmxleFNlYXJjaC5qcyB2MC43LjQxIChCdW5kbGUubW9kdWxlKVxyXG4gKiBBdXRob3IgYW5kIENvcHlyaWdodDogVGhvbWFzIFdpbGtlcmxpbmdcclxuICogTGljZW5jZTogQXBhY2hlLTIuMFxyXG4gKiBIb3N0ZWQgYnkgTmV4dGFwcHMgR21iSFxyXG4gKiBodHRwczovL2dpdGh1Yi5jb20vbmV4dGFwcHMtZGUvZmxleHNlYXJjaFxyXG4gKi9cbnZhciB0O2Z1bmN0aW9uIHUoYSl7cmV0dXJuXCJ1bmRlZmluZWRcIiE9PXR5cGVvZiBhP2E6ITB9ZnVuY3Rpb24gdihhKXtjb25zdCBiPUFycmF5KGEpO2ZvcihsZXQgYz0wO2M8YTtjKyspYltjXT14KCk7cmV0dXJuIGJ9ZnVuY3Rpb24geCgpe3JldHVybiBPYmplY3QuY3JlYXRlKG51bGwpfWZ1bmN0aW9uIGFhKGEsYil7cmV0dXJuIGIubGVuZ3RoLWEubGVuZ3RofWZ1bmN0aW9uIEMoYSl7cmV0dXJuXCJzdHJpbmdcIj09PXR5cGVvZiBhfWZ1bmN0aW9uIEQoYSl7cmV0dXJuXCJvYmplY3RcIj09PXR5cGVvZiBhfWZ1bmN0aW9uIEUoYSl7cmV0dXJuXCJmdW5jdGlvblwiPT09dHlwZW9mIGF9O2Z1bmN0aW9uIEYoYSxiKXt2YXIgYz1iYTtpZihhJiYoYiYmKGE9RyhhLGIpKSx0aGlzLkgmJihhPUcoYSx0aGlzLkgpKSx0aGlzLkomJjE8YS5sZW5ndGgmJihhPUcoYSx0aGlzLkopKSxjfHxcIlwiPT09Yykpe2I9YS5zcGxpdChjKTtpZih0aGlzLmZpbHRlcil7YT10aGlzLmZpbHRlcjtjPWIubGVuZ3RoO2NvbnN0IGQ9W107Zm9yKGxldCBlPTAsZj0wO2U8YztlKyspe2NvbnN0IGg9YltlXTtoJiYhYVtoXSYmKGRbZisrXT1oKX1hPWR9ZWxzZSBhPWI7cmV0dXJuIGF9cmV0dXJuIGF9Y29uc3QgYmE9L1tcXHB7Wn1cXHB7U31cXHB7UH1cXHB7Q31dKy91LGNhPS9bXFx1MDMwMC1cXHUwMzZmXS9nO1xuZnVuY3Rpb24gSShhLGIpe2NvbnN0IGM9T2JqZWN0LmtleXMoYSksZD1jLmxlbmd0aCxlPVtdO2xldCBmPVwiXCIsaD0wO2ZvcihsZXQgZz0wLGssbTtnPGQ7ZysrKWs9Y1tnXSwobT1hW2tdKT8oZVtoKytdPUooYj9cIig/IVxcXFxiKVwiK2srXCIoXFxcXGJ8XylcIjprKSxlW2grK109bSk6Zis9KGY/XCJ8XCI6XCJcIikraztmJiYoZVtoKytdPUooYj9cIig/IVxcXFxiKShcIitmK1wiKShcXFxcYnxfKVwiOlwiKFwiK2YrXCIpXCIpLGVbaF09XCJcIik7cmV0dXJuIGV9ZnVuY3Rpb24gRyhhLGIpe2ZvcihsZXQgYz0wLGQ9Yi5sZW5ndGg7YzxkJiYoYT1hLnJlcGxhY2UoYltjXSxiW2MrMV0pLGEpO2MrPTIpO3JldHVybiBhfWZ1bmN0aW9uIEooYSl7cmV0dXJuIG5ldyBSZWdFeHAoYSxcImdcIil9ZnVuY3Rpb24gZGEoYSl7bGV0IGI9XCJcIixjPVwiXCI7Zm9yKGxldCBkPTAsZT1hLmxlbmd0aCxmO2Q8ZTtkKyspKGY9YVtkXSkhPT1jJiYoYis9Yz1mKTtyZXR1cm4gYn07dmFyIGZhPXtlbmNvZGU6ZWEsRjohMSxHOlwiXCJ9O2Z1bmN0aW9uIGVhKGEpe3JldHVybiBGLmNhbGwodGhpcywoXCJcIithKS50b0xvd2VyQ2FzZSgpLCExKX07Y29uc3QgaGE9e30sSz17fTtmdW5jdGlvbiBpYShhKXtMKGEsXCJhZGRcIik7TChhLFwiYXBwZW5kXCIpO0woYSxcInNlYXJjaFwiKTtMKGEsXCJ1cGRhdGVcIik7TChhLFwicmVtb3ZlXCIpfWZ1bmN0aW9uIEwoYSxiKXthW2IrXCJBc3luY1wiXT1mdW5jdGlvbigpe2NvbnN0IGM9dGhpcyxkPWFyZ3VtZW50czt2YXIgZT1kW2QubGVuZ3RoLTFdO2xldCBmO0UoZSkmJihmPWUsZGVsZXRlIGRbZC5sZW5ndGgtMV0pO2U9bmV3IFByb21pc2UoZnVuY3Rpb24oaCl7c2V0VGltZW91dChmdW5jdGlvbigpe2MuYXN5bmM9ITA7Y29uc3QgZz1jW2JdLmFwcGx5KGMsZCk7Yy5hc3luYz0hMTtoKGcpfSl9KTtyZXR1cm4gZj8oZS50aGVuKGYpLHRoaXMpOmV9fTtmdW5jdGlvbiBqYShhLGIsYyxkKXtjb25zdCBlPWEubGVuZ3RoO2xldCBmPVtdLGgsZyxrPTA7ZCYmKGQ9W10pO2ZvcihsZXQgbT1lLTE7MDw9bTttLS0pe2NvbnN0IG49YVttXSx3PW4ubGVuZ3RoLHE9eCgpO2xldCByPSFoO2ZvcihsZXQgbD0wO2w8dztsKyspe2NvbnN0IHA9bltsXSxBPXAubGVuZ3RoO2lmKEEpZm9yKGxldCBCPTAseix5O0I8QTtCKyspaWYoeT1wW0JdLGgpe2lmKGhbeV0pe2lmKCFtKWlmKGMpYy0tO2Vsc2UgaWYoZltrKytdPXksaz09PWIpcmV0dXJuIGY7aWYobXx8ZClxW3ldPTE7cj0hMH1pZihkJiYoej0oZ1t5XXx8MCkrMSxnW3ldPXosejxlKSl7Y29uc3QgSD1kW3otMl18fChkW3otMl09W10pO0hbSC5sZW5ndGhdPXl9fWVsc2UgcVt5XT0xfWlmKGQpaHx8KGc9cSk7ZWxzZSBpZighcilyZXR1cm5bXTtoPXF9aWYoZClmb3IobGV0IG09ZC5sZW5ndGgtMSxuLHc7MDw9bTttLS0pe249ZFttXTt3PW4ubGVuZ3RoO2ZvcihsZXQgcT0wLHI7cTx3O3ErKylpZihyPVxubltxXSwhaFtyXSl7aWYoYyljLS07ZWxzZSBpZihmW2srK109cixrPT09YilyZXR1cm4gZjtoW3JdPTF9fXJldHVybiBmfWZ1bmN0aW9uIGthKGEsYil7Y29uc3QgYz14KCksZD14KCksZT1bXTtmb3IobGV0IGY9MDtmPGEubGVuZ3RoO2YrKyljW2FbZl1dPTE7Zm9yKGxldCBmPTAsaDtmPGIubGVuZ3RoO2YrKyl7aD1iW2ZdO2ZvcihsZXQgZz0wLGs7ZzxoLmxlbmd0aDtnKyspaz1oW2ddLGNba10mJiFkW2tdJiYoZFtrXT0xLGVbZS5sZW5ndGhdPWspfXJldHVybiBlfTtmdW5jdGlvbiBNKGEpe3RoaXMubD0hMCE9PWEmJmE7dGhpcy5jYWNoZT14KCk7dGhpcy5oPVtdfWZ1bmN0aW9uIGxhKGEsYixjKXtEKGEpJiYoYT1hLnF1ZXJ5KTtsZXQgZD10aGlzLmNhY2hlLmdldChhKTtkfHwoZD10aGlzLnNlYXJjaChhLGIsYyksdGhpcy5jYWNoZS5zZXQoYSxkKSk7cmV0dXJuIGR9TS5wcm90b3R5cGUuc2V0PWZ1bmN0aW9uKGEsYil7aWYoIXRoaXMuY2FjaGVbYV0pe3ZhciBjPXRoaXMuaC5sZW5ndGg7Yz09PXRoaXMubD9kZWxldGUgdGhpcy5jYWNoZVt0aGlzLmhbYy0xXV06YysrO2ZvcigtLWM7MDxjO2MtLSl0aGlzLmhbY109dGhpcy5oW2MtMV07dGhpcy5oWzBdPWF9dGhpcy5jYWNoZVthXT1ifTtNLnByb3RvdHlwZS5nZXQ9ZnVuY3Rpb24oYSl7Y29uc3QgYj10aGlzLmNhY2hlW2FdO2lmKHRoaXMubCYmYiYmKGE9dGhpcy5oLmluZGV4T2YoYSkpKXtjb25zdCBjPXRoaXMuaFthLTFdO3RoaXMuaFthLTFdPXRoaXMuaFthXTt0aGlzLmhbYV09Y31yZXR1cm4gYn07Y29uc3QgbmE9e21lbW9yeTp7Y2hhcnNldDpcImxhdGluOmV4dHJhXCIsRDozLEI6NCxtOiExfSxwZXJmb3JtYW5jZTp7RDozLEI6MyxzOiExLGNvbnRleHQ6e2RlcHRoOjIsRDoxfX0sbWF0Y2g6e2NoYXJzZXQ6XCJsYXRpbjpleHRyYVwiLEc6XCJyZXZlcnNlXCJ9LHNjb3JlOntjaGFyc2V0OlwibGF0aW46YWR2YW5jZWRcIixEOjIwLEI6Myxjb250ZXh0OntkZXB0aDozLEQ6OX19LFwiZGVmYXVsdFwiOnt9fTtmdW5jdGlvbiBvYShhLGIsYyxkLGUsZixoLGcpe3NldFRpbWVvdXQoZnVuY3Rpb24oKXtjb25zdCBrPWEoYz9jK1wiLlwiK2Q6ZCxKU09OLnN0cmluZ2lmeShoKSk7ayYmay50aGVuP2sudGhlbihmdW5jdGlvbigpe2IuZXhwb3J0KGEsYixjLGUsZisxLGcpfSk6Yi5leHBvcnQoYSxiLGMsZSxmKzEsZyl9KX07ZnVuY3Rpb24gTihhLGIpe2lmKCEodGhpcyBpbnN0YW5jZW9mIE4pKXJldHVybiBuZXcgTihhKTt2YXIgYztpZihhKXtDKGEpP2E9bmFbYV06KGM9YS5wcmVzZXQpJiYoYT1PYmplY3QuYXNzaWduKHt9LGNbY10sYSkpO2M9YS5jaGFyc2V0O3ZhciBkPWEubGFuZztDKGMpJiYoLTE9PT1jLmluZGV4T2YoXCI6XCIpJiYoYys9XCI6ZGVmYXVsdFwiKSxjPUtbY10pO0MoZCkmJihkPWhhW2RdKX1lbHNlIGE9e307bGV0IGUsZixoPWEuY29udGV4dHx8e307dGhpcy5lbmNvZGU9YS5lbmNvZGV8fGMmJmMuZW5jb2RlfHxlYTt0aGlzLnJlZ2lzdGVyPWJ8fHgoKTt0aGlzLkQ9ZT1hLnJlc29sdXRpb258fDk7dGhpcy5HPWI9YyYmYy5HfHxhLnRva2VuaXplfHxcInN0cmljdFwiO3RoaXMuZGVwdGg9XCJzdHJpY3RcIj09PWImJmguZGVwdGg7dGhpcy5sPXUoaC5iaWRpcmVjdGlvbmFsKTt0aGlzLnM9Zj11KGEub3B0aW1pemUpO3RoaXMubT11KGEuZmFzdHVwZGF0ZSk7dGhpcy5CPWEubWlubGVuZ3RofHwxO3RoaXMuQz1cbmEuYm9vc3Q7dGhpcy5tYXA9Zj92KGUpOngoKTt0aGlzLkE9ZT1oLnJlc29sdXRpb258fDE7dGhpcy5oPWY/dihlKTp4KCk7dGhpcy5GPWMmJmMuRnx8YS5ydGw7dGhpcy5IPShiPWEubWF0Y2hlcnx8ZCYmZC5IKSYmSShiLCExKTt0aGlzLko9KGI9YS5zdGVtbWVyfHxkJiZkLkopJiZJKGIsITApO2lmKGM9Yj1hLmZpbHRlcnx8ZCYmZC5maWx0ZXIpe2M9YjtkPXgoKTtmb3IobGV0IGc9MCxrPWMubGVuZ3RoO2c8aztnKyspZFtjW2ddXT0xO2M9ZH10aGlzLmZpbHRlcj1jO3RoaXMuY2FjaGU9KGI9YS5jYWNoZSkmJm5ldyBNKGIpfXQ9Ti5wcm90b3R5cGU7dC5hcHBlbmQ9ZnVuY3Rpb24oYSxiKXtyZXR1cm4gdGhpcy5hZGQoYSxiLCEwKX07XG50LmFkZD1mdW5jdGlvbihhLGIsYyxkKXtpZihiJiYoYXx8MD09PWEpKXtpZighZCYmIWMmJnRoaXMucmVnaXN0ZXJbYV0pcmV0dXJuIHRoaXMudXBkYXRlKGEsYik7Yj10aGlzLmVuY29kZShiKTtpZihkPWIubGVuZ3RoKXtjb25zdCBtPXgoKSxuPXgoKSx3PXRoaXMuZGVwdGgscT10aGlzLkQ7Zm9yKGxldCByPTA7cjxkO3IrKyl7bGV0IGw9Ylt0aGlzLkY/ZC0xLXI6cl07dmFyIGU9bC5sZW5ndGg7aWYobCYmZT49dGhpcy5CJiYod3x8IW5bbF0pKXt2YXIgZj1PKHEsZCxyKSxoPVwiXCI7c3dpdGNoKHRoaXMuRyl7Y2FzZSBcImZ1bGxcIjppZigyPGUpe2ZvcihmPTA7ZjxlO2YrKylmb3IodmFyIGc9ZTtnPmY7Zy0tKWlmKGctZj49dGhpcy5CKXt2YXIgaz1PKHEsZCxyLGUsZik7aD1sLnN1YnN0cmluZyhmLGcpO1AodGhpcyxuLGgsayxhLGMpfWJyZWFrfWNhc2UgXCJyZXZlcnNlXCI6aWYoMTxlKXtmb3IoZz1lLTE7MDxnO2ctLSloPWxbZ10raCxoLmxlbmd0aD49dGhpcy5CJiZQKHRoaXMsbixcbmgsTyhxLGQscixlLGcpLGEsYyk7aD1cIlwifWNhc2UgXCJmb3J3YXJkXCI6aWYoMTxlKXtmb3IoZz0wO2c8ZTtnKyspaCs9bFtnXSxoLmxlbmd0aD49dGhpcy5CJiZQKHRoaXMsbixoLGYsYSxjKTticmVha31kZWZhdWx0OmlmKHRoaXMuQyYmKGY9TWF0aC5taW4oZi90aGlzLkMoYixsLHIpfDAscS0xKSksUCh0aGlzLG4sbCxmLGEsYyksdyYmMTxkJiZyPGQtMSlmb3IoZT14KCksaD10aGlzLkEsZj1sLGc9TWF0aC5taW4odysxLGQtciksZVtmXT0xLGs9MTtrPGc7aysrKWlmKChsPWJbdGhpcy5GP2QtMS1yLWs6citrXSkmJmwubGVuZ3RoPj10aGlzLkImJiFlW2xdKXtlW2xdPTE7Y29uc3QgcD10aGlzLmwmJmw+ZjtQKHRoaXMsbSxwP2Y6bCxPKGgrKGQvMj5oPzA6MSksZCxyLGctMSxrLTEpLGEsYyxwP2w6Zil9fX19dGhpcy5tfHwodGhpcy5yZWdpc3RlclthXT0xKX19cmV0dXJuIHRoaXN9O1xuZnVuY3Rpb24gTyhhLGIsYyxkLGUpe3JldHVybiBjJiYxPGE/YisoZHx8MCk8PWE/YysoZXx8MCk6KGEtMSkvKGIrKGR8fDApKSooYysoZXx8MCkpKzF8MDowfWZ1bmN0aW9uIFAoYSxiLGMsZCxlLGYsaCl7bGV0IGc9aD9hLmg6YS5tYXA7aWYoIWJbY118fGgmJiFiW2NdW2hdKWEucyYmKGc9Z1tkXSksaD8oYj1iW2NdfHwoYltjXT14KCkpLGJbaF09MSxnPWdbaF18fChnW2hdPXgoKSkpOmJbY109MSxnPWdbY118fChnW2NdPVtdKSxhLnN8fChnPWdbZF18fChnW2RdPVtdKSksZiYmZy5pbmNsdWRlcyhlKXx8KGdbZy5sZW5ndGhdPWUsYS5tJiYoYT1hLnJlZ2lzdGVyW2VdfHwoYS5yZWdpc3RlcltlXT1bXSksYVthLmxlbmd0aF09ZykpfVxudC5zZWFyY2g9ZnVuY3Rpb24oYSxiLGMpe2N8fCghYiYmRChhKT8oYz1hLGE9Yy5xdWVyeSk6RChiKSYmKGM9YikpO2xldCBkPVtdLGU7bGV0IGYsaD0wO2lmKGMpe2E9Yy5xdWVyeXx8YTtiPWMubGltaXQ7aD1jLm9mZnNldHx8MDt2YXIgZz1jLmNvbnRleHQ7Zj1jLnN1Z2dlc3R9aWYoYSYmKGE9dGhpcy5lbmNvZGUoXCJcIithKSxlPWEubGVuZ3RoLDE8ZSkpe2M9eCgpO3ZhciBrPVtdO2ZvcihsZXQgbj0wLHc9MCxxO248ZTtuKyspaWYoKHE9YVtuXSkmJnEubGVuZ3RoPj10aGlzLkImJiFjW3FdKWlmKHRoaXMuc3x8Znx8dGhpcy5tYXBbcV0pa1t3KytdPXEsY1txXT0xO2Vsc2UgcmV0dXJuIGQ7YT1rO2U9YS5sZW5ndGh9aWYoIWUpcmV0dXJuIGQ7Ynx8KGI9MTAwKTtnPXRoaXMuZGVwdGgmJjE8ZSYmITEhPT1nO2M9MDtsZXQgbTtnPyhtPWFbMF0sYz0xKToxPGUmJmEuc29ydChhYSk7Zm9yKGxldCBuLHc7YzxlO2MrKyl7dz1hW2NdO2c/KG49cGEodGhpcyxkLGYsYixoLDI9PT1lLHcsXG5tKSxmJiYhMT09PW4mJmQubGVuZ3RofHwobT13KSk6bj1wYSh0aGlzLGQsZixiLGgsMT09PWUsdyk7aWYobilyZXR1cm4gbjtpZihmJiZjPT09ZS0xKXtrPWQubGVuZ3RoO2lmKCFrKXtpZihnKXtnPTA7Yz0tMTtjb250aW51ZX1yZXR1cm4gZH1pZigxPT09aylyZXR1cm4gcWEoZFswXSxiLGgpfX1yZXR1cm4gamEoZCxiLGgsZil9O1xuZnVuY3Rpb24gcGEoYSxiLGMsZCxlLGYsaCxnKXtsZXQgaz1bXSxtPWc/YS5oOmEubWFwO2Euc3x8KG09cmEobSxoLGcsYS5sKSk7aWYobSl7bGV0IG49MDtjb25zdCB3PU1hdGgubWluKG0ubGVuZ3RoLGc/YS5BOmEuRCk7Zm9yKGxldCBxPTAscj0wLGwscDtxPHc7cSsrKWlmKGw9bVtxXSlpZihhLnMmJihsPXJhKGwsaCxnLGEubCkpLGUmJmwmJmYmJihwPWwubGVuZ3RoLHA8PWU/KGUtPXAsbD1udWxsKToobD1sLnNsaWNlKGUpLGU9MCkpLGwmJihrW24rK109bCxmJiYocis9bC5sZW5ndGgscj49ZCkpKWJyZWFrO2lmKG4pe2lmKGYpcmV0dXJuIHFhKGssZCwwKTtiW2IubGVuZ3RoXT1rO3JldHVybn19cmV0dXJuIWMmJmt9ZnVuY3Rpb24gcWEoYSxiLGMpe2E9MT09PWEubGVuZ3RoP2FbMF06W10uY29uY2F0LmFwcGx5KFtdLGEpO3JldHVybiBjfHxhLmxlbmd0aD5iP2Euc2xpY2UoYyxjK2IpOmF9XG5mdW5jdGlvbiByYShhLGIsYyxkKXtjPyhkPWQmJmI+YyxhPShhPWFbZD9iOmNdKSYmYVtkP2M6Yl0pOmE9YVtiXTtyZXR1cm4gYX10LmNvbnRhaW49ZnVuY3Rpb24oYSl7cmV0dXJuISF0aGlzLnJlZ2lzdGVyW2FdfTt0LnVwZGF0ZT1mdW5jdGlvbihhLGIpe3JldHVybiB0aGlzLnJlbW92ZShhKS5hZGQoYSxiKX07XG50LnJlbW92ZT1mdW5jdGlvbihhLGIpe2NvbnN0IGM9dGhpcy5yZWdpc3RlclthXTtpZihjKXtpZih0aGlzLm0pZm9yKGxldCBkPTAsZTtkPGMubGVuZ3RoO2QrKyllPWNbZF0sZS5zcGxpY2UoZS5pbmRleE9mKGEpLDEpO2Vsc2UgUSh0aGlzLm1hcCxhLHRoaXMuRCx0aGlzLnMpLHRoaXMuZGVwdGgmJlEodGhpcy5oLGEsdGhpcy5BLHRoaXMucyk7Ynx8ZGVsZXRlIHRoaXMucmVnaXN0ZXJbYV07aWYodGhpcy5jYWNoZSl7Yj10aGlzLmNhY2hlO2ZvcihsZXQgZD0wLGUsZjtkPGIuaC5sZW5ndGg7ZCsrKWY9Yi5oW2RdLGU9Yi5jYWNoZVtmXSxlLmluY2x1ZGVzKGEpJiYoYi5oLnNwbGljZShkLS0sMSksZGVsZXRlIGIuY2FjaGVbZl0pfX1yZXR1cm4gdGhpc307XG5mdW5jdGlvbiBRKGEsYixjLGQsZSl7bGV0IGY9MDtpZihhLmNvbnN0cnVjdG9yPT09QXJyYXkpaWYoZSliPWEuaW5kZXhPZihiKSwtMSE9PWI/MTxhLmxlbmd0aCYmKGEuc3BsaWNlKGIsMSksZisrKTpmKys7ZWxzZXtlPU1hdGgubWluKGEubGVuZ3RoLGMpO2ZvcihsZXQgaD0wLGc7aDxlO2grKylpZihnPWFbaF0pZj1RKGcsYixjLGQsZSksZHx8Znx8ZGVsZXRlIGFbaF19ZWxzZSBmb3IobGV0IGggaW4gYSkoZj1RKGFbaF0sYixjLGQsZSkpfHxkZWxldGUgYVtoXTtyZXR1cm4gZn10LnNlYXJjaENhY2hlPWxhO1xudC5leHBvcnQ9ZnVuY3Rpb24oYSxiLGMsZCxlLGYpe2xldCBoPSEwO1widW5kZWZpbmVkXCI9PT10eXBlb2YgZiYmKGg9bmV3IFByb21pc2UobT0+e2Y9bX0pKTtsZXQgZyxrO3N3aXRjaChlfHwoZT0wKSl7Y2FzZSAwOmc9XCJyZWdcIjtpZih0aGlzLm0pe2s9eCgpO2ZvcihsZXQgbSBpbiB0aGlzLnJlZ2lzdGVyKWtbbV09MX1lbHNlIGs9dGhpcy5yZWdpc3RlcjticmVhaztjYXNlIDE6Zz1cImNmZ1wiO2s9e2RvYzowLG9wdDp0aGlzLnM/MTowfTticmVhaztjYXNlIDI6Zz1cIm1hcFwiO2s9dGhpcy5tYXA7YnJlYWs7Y2FzZSAzOmc9XCJjdHhcIjtrPXRoaXMuaDticmVhaztkZWZhdWx0OlwidW5kZWZpbmVkXCI9PT10eXBlb2YgYyYmZiYmZigpO3JldHVybn1vYShhLGJ8fHRoaXMsYyxnLGQsZSxrLGYpO3JldHVybiBofTtcbnQuaW1wb3J0PWZ1bmN0aW9uKGEsYil7aWYoYilzd2l0Y2goQyhiKSYmKGI9SlNPTi5wYXJzZShiKSksYSl7Y2FzZSBcImNmZ1wiOnRoaXMucz0hIWIub3B0O2JyZWFrO2Nhc2UgXCJyZWdcIjp0aGlzLm09ITE7dGhpcy5yZWdpc3Rlcj1iO2JyZWFrO2Nhc2UgXCJtYXBcIjp0aGlzLm1hcD1iO2JyZWFrO2Nhc2UgXCJjdHhcIjp0aGlzLmg9Yn19O2lhKE4ucHJvdG90eXBlKTtmdW5jdGlvbiBzYShhKXthPWEuZGF0YTt2YXIgYj1zZWxmLl9pbmRleDtjb25zdCBjPWEuYXJnczt2YXIgZD1hLnRhc2s7c3dpdGNoKGQpe2Nhc2UgXCJpbml0XCI6ZD1hLm9wdGlvbnN8fHt9O2E9YS5mYWN0b3J5O2I9ZC5lbmNvZGU7ZC5jYWNoZT0hMTtiJiYwPT09Yi5pbmRleE9mKFwiZnVuY3Rpb25cIikmJihkLmVuY29kZT1GdW5jdGlvbihcInJldHVybiBcIitiKSgpKTthPyhGdW5jdGlvbihcInJldHVybiBcIithKSgpKHNlbGYpLHNlbGYuX2luZGV4PW5ldyBzZWxmLkZsZXhTZWFyY2guSW5kZXgoZCksZGVsZXRlIHNlbGYuRmxleFNlYXJjaCk6c2VsZi5faW5kZXg9bmV3IE4oZCk7YnJlYWs7ZGVmYXVsdDphPWEuaWQsYj1iW2RdLmFwcGx5KGIsYykscG9zdE1lc3NhZ2UoXCJzZWFyY2hcIj09PWQ/e2lkOmEsbXNnOmJ9OntpZDphfSl9fTtsZXQgdGE9MDtmdW5jdGlvbiBTKGEpe2lmKCEodGhpcyBpbnN0YW5jZW9mIFMpKXJldHVybiBuZXcgUyhhKTt2YXIgYjthP0UoYj1hLmVuY29kZSkmJihhLmVuY29kZT1iLnRvU3RyaW5nKCkpOmE9e307KGI9KHNlbGZ8fHdpbmRvdykuX2ZhY3RvcnkpJiYoYj1iLnRvU3RyaW5nKCkpO2NvbnN0IGM9XCJ1bmRlZmluZWRcIj09PXR5cGVvZiB3aW5kb3cmJnNlbGYuZXhwb3J0cyxkPXRoaXM7dGhpcy5vPXVhKGIsYyxhLndvcmtlcik7dGhpcy5oPXgoKTtpZih0aGlzLm8pe2lmKGMpdGhpcy5vLm9uKFwibWVzc2FnZVwiLGZ1bmN0aW9uKGUpe2QuaFtlLmlkXShlLm1zZyk7ZGVsZXRlIGQuaFtlLmlkXX0pO2Vsc2UgdGhpcy5vLm9ubWVzc2FnZT1mdW5jdGlvbihlKXtlPWUuZGF0YTtkLmhbZS5pZF0oZS5tc2cpO2RlbGV0ZSBkLmhbZS5pZF19O3RoaXMuby5wb3N0TWVzc2FnZSh7dGFzazpcImluaXRcIixmYWN0b3J5OmIsb3B0aW9uczphfSl9fVQoXCJhZGRcIik7VChcImFwcGVuZFwiKTtUKFwic2VhcmNoXCIpO1xuVChcInVwZGF0ZVwiKTtUKFwicmVtb3ZlXCIpO2Z1bmN0aW9uIFQoYSl7Uy5wcm90b3R5cGVbYV09Uy5wcm90b3R5cGVbYStcIkFzeW5jXCJdPWZ1bmN0aW9uKCl7Y29uc3QgYj10aGlzLGM9W10uc2xpY2UuY2FsbChhcmd1bWVudHMpO3ZhciBkPWNbYy5sZW5ndGgtMV07bGV0IGU7RShkKSYmKGU9ZCxjLnNwbGljZShjLmxlbmd0aC0xLDEpKTtkPW5ldyBQcm9taXNlKGZ1bmN0aW9uKGYpe3NldFRpbWVvdXQoZnVuY3Rpb24oKXtiLmhbKyt0YV09ZjtiLm8ucG9zdE1lc3NhZ2Uoe3Rhc2s6YSxpZDp0YSxhcmdzOmN9KX0pfSk7cmV0dXJuIGU/KGQudGhlbihlKSx0aGlzKTpkfX1cbmZ1bmN0aW9uIHVhKGEsYixjKXtsZXQgZDt0cnl7ZD1iP25ldyAocmVxdWlyZShcIndvcmtlcl90aHJlYWRzXCIpW1wiV29ya2VyXCJdKShfX2Rpcm5hbWUgKyBcIi9ub2RlL25vZGUuanNcIik6YT9uZXcgV29ya2VyKFVSTC5jcmVhdGVPYmplY3RVUkwobmV3IEJsb2IoW1wib25tZXNzYWdlPVwiK3NhLnRvU3RyaW5nKCldLHt0eXBlOlwidGV4dC9qYXZhc2NyaXB0XCJ9KSkpOm5ldyBXb3JrZXIoQyhjKT9jOlwid29ya2VyL3dvcmtlci5qc1wiLHt0eXBlOlwibW9kdWxlXCJ9KX1jYXRjaChlKXt9cmV0dXJuIGR9O2Z1bmN0aW9uIFUoYSl7aWYoISh0aGlzIGluc3RhbmNlb2YgVSkpcmV0dXJuIG5ldyBVKGEpO3ZhciBiPWEuZG9jdW1lbnR8fGEuZG9jfHxhLGM7dGhpcy5LPVtdO3RoaXMuaD1bXTt0aGlzLkE9W107dGhpcy5yZWdpc3Rlcj14KCk7dGhpcy5rZXk9KGM9Yi5rZXl8fGIuaWQpJiZWKGMsdGhpcy5BKXx8XCJpZFwiO3RoaXMubT11KGEuZmFzdHVwZGF0ZSk7dGhpcy5DPShjPWIuc3RvcmUpJiYhMCE9PWMmJltdO3RoaXMuc3RvcmU9YyYmeCgpO3RoaXMuST0oYz1iLnRhZykmJlYoYyx0aGlzLkEpO3RoaXMubD1jJiZ4KCk7dGhpcy5jYWNoZT0oYz1hLmNhY2hlKSYmbmV3IE0oYyk7YS5jYWNoZT0hMTt0aGlzLm89YS53b3JrZXI7dGhpcy5hc3luYz0hMTtjPXgoKTtsZXQgZD1iLmluZGV4fHxiLmZpZWxkfHxiO0MoZCkmJihkPVtkXSk7Zm9yKGxldCBlPTAsZixoO2U8ZC5sZW5ndGg7ZSsrKWY9ZFtlXSxDKGYpfHwoaD1mLGY9Zi5maWVsZCksaD1EKGgpP09iamVjdC5hc3NpZ24oe30sYSxoKTphLFxudGhpcy5vJiYoY1tmXT1uZXcgUyhoKSxjW2ZdLm98fCh0aGlzLm89ITEpKSx0aGlzLm98fChjW2ZdPW5ldyBOKGgsdGhpcy5yZWdpc3RlcikpLHRoaXMuS1tlXT1WKGYsdGhpcy5BKSx0aGlzLmhbZV09ZjtpZih0aGlzLkMpZm9yKGE9Yi5zdG9yZSxDKGEpJiYoYT1bYV0pLGI9MDtiPGEubGVuZ3RoO2IrKyl0aGlzLkNbYl09VihhW2JdLHRoaXMuQSk7dGhpcy5pbmRleD1jfWZ1bmN0aW9uIFYoYSxiKXtjb25zdCBjPWEuc3BsaXQoXCI6XCIpO2xldCBkPTA7Zm9yKGxldCBlPTA7ZTxjLmxlbmd0aDtlKyspYT1jW2VdLDA8PWEuaW5kZXhPZihcIltdXCIpJiYoYT1hLnN1YnN0cmluZygwLGEubGVuZ3RoLTIpKSYmKGJbZF09ITApLGEmJihjW2QrK109YSk7ZDxjLmxlbmd0aCYmKGMubGVuZ3RoPWQpO3JldHVybiAxPGQ/YzpjWzBdfWZ1bmN0aW9uIFgoYSxiKXtpZihDKGIpKWE9YVtiXTtlbHNlIGZvcihsZXQgYz0wO2EmJmM8Yi5sZW5ndGg7YysrKWE9YVtiW2NdXTtyZXR1cm4gYX1cbmZ1bmN0aW9uIFkoYSxiLGMsZCxlKXthPWFbZV07aWYoZD09PWMubGVuZ3RoLTEpYltlXT1hO2Vsc2UgaWYoYSlpZihhLmNvbnN0cnVjdG9yPT09QXJyYXkpZm9yKGI9YltlXT1BcnJheShhLmxlbmd0aCksZT0wO2U8YS5sZW5ndGg7ZSsrKVkoYSxiLGMsZCxlKTtlbHNlIGI9YltlXXx8KGJbZV09eCgpKSxlPWNbKytkXSxZKGEsYixjLGQsZSl9ZnVuY3Rpb24gWihhLGIsYyxkLGUsZixoLGcpe2lmKGE9YVtoXSlpZihkPT09Yi5sZW5ndGgtMSl7aWYoYS5jb25zdHJ1Y3Rvcj09PUFycmF5KXtpZihjW2RdKXtmb3IoYj0wO2I8YS5sZW5ndGg7YisrKWUuYWRkKGYsYVtiXSwhMCwhMCk7cmV0dXJufWE9YS5qb2luKFwiIFwiKX1lLmFkZChmLGEsZywhMCl9ZWxzZSBpZihhLmNvbnN0cnVjdG9yPT09QXJyYXkpZm9yKGg9MDtoPGEubGVuZ3RoO2grKylaKGEsYixjLGQsZSxmLGgsZyk7ZWxzZSBoPWJbKytkXSxaKGEsYixjLGQsZSxmLGgsZyl9dD1VLnByb3RvdHlwZTtcbnQuYWRkPWZ1bmN0aW9uKGEsYixjKXtEKGEpJiYoYj1hLGE9WChiLHRoaXMua2V5KSk7aWYoYiYmKGF8fDA9PT1hKSl7aWYoIWMmJnRoaXMucmVnaXN0ZXJbYV0pcmV0dXJuIHRoaXMudXBkYXRlKGEsYik7Zm9yKGxldCBkPTAsZSxmO2Q8dGhpcy5oLmxlbmd0aDtkKyspZj10aGlzLmhbZF0sZT10aGlzLktbZF0sQyhlKSYmKGU9W2VdKSxaKGIsZSx0aGlzLkEsMCx0aGlzLmluZGV4W2ZdLGEsZVswXSxjKTtpZih0aGlzLkkpe2xldCBkPVgoYix0aGlzLkkpLGU9eCgpO0MoZCkmJihkPVtkXSk7Zm9yKGxldCBmPTAsaCxnO2Y8ZC5sZW5ndGg7ZisrKWlmKGg9ZFtmXSwhZVtoXSYmKGVbaF09MSxnPXRoaXMubFtoXXx8KHRoaXMubFtoXT1bXSksIWN8fCFnLmluY2x1ZGVzKGEpKSlpZihnW2cubGVuZ3RoXT1hLHRoaXMubSl7Y29uc3Qgaz10aGlzLnJlZ2lzdGVyW2FdfHwodGhpcy5yZWdpc3RlclthXT1bXSk7a1trLmxlbmd0aF09Z319aWYodGhpcy5zdG9yZSYmKCFjfHwhdGhpcy5zdG9yZVthXSkpe2xldCBkO1xuaWYodGhpcy5DKXtkPXgoKTtmb3IobGV0IGU9MCxmO2U8dGhpcy5DLmxlbmd0aDtlKyspZj10aGlzLkNbZV0sQyhmKT9kW2ZdPWJbZl06WShiLGQsZiwwLGZbMF0pfXRoaXMuc3RvcmVbYV09ZHx8Yn19cmV0dXJuIHRoaXN9O3QuYXBwZW5kPWZ1bmN0aW9uKGEsYil7cmV0dXJuIHRoaXMuYWRkKGEsYiwhMCl9O3QudXBkYXRlPWZ1bmN0aW9uKGEsYil7cmV0dXJuIHRoaXMucmVtb3ZlKGEpLmFkZChhLGIpfTtcbnQucmVtb3ZlPWZ1bmN0aW9uKGEpe0QoYSkmJihhPVgoYSx0aGlzLmtleSkpO2lmKHRoaXMucmVnaXN0ZXJbYV0pe2Zvcih2YXIgYj0wO2I8dGhpcy5oLmxlbmd0aCYmKHRoaXMuaW5kZXhbdGhpcy5oW2JdXS5yZW1vdmUoYSwhdGhpcy5vKSwhdGhpcy5tKTtiKyspO2lmKHRoaXMuSSYmIXRoaXMubSlmb3IobGV0IGMgaW4gdGhpcy5sKXtiPXRoaXMubFtjXTtjb25zdCBkPWIuaW5kZXhPZihhKTstMSE9PWQmJigxPGIubGVuZ3RoP2Iuc3BsaWNlKGQsMSk6ZGVsZXRlIHRoaXMubFtjXSl9dGhpcy5zdG9yZSYmZGVsZXRlIHRoaXMuc3RvcmVbYV07ZGVsZXRlIHRoaXMucmVnaXN0ZXJbYV19cmV0dXJuIHRoaXN9O1xudC5zZWFyY2g9ZnVuY3Rpb24oYSxiLGMsZCl7Y3x8KCFiJiZEKGEpPyhjPWEsYT1cIlwiKTpEKGIpJiYoYz1iLGI9MCkpO2xldCBlPVtdLGY9W10saCxnLGssbSxuLHcscT0wO2lmKGMpaWYoYy5jb25zdHJ1Y3Rvcj09PUFycmF5KWs9YyxjPW51bGw7ZWxzZXthPWMucXVlcnl8fGE7az0oaD1jLnBsdWNrKXx8Yy5pbmRleHx8Yy5maWVsZDttPWMudGFnO2c9dGhpcy5zdG9yZSYmYy5lbnJpY2g7bj1cImFuZFwiPT09Yy5ib29sO2I9Yy5saW1pdHx8Ynx8MTAwO3c9Yy5vZmZzZXR8fDA7aWYobSYmKEMobSkmJihtPVttXSksIWEpKXtmb3IobGV0IGw9MCxwO2w8bS5sZW5ndGg7bCsrKWlmKHA9dmEuY2FsbCh0aGlzLG1bbF0sYix3LGcpKWVbZS5sZW5ndGhdPXAscSsrO3JldHVybiBxP2U6W119QyhrKSYmKGs9W2tdKX1rfHwoaz10aGlzLmgpO249biYmKDE8ay5sZW5ndGh8fG0mJjE8bS5sZW5ndGgpO2NvbnN0IHI9IWQmJih0aGlzLm98fHRoaXMuYXN5bmMpJiZbXTtmb3IobGV0IGw9MCxwLEEsQjtsPFxuay5sZW5ndGg7bCsrKXtsZXQgejtBPWtbbF07QyhBKXx8KHo9QSxBPXouZmllbGQsYT16LnF1ZXJ5fHxhLGI9ei5saW1pdHx8YixnPXouZW5yaWNofHxnKTtpZihyKXJbbF09dGhpcy5pbmRleFtBXS5zZWFyY2hBc3luYyhhLGIsenx8Yyk7ZWxzZXtkP3A9ZFtsXTpwPXRoaXMuaW5kZXhbQV0uc2VhcmNoKGEsYix6fHxjKTtCPXAmJnAubGVuZ3RoO2lmKG0mJkIpe2NvbnN0IHk9W107bGV0IEg9MDtuJiYoeVswXT1bcF0pO2ZvcihsZXQgVz0wLG1hLFI7VzxtLmxlbmd0aDtXKyspaWYobWE9bVtXXSxCPShSPXRoaXMubFttYV0pJiZSLmxlbmd0aClIKysseVt5Lmxlbmd0aF09bj9bUl06UjtIJiYocD1uP2phKHksYnx8MTAwLHd8fDApOmthKHAseSksQj1wLmxlbmd0aCl9aWYoQilmW3FdPUEsZVtxKytdPXA7ZWxzZSBpZihuKXJldHVybltdfX1pZihyKXtjb25zdCBsPXRoaXM7cmV0dXJuIG5ldyBQcm9taXNlKGZ1bmN0aW9uKHApe1Byb21pc2UuYWxsKHIpLnRoZW4oZnVuY3Rpb24oQSl7cChsLnNlYXJjaChhLFxuYixjLEEpKX0pfSl9aWYoIXEpcmV0dXJuW107aWYoaCYmKCFnfHwhdGhpcy5zdG9yZSkpcmV0dXJuIGVbMF07Zm9yKGxldCBsPTAscDtsPGYubGVuZ3RoO2wrKyl7cD1lW2xdO3AubGVuZ3RoJiZnJiYocD13YS5jYWxsKHRoaXMscCkpO2lmKGgpcmV0dXJuIHA7ZVtsXT17ZmllbGQ6ZltsXSxyZXN1bHQ6cH19cmV0dXJuIGV9O2Z1bmN0aW9uIHZhKGEsYixjLGQpe2xldCBlPXRoaXMubFthXSxmPWUmJmUubGVuZ3RoLWM7aWYoZiYmMDxmKXtpZihmPmJ8fGMpZT1lLnNsaWNlKGMsYytiKTtkJiYoZT13YS5jYWxsKHRoaXMsZSkpO3JldHVybnt0YWc6YSxyZXN1bHQ6ZX19fWZ1bmN0aW9uIHdhKGEpe2NvbnN0IGI9QXJyYXkoYS5sZW5ndGgpO2ZvcihsZXQgYz0wLGQ7YzxhLmxlbmd0aDtjKyspZD1hW2NdLGJbY109e2lkOmQsZG9jOnRoaXMuc3RvcmVbZF19O3JldHVybiBifXQuY29udGFpbj1mdW5jdGlvbihhKXtyZXR1cm4hIXRoaXMucmVnaXN0ZXJbYV19O3QuZ2V0PWZ1bmN0aW9uKGEpe3JldHVybiB0aGlzLnN0b3JlW2FdfTtcbnQuc2V0PWZ1bmN0aW9uKGEsYil7dGhpcy5zdG9yZVthXT1iO3JldHVybiB0aGlzfTt0LnNlYXJjaENhY2hlPWxhO3QuZXhwb3J0PWZ1bmN0aW9uKGEsYixjLGQsZSxmKXtsZXQgaDtcInVuZGVmaW5lZFwiPT09dHlwZW9mIGYmJihoPW5ldyBQcm9taXNlKGc9PntmPWd9KSk7ZXx8KGU9MCk7ZHx8KGQ9MCk7aWYoZDx0aGlzLmgubGVuZ3RoKXtjb25zdCBnPXRoaXMuaFtkXSxrPXRoaXMuaW5kZXhbZ107Yj10aGlzO3NldFRpbWVvdXQoZnVuY3Rpb24oKXtrLmV4cG9ydChhLGIsZT9nOlwiXCIsZCxlKyssZil8fChkKyssZT0xLGIuZXhwb3J0KGEsYixnLGQsZSxmKSl9KX1lbHNle2xldCBnLGs7c3dpdGNoKGUpe2Nhc2UgMTpnPVwidGFnXCI7az10aGlzLmw7Yz1udWxsO2JyZWFrO2Nhc2UgMjpnPVwic3RvcmVcIjtrPXRoaXMuc3RvcmU7Yz1udWxsO2JyZWFrO2RlZmF1bHQ6ZigpO3JldHVybn1vYShhLHRoaXMsYyxnLGQsZSxrLGYpfXJldHVybiBofTtcbnQuaW1wb3J0PWZ1bmN0aW9uKGEsYil7aWYoYilzd2l0Y2goQyhiKSYmKGI9SlNPTi5wYXJzZShiKSksYSl7Y2FzZSBcInRhZ1wiOnRoaXMubD1iO2JyZWFrO2Nhc2UgXCJyZWdcIjp0aGlzLm09ITE7dGhpcy5yZWdpc3Rlcj1iO2ZvcihsZXQgZD0wLGU7ZDx0aGlzLmgubGVuZ3RoO2QrKyllPXRoaXMuaW5kZXhbdGhpcy5oW2RdXSxlLnJlZ2lzdGVyPWIsZS5tPSExO2JyZWFrO2Nhc2UgXCJzdG9yZVwiOnRoaXMuc3RvcmU9YjticmVhaztkZWZhdWx0OmE9YS5zcGxpdChcIi5cIik7Y29uc3QgYz1hWzBdO2E9YVsxXTtjJiZhJiZ0aGlzLmluZGV4W2NdLmltcG9ydChhLGIpfX07aWEoVS5wcm90b3R5cGUpO3ZhciB5YT17ZW5jb2RlOnhhLEY6ITEsRzpcIlwifTtjb25zdCB6YT1bSihcIltcXHUwMGUwXFx1MDBlMVxcdTAwZTJcXHUwMGUzXFx1MDBlNFxcdTAwZTVdXCIpLFwiYVwiLEooXCJbXFx1MDBlOFxcdTAwZTlcXHUwMGVhXFx1MDBlYl1cIiksXCJlXCIsSihcIltcXHUwMGVjXFx1MDBlZFxcdTAwZWVcXHUwMGVmXVwiKSxcImlcIixKKFwiW1xcdTAwZjJcXHUwMGYzXFx1MDBmNFxcdTAwZjVcXHUwMGY2XFx1MDE1MV1cIiksXCJvXCIsSihcIltcXHUwMGY5XFx1MDBmYVxcdTAwZmJcXHUwMGZjXFx1MDE3MV1cIiksXCJ1XCIsSihcIltcXHUwMGZkXFx1MDE3N1xcdTAwZmZdXCIpLFwieVwiLEooXCJcXHUwMGYxXCIpLFwiblwiLEooXCJbXFx1MDBlN2NdXCIpLFwia1wiLEooXCJcXHUwMGRmXCIpLFwic1wiLEooXCIgJiBcIiksXCIgYW5kIFwiXTtmdW5jdGlvbiB4YShhKXt2YXIgYj1hPVwiXCIrYTtiLm5vcm1hbGl6ZSYmKGI9Yi5ub3JtYWxpemUoXCJORkRcIikucmVwbGFjZShjYSxcIlwiKSk7cmV0dXJuIEYuY2FsbCh0aGlzLGIudG9Mb3dlckNhc2UoKSwhYS5ub3JtYWxpemUmJnphKX07dmFyIEJhPXtlbmNvZGU6QWEsRjohMSxHOlwic3RyaWN0XCJ9O2NvbnN0IENhPS9bXmEtejAtOV0rLyxEYT17YjpcInBcIix2OlwiZlwiLHc6XCJmXCIsejpcInNcIix4Olwic1wiLFwiXFx1MDBkZlwiOlwic1wiLGQ6XCJ0XCIsbjpcIm1cIixjOlwia1wiLGc6XCJrXCIsajpcImtcIixxOlwia1wiLGk6XCJlXCIseTpcImVcIix1Olwib1wifTtmdW5jdGlvbiBBYShhKXthPXhhLmNhbGwodGhpcyxhKS5qb2luKFwiIFwiKTtjb25zdCBiPVtdO2lmKGEpe2NvbnN0IGM9YS5zcGxpdChDYSksZD1jLmxlbmd0aDtmb3IobGV0IGU9MCxmLGg9MDtlPGQ7ZSsrKWlmKChhPWNbZV0pJiYoIXRoaXMuZmlsdGVyfHwhdGhpcy5maWx0ZXJbYV0pKXtmPWFbMF07bGV0IGc9RGFbZl18fGYsaz1nO2ZvcihsZXQgbT0xO208YS5sZW5ndGg7bSsrKXtmPWFbbV07Y29uc3Qgbj1EYVtmXXx8ZjtuJiZuIT09ayYmKGcrPW4saz1uKX1iW2grK109Z319cmV0dXJuIGJ9O3ZhciBGYT17ZW5jb2RlOkVhLEY6ITEsRzpcIlwifTtjb25zdCBHYT1bSihcImFlXCIpLFwiYVwiLEooXCJvZVwiKSxcIm9cIixKKFwic2hcIiksXCJzXCIsSihcInRoXCIpLFwidFwiLEooXCJwaFwiKSxcImZcIixKKFwicGZcIiksXCJmXCIsSihcIig/IVthZW9dKWgoPyFbYWVvXSlcIiksXCJcIixKKFwiKD8hXlthZW9dKWgoPyFeW2Flb10pXCIpLFwiXCJdO2Z1bmN0aW9uIEVhKGEsYil7YSYmKGE9QWEuY2FsbCh0aGlzLGEpLmpvaW4oXCIgXCIpLDI8YS5sZW5ndGgmJihhPUcoYSxHYSkpLGJ8fCgxPGEubGVuZ3RoJiYoYT1kYShhKSksYSYmKGE9YS5zcGxpdChcIiBcIikpKSk7cmV0dXJuIGF8fFtdfTt2YXIgSWE9e2VuY29kZTpIYSxGOiExLEc6XCJcIn07Y29uc3QgSmE9SihcIig/IVxcXFxiKVthZW9dXCIpO2Z1bmN0aW9uIEhhKGEpe2EmJihhPUVhLmNhbGwodGhpcyxhLCEwKSwxPGEubGVuZ3RoJiYoYT1hLnJlcGxhY2UoSmEsXCJcIikpLDE8YS5sZW5ndGgmJihhPWRhKGEpKSxhJiYoYT1hLnNwbGl0KFwiIFwiKSkpO3JldHVybiBhfHxbXX07S1tcImxhdGluOmRlZmF1bHRcIl09ZmE7S1tcImxhdGluOnNpbXBsZVwiXT15YTtLW1wibGF0aW46YmFsYW5jZVwiXT1CYTtLW1wibGF0aW46YWR2YW5jZWRcIl09RmE7S1tcImxhdGluOmV4dHJhXCJdPUlhO2V4cG9ydCBkZWZhdWx0IHtJbmRleDpOLERvY3VtZW50OlUsV29ya2VyOlMscmVnaXN0ZXJDaGFyc2V0OmZ1bmN0aW9uKGEsYil7S1thXT1ifSxyZWdpc3Rlckxhbmd1YWdlOmZ1bmN0aW9uKGEsYil7aGFbYV09Yn19O1xuIiwgIi8qIVxuICogRmxleFNlYXJjaCBmb3IgQm9vdHN0cmFwIGJhc2VkIFRodWxpdGUgc2l0ZXNcbiAqIENvcHlyaWdodCAyMDIxLTIwMjQgVGh1bGl0ZVxuICogTGljZW5zZWQgdW5kZXIgdGhlIE1JVCBMaWNlbnNlXG4gKiBCYXNlZCBvbiBodHRwczovL2dpdGh1Yi5jb20vZnJqby9odWdvLXRoZW1lLXplbi9ibG9iL21haW4vYXNzZXRzL2pzL3NlYXJjaC5qc1xuICovXG5cbi8qIGVzbGludC1kaXNhYmxlIG5vLXVuZGVmLCBndWFyZC1mb3ItaW4gKi9cblxuLyoqXG4gKiBAZmlsZVxuICogQSBKYXZhU2NyaXB0IGZpbGUgZm9yIGZsZXhzZWFyY2guXG4gKi9cblxuLy8gaW1wb3J0ICogYXMgRmxleFNlYXJjaCBmcm9tICdmbGV4c2VhcmNoJztcbmltcG9ydCBJbmRleCBmcm9tICdmbGV4c2VhcmNoJztcblxuKGZ1bmN0aW9uICgpIHtcblxuICAndXNlIHN0cmljdCc7XG5cbiAgLy8gY29uc3QgaW5kZXggPSBuZXcgRmxleFNlYXJjaC5Eb2N1bWVudCh7XG4gIGNvbnN0IGluZGV4ID0gbmV3IEluZGV4LkRvY3VtZW50KHtcbiAgICB0b2tlbml6ZTogJ2ZvcndhcmQnLFxuICAgIGRvY3VtZW50OiB7XG4gICAgICBpZDogJ2lkJyxcbiAgICAgIGluZGV4OiBbXG4gICAgICAgIHtcbiAgICAgICAgICBmaWVsZDogJ3RpdGxlJ1xuICAgICAgICB9LFxuICAgICAgICB7XG4gICAgICAgICAgZmllbGQ6ICd0YWdzJ1xuICAgICAgICB9LFxuICAgICAgICB7XG4gICAgICAgICAgZmllbGQ6ICdjb250ZW50J1xuICAgICAgICB9LFxuICAgICAgICB7XG4gICAgICAgICAgZmllbGQ6ICAnZGF0ZScsXG4gICAgICAgICAgdG9rZW5pemU6ICdzdHJpY3QnLFxuICAgICAgICAgIGVuY29kZTogZmFsc2VcbiAgICAgICAgfVxuICAgICAgXSxcbiAgICAgIHN0b3JlOiBbJ3RpdGxlJywnc3VtbWFyeScsJ2RhdGUnLCdwZXJtYWxpbmsnXVxuICAgIH1cbiAgfSk7XG5cbiAgZnVuY3Rpb24gc2hvd1Jlc3VsdHMoaXRlbXMpIHtcbiAgICBjb25zdCB0ZW1wbGF0ZSA9IGRvY3VtZW50LnF1ZXJ5U2VsZWN0b3IoJ3RlbXBsYXRlJykuY29udGVudDtcbiAgICBjb25zdCBmcmFnbWVudCA9IGRvY3VtZW50LmNyZWF0ZURvY3VtZW50RnJhZ21lbnQoKTtcblxuICAgIGNvbnN0IHJlc3VsdHMgPSBkb2N1bWVudC5xdWVyeVNlbGVjdG9yKCcuc2VhcmNoLXJlc3VsdHMnKTtcbiAgICByZXN1bHRzLnRleHRDb250ZW50ID0gJyc7XG5cbiAgICBjb25zdCBpdGVtc0xlbmd0aCA9IE9iamVjdC5rZXlzKGl0ZW1zKS5sZW5ndGg7XG5cbiAgICAvLyBTaG93L2hpZGUgXCJObyByZWNlbnQgc2VhcmNoZXNcIiBhbmQgXCJObyBzZWFyY2ggcmVzdWx0c1wiIG1lc3NhZ2VzXG4gICAgaWYgKChpdGVtc0xlbmd0aCA9PT0gMCkgJiYgKHF1ZXJ5LnZhbHVlID09PSAnJykpIHtcbiAgICAgIC8vIEhpZGUgXCJObyBzZWFyY2ggcmVzdWx0c1wiIG1lc3NhZ2VcbiAgICAgIGRvY3VtZW50LnF1ZXJ5U2VsZWN0b3IoJy5zZWFyY2gtbm8tcmVzdWx0cycpLmNsYXNzTGlzdC5hZGQoJ2Qtbm9uZScpO1xuICAgICAgLy8gU2hvdyBcIk5vIHJlY2VudCBzZWFyY2hlc1wiIG1lc3NhZ2VcbiAgICAgIGRvY3VtZW50LnF1ZXJ5U2VsZWN0b3IoJy5zZWFyY2gtbm8tcmVjZW50JykuY2xhc3NMaXN0LnJlbW92ZSgnZC1ub25lJyk7XG4gICAgfSBlbHNlIGlmICgoaXRlbXNMZW5ndGggPT09IDApICYmIChxdWVyeS52YWx1ZSAhPT0gJycpKSB7XG4gICAgICAvLyBIaWRlIFwiTm8gcmVjZW50IHNlYXJjaGVzXCIgbWVzc2FnZVxuICAgICAgZG9jdW1lbnQucXVlcnlTZWxlY3RvcignLnNlYXJjaC1uby1yZWNlbnQnKS5jbGFzc0xpc3QuYWRkKCdkLW5vbmUnKTtcbiAgICAgIC8vIFNob3cgXCJObyBzZWFyY2ggcmVzdWx0c1wiIG1lc3NhZ2VcbiAgICAgIGNvbnN0IHF1ZXJ5Tm9SZXN1bHRzID0gZG9jdW1lbnQucXVlcnlTZWxlY3RvcignLnF1ZXJ5LW5vLXJlc3VsdHMnKTtcbiAgICAgIHF1ZXJ5Tm9SZXN1bHRzLmlubmVyVGV4dCA9IHF1ZXJ5LnZhbHVlO1xuICAgICAgZG9jdW1lbnQucXVlcnlTZWxlY3RvcignLnNlYXJjaC1uby1yZXN1bHRzJykuY2xhc3NMaXN0LnJlbW92ZSgnZC1ub25lJyk7XG4gICAgfSBlbHNlIHtcbiAgICAgIC8vIEhpZGUgYm90aCBcIk5vIHJlY2VudCBzZWFyY2hlc1wiIGFuZCBcIk5vIHNlYXJjaCByZXN1bHRzXCIgbWVzc2FnZXNcbiAgICAgIGRvY3VtZW50LnF1ZXJ5U2VsZWN0b3IoJy5zZWFyY2gtbm8tcmVjZW50JykuY2xhc3NMaXN0LmFkZCgnZC1ub25lJyk7XG4gICAgICBkb2N1bWVudC5xdWVyeVNlbGVjdG9yKCcuc2VhcmNoLW5vLXJlc3VsdHMnKS5jbGFzc0xpc3QuYWRkKCdkLW5vbmUnKTtcbiAgICB9XG5cbiAgICBmb3IgKGNvbnN0IGlkIGluIGl0ZW1zKSB7XG4gICAgICBjb25zdCBpdGVtID0gaXRlbXNbaWRdO1xuICAgICAgY29uc3QgcmVzdWx0ID0gdGVtcGxhdGUuY2xvbmVOb2RlKHRydWUpO1xuICAgICAgY29uc3QgYSA9IHJlc3VsdC5xdWVyeVNlbGVjdG9yKCdhJyk7XG4gICAgICBjb25zdCB0aW1lID0gcmVzdWx0LnF1ZXJ5U2VsZWN0b3IoJ3RpbWUnKTtcbiAgICAgIGNvbnN0IGNvbnRlbnQgPSByZXN1bHQucXVlcnlTZWxlY3RvcignLmNvbnRlbnQnKTtcbiAgICAgIGEuaW5uZXJIVE1MID0gaXRlbS50aXRsZTtcbiAgICAgIGEuaHJlZiA9IGl0ZW0ucGVybWFsaW5rO1xuICAgICAgdGltZS5pbm5lclRleHQgPSBpdGVtLmRhdGU7XG4gICAgICBjb250ZW50LmlubmVySFRNTCA9IGl0ZW0uc3VtbWFyeTtcbiAgICAgIGZyYWdtZW50LmFwcGVuZENoaWxkKHJlc3VsdCk7XG4gICAgfVxuXG4gICAgcmVzdWx0cy5hcHBlbmRDaGlsZChmcmFnbWVudCk7XG4gIH1cblxuICBmdW5jdGlvbiBkb1NlYXJjaCgpIHtcbiAgICBjb25zdCBxdWVyeSA9IGRvY3VtZW50LnF1ZXJ5U2VsZWN0b3IoJy5zZWFyY2gtdGV4dCcpLnZhbHVlLnRyaW0oKTtcbiAgICBjb25zdCBsaW1pdCA9IDk5O1xuICAgIGNvbnN0IHJlc3VsdHMgPSBpbmRleC5zZWFyY2goe1xuICAgICAgcXVlcnk6IHF1ZXJ5LFxuICAgICAgZW5yaWNoOiB0cnVlLFxuICAgICAgbGltaXQ6IGxpbWl0LFxuICAgIH0pO1xuICAgIGNvbnN0IGl0ZW1zID0ge307XG5cbiAgICByZXN1bHRzLmZvckVhY2goZnVuY3Rpb24gKHJlc3VsdCkge1xuICAgICAgcmVzdWx0LnJlc3VsdC5mb3JFYWNoKGZ1bmN0aW9uIChyKSB7XG4gICAgICAgIGl0ZW1zW3IuaWRdID0gci5kb2M7XG4gICAgICB9KTtcbiAgICB9KTtcblxuICAgIHNob3dSZXN1bHRzKGl0ZW1zKTtcbiAgfVxuXG4gIGZ1bmN0aW9uIGVuYWJsZVVJKCkge1xuICAgIGNvbnN0IHNlYXJjaGZvcm0gPSBkb2N1bWVudC5xdWVyeVNlbGVjdG9yKCcuc2VhcmNoLWZvcm0nKTtcbiAgICBzZWFyY2hmb3JtLmFkZEV2ZW50TGlzdGVuZXIoJ3N1Ym1pdCcsIGZ1bmN0aW9uIChlKSB7XG4gICAgICBlLnByZXZlbnREZWZhdWx0KCk7XG4gICAgICBkb1NlYXJjaCgpO1xuICAgIH0pO1xuICAgIHNlYXJjaGZvcm0uYWRkRXZlbnRMaXN0ZW5lcignaW5wdXQnLCBmdW5jdGlvbiAoKSB7XG4gICAgICBkb1NlYXJjaCgpO1xuICAgIH0pO1xuICAgIGRvY3VtZW50LnF1ZXJ5U2VsZWN0b3IoJy5zZWFyY2gtbG9hZGluZycpLmNsYXNzTGlzdC5hZGQoJ2Qtbm9uZScpO1xuICAgIGRvY3VtZW50LnF1ZXJ5U2VsZWN0b3IoJy5zZWFyY2gtaW5wdXQnKS5jbGFzc0xpc3QucmVtb3ZlKCdkLW5vbmUnKTtcbiAgICBkb2N1bWVudC5xdWVyeVNlbGVjdG9yKCcuc2VhcmNoLXRleHQnKS5mb2N1cygpO1xuICB9XG5cbiAgZnVuY3Rpb24gYnVpbGRJbmRleCgpIHtcbiAgICBkb2N1bWVudC5xdWVyeVNlbGVjdG9yKCcuc2VhcmNoLWxvYWRpbmcnKS5jbGFzc0xpc3QucmVtb3ZlKCdkLW5vbmUnKTtcbiAgICBmZXRjaChcIi9nZGcvc2VhcmNoLWluZGV4Lmpzb25cIilcbiAgICAgIC50aGVuKGZ1bmN0aW9uIChyZXNwb25zZSkge1xuICAgICAgICByZXR1cm4gcmVzcG9uc2UuanNvbigpO1xuICAgICAgfSlcbiAgICAgIC50aGVuKGZ1bmN0aW9uIChkYXRhKSB7XG4gICAgICAgIGRhdGEuZm9yRWFjaChmdW5jdGlvbiAoaXRlbSkge1xuICAgICAgICAgIGluZGV4LmFkZChpdGVtKTtcbiAgICAgICAgfSk7XG4gICAgICB9KTtcbiAgfVxuXG4gIGJ1aWxkSW5kZXgoKTtcbiAgZW5hYmxlVUkoKTtcbn0pKCk7XG4iXSwKICAibWFwcGluZ3MiOiAiOzs7Ozs7Ozs7Ozs7O0FBT0EsTUFBSTtBQUFFLFdBQVMsRUFBRSxHQUFFO0FBQUMsV0FBTSxnQkFBYyxPQUFPLElBQUUsSUFBRTtBQUFBLEVBQUU7QUFBQyxXQUFTLEVBQUUsR0FBRTtBQUFDLFVBQU0sSUFBRSxNQUFNLENBQUM7QUFBRSxhQUFRLElBQUUsR0FBRSxJQUFFLEdBQUUsSUFBSSxHQUFFLENBQUMsSUFBRSxFQUFFO0FBQUUsV0FBTztBQUFBLEVBQUM7QUFBQyxXQUFTLElBQUc7QUFBQyxXQUFPLHVCQUFPLE9BQU8sSUFBSTtBQUFBLEVBQUM7QUFBQyxXQUFTLEdBQUcsR0FBRSxHQUFFO0FBQUMsV0FBTyxFQUFFLFNBQU8sRUFBRTtBQUFBLEVBQU07QUFBQyxXQUFTLEVBQUUsR0FBRTtBQUFDLFdBQU0sYUFBVyxPQUFPO0FBQUEsRUFBQztBQUFDLFdBQVMsRUFBRSxHQUFFO0FBQUMsV0FBTSxhQUFXLE9BQU87QUFBQSxFQUFDO0FBQUMsV0FBUyxFQUFFLEdBQUU7QUFBQyxXQUFNLGVBQWEsT0FBTztBQUFBLEVBQUM7QUFBRSxXQUFTLEVBQUUsR0FBRSxHQUFFO0FBQUMsUUFBSSxJQUFFO0FBQUcsUUFBRyxNQUFJLE1BQUksSUFBRSxFQUFFLEdBQUUsQ0FBQyxJQUFHLEtBQUssTUFBSSxJQUFFLEVBQUUsR0FBRSxLQUFLLENBQUMsSUFBRyxLQUFLLEtBQUcsSUFBRSxFQUFFLFdBQVMsSUFBRSxFQUFFLEdBQUUsS0FBSyxDQUFDLElBQUcsS0FBRyxPQUFLLElBQUc7QUFBQyxVQUFFLEVBQUUsTUFBTSxDQUFDO0FBQUUsVUFBRyxLQUFLLFFBQU87QUFBQyxZQUFFLEtBQUs7QUFBTyxZQUFFLEVBQUU7QUFBTyxjQUFNLElBQUUsQ0FBQztBQUFFLGlCQUFRLElBQUUsR0FBRSxJQUFFLEdBQUUsSUFBRSxHQUFFLEtBQUk7QUFBQyxnQkFBTSxJQUFFLEVBQUUsQ0FBQztBQUFFLGVBQUcsQ0FBQyxFQUFFLENBQUMsTUFBSSxFQUFFLEdBQUcsSUFBRTtBQUFBLFFBQUU7QUFBQyxZQUFFO0FBQUEsTUFBQyxNQUFNLEtBQUU7QUFBRSxhQUFPO0FBQUEsSUFBQztBQUFDLFdBQU87QUFBQSxFQUFDO0FBQUMsTUFBTSxLQUFHO0FBQVQsTUFBb0MsS0FBRztBQUNob0IsV0FBUyxFQUFFLEdBQUUsR0FBRTtBQUFDLFVBQU0sSUFBRSxPQUFPLEtBQUssQ0FBQyxHQUFFLElBQUUsRUFBRSxRQUFPLElBQUUsQ0FBQztBQUFFLFFBQUksSUFBRSxJQUFHLElBQUU7QUFBRSxhQUFRLElBQUUsR0FBRSxHQUFFLEdBQUUsSUFBRSxHQUFFLElBQUksS0FBRSxFQUFFLENBQUMsSUFBRyxJQUFFLEVBQUUsQ0FBQyxNQUFJLEVBQUUsR0FBRyxJQUFFLEVBQUUsSUFBRSxZQUFVLElBQUUsWUFBVSxDQUFDLEdBQUUsRUFBRSxHQUFHLElBQUUsS0FBRyxNQUFJLElBQUUsTUFBSSxNQUFJO0FBQUUsVUFBSSxFQUFFLEdBQUcsSUFBRSxFQUFFLElBQUUsYUFBVyxJQUFFLGFBQVcsTUFBSSxJQUFFLEdBQUcsR0FBRSxFQUFFLENBQUMsSUFBRTtBQUFJLFdBQU87QUFBQSxFQUFDO0FBQUMsV0FBUyxFQUFFLEdBQUUsR0FBRTtBQUFDLGFBQVEsSUFBRSxHQUFFLElBQUUsRUFBRSxRQUFPLElBQUUsTUFBSSxJQUFFLEVBQUUsUUFBUSxFQUFFLENBQUMsR0FBRSxFQUFFLElBQUUsQ0FBQyxDQUFDLEdBQUUsSUFBRyxLQUFHLEVBQUU7QUFBQyxXQUFPO0FBQUEsRUFBQztBQUFDLFdBQVMsRUFBRSxHQUFFO0FBQUMsV0FBTyxJQUFJLE9BQU8sR0FBRSxHQUFHO0FBQUEsRUFBQztBQUFDLFdBQVMsR0FBRyxHQUFFO0FBQUMsUUFBSSxJQUFFLElBQUcsSUFBRTtBQUFHLGFBQVEsSUFBRSxHQUFFLElBQUUsRUFBRSxRQUFPLEdBQUUsSUFBRSxHQUFFLElBQUksRUFBQyxJQUFFLEVBQUUsQ0FBQyxPQUFLLE1BQUksS0FBRyxJQUFFO0FBQUcsV0FBTztBQUFBLEVBQUM7QUFBRSxNQUFJLEtBQUcsRUFBQyxRQUFPLElBQUcsR0FBRSxPQUFHLEdBQUUsR0FBRTtBQUFFLFdBQVMsR0FBRyxHQUFFO0FBQUMsV0FBTyxFQUFFLEtBQUssT0FBTSxLQUFHLEdBQUcsWUFBWSxHQUFFLEtBQUU7QUFBQSxFQUFDO0FBQUUsTUFBTSxLQUFHLENBQUM7QUFBVixNQUFZLElBQUUsQ0FBQztBQUFFLFdBQVMsR0FBRyxHQUFFO0FBQUMsTUFBRSxHQUFFLEtBQUs7QUFBRSxNQUFFLEdBQUUsUUFBUTtBQUFFLE1BQUUsR0FBRSxRQUFRO0FBQUUsTUFBRSxHQUFFLFFBQVE7QUFBRSxNQUFFLEdBQUUsUUFBUTtBQUFBLEVBQUM7QUFBQyxXQUFTLEVBQUUsR0FBRSxHQUFFO0FBQUMsTUFBRSxJQUFFLE9BQU8sSUFBRSxXQUFVO0FBQUMsWUFBTSxJQUFFLE1BQUssSUFBRTtBQUFVLFVBQUksSUFBRSxFQUFFLEVBQUUsU0FBTyxDQUFDO0FBQUUsVUFBSTtBQUFFLFFBQUUsQ0FBQyxNQUFJLElBQUUsR0FBRSxPQUFPLEVBQUUsRUFBRSxTQUFPLENBQUM7QUFBRyxVQUFFLElBQUksUUFBUSxTQUFTLEdBQUU7QUFBQyxtQkFBVyxXQUFVO0FBQUMsWUFBRSxRQUFNO0FBQUcsZ0JBQU0sSUFBRSxFQUFFLENBQUMsRUFBRSxNQUFNLEdBQUUsQ0FBQztBQUFFLFlBQUUsUUFBTTtBQUFHLFlBQUUsQ0FBQztBQUFBLFFBQUMsQ0FBQztBQUFBLE1BQUMsQ0FBQztBQUFFLGFBQU8sS0FBRyxFQUFFLEtBQUssQ0FBQyxHQUFFLFFBQU07QUFBQSxJQUFDO0FBQUEsRUFBQztBQUFFLFdBQVMsR0FBRyxHQUFFLEdBQUUsR0FBRSxHQUFFO0FBQUMsVUFBTSxJQUFFLEVBQUU7QUFBTyxRQUFJLElBQUUsQ0FBQyxHQUFFLEdBQUUsR0FBRSxJQUFFO0FBQUUsVUFBSSxJQUFFLENBQUM7QUFBRyxhQUFRLElBQUUsSUFBRSxHQUFFLEtBQUcsR0FBRSxLQUFJO0FBQUMsWUFBTSxJQUFFLEVBQUUsQ0FBQyxHQUFFLElBQUUsRUFBRSxRQUFPLElBQUUsRUFBRTtBQUFFLFVBQUksSUFBRSxDQUFDO0FBQUUsZUFBUSxJQUFFLEdBQUUsSUFBRSxHQUFFLEtBQUk7QUFBQyxjQUFNLElBQUUsRUFBRSxDQUFDLEdBQUUsSUFBRSxFQUFFO0FBQU8sWUFBRyxFQUFFLFVBQVEsSUFBRSxHQUFFLEdBQUUsR0FBRSxJQUFFLEdBQUUsSUFBSSxLQUFHLElBQUUsRUFBRSxDQUFDLEdBQUUsR0FBRTtBQUFDLGNBQUcsRUFBRSxDQUFDLEdBQUU7QUFBQyxnQkFBRyxDQUFDO0FBQUUsa0JBQUcsRUFBRTtBQUFBLHVCQUFZLEVBQUUsR0FBRyxJQUFFLEdBQUUsTUFBSSxFQUFFLFFBQU87QUFBQTtBQUFFLGdCQUFHLEtBQUcsRUFBRSxHQUFFLENBQUMsSUFBRTtBQUFFLGdCQUFFO0FBQUEsVUFBRTtBQUFDLGNBQUcsTUFBSSxLQUFHLEVBQUUsQ0FBQyxLQUFHLEtBQUcsR0FBRSxFQUFFLENBQUMsSUFBRSxHQUFFLElBQUUsSUFBRztBQUFDLGtCQUFNLElBQUUsRUFBRSxJQUFFLENBQUMsTUFBSSxFQUFFLElBQUUsQ0FBQyxJQUFFLENBQUM7QUFBRyxjQUFFLEVBQUUsTUFBTSxJQUFFO0FBQUEsVUFBQztBQUFBLFFBQUMsTUFBTSxHQUFFLENBQUMsSUFBRTtBQUFBLE1BQUM7QUFBQyxVQUFHLEVBQUUsT0FBSSxJQUFFO0FBQUEsZUFBVyxDQUFDLEVBQUUsUUFBTSxDQUFDO0FBQUUsVUFBRTtBQUFBLElBQUM7QUFBQyxRQUFHLEVBQUUsVUFBUSxJQUFFLEVBQUUsU0FBTyxHQUFFLEdBQUUsR0FBRSxLQUFHLEdBQUUsS0FBSTtBQUFDLFVBQUUsRUFBRSxDQUFDO0FBQUUsVUFBRSxFQUFFO0FBQU8sZUFBUSxJQUFFLEdBQUUsR0FBRSxJQUFFLEdBQUUsSUFBSSxLQUFHLElBQy8zQyxFQUFFLENBQUMsR0FBRSxDQUFDLEVBQUUsQ0FBQyxHQUFFO0FBQUMsWUFBRyxFQUFFO0FBQUEsaUJBQVksRUFBRSxHQUFHLElBQUUsR0FBRSxNQUFJLEVBQUUsUUFBTztBQUFFLFVBQUUsQ0FBQyxJQUFFO0FBQUEsTUFBQztBQUFBLElBQUM7QUFBQyxXQUFPO0FBQUEsRUFBQztBQUFDLFdBQVMsR0FBRyxHQUFFLEdBQUU7QUFBQyxVQUFNLElBQUUsRUFBRSxHQUFFLElBQUUsRUFBRSxHQUFFLElBQUUsQ0FBQztBQUFFLGFBQVEsSUFBRSxHQUFFLElBQUUsRUFBRSxRQUFPLElBQUksR0FBRSxFQUFFLENBQUMsQ0FBQyxJQUFFO0FBQUUsYUFBUSxJQUFFLEdBQUUsR0FBRSxJQUFFLEVBQUUsUUFBTyxLQUFJO0FBQUMsVUFBRSxFQUFFLENBQUM7QUFBRSxlQUFRLElBQUUsR0FBRSxHQUFFLElBQUUsRUFBRSxRQUFPLElBQUksS0FBRSxFQUFFLENBQUMsR0FBRSxFQUFFLENBQUMsS0FBRyxDQUFDLEVBQUUsQ0FBQyxNQUFJLEVBQUUsQ0FBQyxJQUFFLEdBQUUsRUFBRSxFQUFFLE1BQU0sSUFBRTtBQUFBLElBQUU7QUFBQyxXQUFPO0FBQUEsRUFBQztBQUFFLFdBQVMsRUFBRSxHQUFFO0FBQUMsU0FBSyxJQUFFLFNBQUssS0FBRztBQUFFLFNBQUssUUFBTSxFQUFFO0FBQUUsU0FBSyxJQUFFLENBQUM7QUFBQSxFQUFDO0FBQUMsV0FBUyxHQUFHLEdBQUUsR0FBRSxHQUFFO0FBQUMsTUFBRSxDQUFDLE1BQUksSUFBRSxFQUFFO0FBQU8sUUFBSSxJQUFFLEtBQUssTUFBTSxJQUFJLENBQUM7QUFBRSxVQUFJLElBQUUsS0FBSyxPQUFPLEdBQUUsR0FBRSxDQUFDLEdBQUUsS0FBSyxNQUFNLElBQUksR0FBRSxDQUFDO0FBQUcsV0FBTztBQUFB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U0FBSyxJQUFFLEVBQUUsRUFBRSxVQUFVO0FBQUUsU0FBSyxJQUFFLEVBQUUsYUFBVztBQUFFLFNBQUssSUFDeG9ELEVBQUU7QUFBTSxTQUFLLE1BQUksSUFBRSxFQUFFLENBQUMsSUFBRSxFQUFFO0FBQUUsU0FBSyxJQUFFLElBQUUsRUFBRSxjQUFZO0FBQUUsU0FBSyxJQUFFLElBQUUsRUFBRSxDQUFDLElBQUUsRUFBRTtBQUFFLFNBQUssSUFBRSxLQUFHLEVBQUUsS0FBRyxFQUFFO0FBQUksU0FBSyxLQUFHLElBQUUsRUFBRSxXQUFTLEtBQUcsRUFBRSxNQUFJLEVBQUUsR0FBRSxLQUFFO0FBQUUsU0FBSyxLQUFHLElBQUUsRUFBRSxXQUFTLEtBQUcsRUFBRSxNQUFJLEVBQUUsR0FBRSxJQUFFO0FBQUUsUUFBRyxJQUFFLElBQUUsRUFBRSxVQUFRLEtBQUcsRUFBRSxRQUFPO0FBQUMsVUFBRTtBQUFFLFVBQUUsRUFBRTtBQUFFLGVBQVEsSUFBRSxHQUFFLElBQUUsRUFBRSxRQUFPLElBQUUsR0FBRSxJQUFJLEdBQUUsRUFBRSxDQUFDLENBQUMsSUFBRTtBQUFFLFVBQUU7QUFBQSxJQUFDO0FBQUMsU0FBSyxTQUFPO0FBQUUsU0FBSyxTQUFPLElBQUUsRUFBRSxVQUFRLElBQUksRUFBRSxDQUFDO0FBQUEsRUFBQztBQUFDLE1BQUUsRUFBRTtBQUFVLElBQUUsU0FBTyxTQUFTLEdBQUUsR0FBRTtBQUFDLFdBQU8sS0FBSyxJQUFJLEdBQUUsR0FBRSxJQUFFO0FBQUEsRUFBQztBQUN4VyxJQUFFLE1BQUksU0FBUyxHQUFFLEdBQUUsR0FBRSxHQUFFO0FBQUMsUUFBRyxNQUFJLEtBQUcsTUFBSSxJQUFHO0FBQUMsVUFBRyxDQUFDLEtBQUcsQ0FBQyxLQUFHLEtBQUssU0FBUyxDQUFDLEVBQUUsUUFBTyxLQUFLLE9BQU8sR0FBRSxDQUFDO0FBQUUsVUFBRSxLQUFLLE9BQU8sQ0FBQztBQUFFLFVBQUcsSUFBRSxFQUFFLFFBQU87QUFBQyxjQUFNLElBQUUsRUFBRSxHQUFFLElBQUUsRUFBRSxHQUFFLElBQUUsS0FBSyxPQUFNLElBQUUsS0FBSztBQUFFLGlCQUFRLElBQUUsR0FBRSxJQUFFLEdBQUUsS0FBSTtBQUFDLGNBQUksSUFBRSxFQUFFLEtBQUssSUFBRSxJQUFFLElBQUUsSUFBRSxDQUFDO0FBQUUsY0FBSSxJQUFFLEVBQUU7QUFBTyxjQUFHLEtBQUcsS0FBRyxLQUFLLE1BQUksS0FBRyxDQUFDLEVBQUUsQ0FBQyxJQUFHO0FBQUMsZ0JBQUksSUFBRSxFQUFFLEdBQUUsR0FBRSxDQUFDLEdBQUUsSUFBRTtBQUFHLG9CQUFPLEtBQUssR0FBRTtBQUFBLGNBQUMsS0FBSztBQUFPLG9CQUFHLElBQUUsR0FBRTtBQUFDLHVCQUFJLElBQUUsR0FBRSxJQUFFLEdBQUUsSUFBSSxVQUFRLElBQUUsR0FBRSxJQUFFLEdBQUUsSUFBSSxLQUFHLElBQUUsS0FBRyxLQUFLLEdBQUU7QUFBQyx3QkFBSSxJQUFFLEVBQUUsR0FBRSxHQUFFLEdBQUUsR0FBRSxDQUFDO0FBQUUsd0JBQUUsRUFBRSxVQUFVLEdBQUUsQ0FBQztBQUFFLHNCQUFFLE1BQUssR0FBRSxHQUFFLEdBQUUsR0FBRSxDQUFDO0FBQUEsa0JBQUM7QUFBQztBQUFBLGdCQUFLO0FBQUEsY0FBQyxLQUFLO0FBQVUsb0JBQUcsSUFBRSxHQUFFO0FBQUMsdUJBQUksSUFBRSxJQUFFLEdBQUUsSUFBRSxHQUFFLElBQUksS0FBRSxFQUFFLENBQUMsSUFBRSxHQUFFLEVBQUUsVUFBUSxLQUFLLEtBQUc7QUFBQSxvQkFBRTtBQUFBLG9CQUFLO0FBQUEsb0JBQ25mO0FBQUEsb0JBQUUsRUFBRSxHQUFFLEdBQUUsR0FBRSxHQUFFLENBQUM7QUFBQSxvQkFBRTtBQUFBLG9CQUFFO0FBQUEsa0JBQUM7QUFBRSxzQkFBRTtBQUFBLGdCQUFFO0FBQUEsY0FBQyxLQUFLO0FBQVUsb0JBQUcsSUFBRSxHQUFFO0FBQUMsdUJBQUksSUFBRSxHQUFFLElBQUUsR0FBRSxJQUFJLE1BQUcsRUFBRSxDQUFDLEdBQUUsRUFBRSxVQUFRLEtBQUssS0FBRyxFQUFFLE1BQUssR0FBRSxHQUFFLEdBQUUsR0FBRSxDQUFDO0FBQUU7QUFBQSxnQkFBSztBQUFBLGNBQUM7QUFBUSxvQkFBRyxLQUFLLE1BQUksSUFBRSxLQUFLLElBQUksSUFBRSxLQUFLLEVBQUUsR0FBRSxHQUFFLENBQUMsSUFBRSxHQUFFLElBQUUsQ0FBQyxJQUFHLEVBQUUsTUFBSyxHQUFFLEdBQUUsR0FBRSxHQUFFLENBQUMsR0FBRSxLQUFHLElBQUUsS0FBRyxJQUFFLElBQUU7QUFBRSx1QkFBSSxJQUFFLEVBQUUsR0FBRSxJQUFFLEtBQUssR0FBRSxJQUFFLEdBQUUsSUFBRSxLQUFLLElBQUksSUFBRSxHQUFFLElBQUUsQ0FBQyxHQUFFLEVBQUUsQ0FBQyxJQUFFLEdBQUUsSUFBRSxHQUFFLElBQUUsR0FBRSxJQUFJLE1BQUksSUFBRSxFQUFFLEtBQUssSUFBRSxJQUFFLElBQUUsSUFBRSxJQUFFLElBQUUsQ0FBQyxNQUFJLEVBQUUsVUFBUSxLQUFLLEtBQUcsQ0FBQyxFQUFFLENBQUMsR0FBRTtBQUFDLHNCQUFFLENBQUMsSUFBRTtBQUFFLDBCQUFNLElBQUUsS0FBSyxLQUFHLElBQUU7QUFBRSxzQkFBRSxNQUFLLEdBQUUsSUFBRSxJQUFFLEdBQUUsRUFBRSxLQUFHLElBQUUsSUFBRSxJQUFFLElBQUUsSUFBRyxHQUFFLEdBQUUsSUFBRSxHQUFFLElBQUUsQ0FBQyxHQUFFLEdBQUUsR0FBRSxJQUFFLElBQUUsQ0FBQztBQUFBLGtCQUFDO0FBQUE7QUFBQSxZQUFDO0FBQUEsVUFBQztBQUFBLFFBQUM7QUFBQyxhQUFLLE1BQUksS0FBSyxTQUFTLENBQUMsSUFBRTtBQUFBLE1BQUU7QUFBQSxJQUFDO0FBQUMsV0FBTztBQUFBLEVBQUk7QUFDNWIsV0FBUyxFQUFFLEdBQUUsR0FBRSxHQUFFLEdBQUUsR0FBRTtBQUFDLFdBQU8sS0FBRyxJQUFFLElBQUUsS0FBRyxLQUFHLE1BQUksSUFBRSxLQUFHLEtBQUcsTUFBSSxJQUFFLE1BQUksS0FBRyxLQUFHLE9BQUssS0FBRyxLQUFHLE1BQUksSUFBRSxJQUFFO0FBQUEsRUFBQztBQUFDLFdBQVMsRUFBRSxHQUFFLEdBQUUsR0FBRSxHQUFFLEdBQUUsR0FBRSxHQUFFO0FBQUMsUUFBSSxJQUFFLElBQUUsRUFBRSxJQUFFLEVBQUU7QUFBSSxRQUFHLENBQUMsRUFBRSxDQUFDLEtBQUcsS0FBRyxDQUFDLEVBQUUsQ0FBQyxFQUFFLENBQUMsRUFBRSxHQUFFLE1BQUksSUFBRSxFQUFFLENBQUMsSUFBRyxLQUFHLElBQUUsRUFBRSxDQUFDLE1BQUksRUFBRSxDQUFDLElBQUUsRUFBRSxJQUFHLEVBQUUsQ0FBQyxJQUFFLEdBQUUsSUFBRSxFQUFFLENBQUMsTUFBSSxFQUFFLENBQUMsSUFBRSxFQUFFLE1BQUksRUFBRSxDQUFDLElBQUUsR0FBRSxJQUFFLEVBQUUsQ0FBQyxNQUFJLEVBQUUsQ0FBQyxJQUFFLENBQUMsSUFBRyxFQUFFLE1BQUksSUFBRSxFQUFFLENBQUMsTUFBSSxFQUFFLENBQUMsSUFBRSxDQUFDLEtBQUksS0FBRyxFQUFFLFNBQVMsQ0FBQyxNQUFJLEVBQUUsRUFBRSxNQUFNLElBQUUsR0FBRSxFQUFFLE1BQUksSUFBRSxFQUFFLFNBQVMsQ0FBQyxNQUFJLEVBQUUsU0FBUyxDQUFDLElBQUUsQ0FBQyxJQUFHLEVBQUUsRUFBRSxNQUFNLElBQUU7QUFBQSxFQUFHO0FBQ3hXLElBQUUsU0FBTyxTQUFTLEdBQUUsR0FBRSxHQUFFO0FBQUMsVUFBSSxDQUFDLEtBQUcsRUFBRSxDQUFDLEtBQUcsSUFBRSxHQUFFLElBQUUsRUFBRSxTQUFPLEVBQUUsQ0FBQyxNQUFJLElBQUU7QUFBSSxRQUFJLElBQUUsQ0FBQyxHQUFFO0FBQUUsUUFBSSxHQUFFLElBQUU7QUFBRSxRQUFHLEdBQUU7QUFBQyxVQUFFLEVBQUUsU0FBTztBQUFFLFVBQUUsRUFBRTtBQUFNLFVBQUUsRUFBRSxVQUFRO0FBQUUsVUFBSSxJQUFFLEVBQUU7QUFBUSxVQUFFLEVBQUU7QUFBQSxJQUFPO0FBQUMsUUFBRyxNQUFJLElBQUUsS0FBSyxPQUFPLEtBQUcsQ0FBQyxHQUFFLElBQUUsRUFBRSxRQUFPLElBQUUsSUFBRztBQUFDLFVBQUUsRUFBRTtBQUFFLFVBQUksSUFBRSxDQUFDO0FBQUUsZUFBUSxJQUFFLEdBQUUsSUFBRSxHQUFFLEdBQUUsSUFBRSxHQUFFLElBQUksTUFBSSxJQUFFLEVBQUUsQ0FBQyxNQUFJLEVBQUUsVUFBUSxLQUFLLEtBQUcsQ0FBQyxFQUFFLENBQUMsRUFBRSxLQUFHLEtBQUssS0FBRyxLQUFHLEtBQUssSUFBSSxDQUFDLEVBQUUsR0FBRSxHQUFHLElBQUUsR0FBRSxFQUFFLENBQUMsSUFBRTtBQUFBLFVBQU8sUUFBTztBQUFFLFVBQUU7QUFBRSxVQUFFLEVBQUU7QUFBQSxJQUFNO0FBQUMsUUFBRyxDQUFDLEVBQUUsUUFBTztBQUFFLFVBQUksSUFBRTtBQUFLLFFBQUUsS0FBSyxTQUFPLElBQUUsS0FBRyxVQUFLO0FBQUUsUUFBRTtBQUFFLFFBQUk7QUFBRSxTQUFHLElBQUUsRUFBRSxDQUFDLEdBQUUsSUFBRSxLQUFHLElBQUUsS0FBRyxFQUFFLEtBQUssRUFBRTtBQUFFLGFBQVEsR0FBRSxHQUFFLElBQUUsR0FBRSxLQUFJO0FBQUMsVUFBRSxFQUFFLENBQUM7QUFBRSxXQUFHLElBQUU7QUFBQSxRQUFHO0FBQUEsUUFBSztBQUFBLFFBQUU7QUFBQSxRQUFFO0FBQUEsUUFBRTtBQUFBLFFBQUUsTUFBSTtBQUFBLFFBQUU7QUFBQSxRQUNwZjtBQUFBLE1BQUMsR0FBRSxLQUFHLFVBQUssS0FBRyxFQUFFLFdBQVMsSUFBRSxNQUFJLElBQUUsR0FBRyxNQUFLLEdBQUUsR0FBRSxHQUFFLEdBQUUsTUFBSSxHQUFFLENBQUM7QUFBRSxVQUFHLEVBQUUsUUFBTztBQUFFLFVBQUcsS0FBRyxNQUFJLElBQUUsR0FBRTtBQUFDLFlBQUUsRUFBRTtBQUFPLFlBQUcsQ0FBQyxHQUFFO0FBQUMsY0FBRyxHQUFFO0FBQUMsZ0JBQUU7QUFBRSxnQkFBRTtBQUFHO0FBQUEsVUFBUTtBQUFDLGlCQUFPO0FBQUEsUUFBQztBQUFDLFlBQUcsTUFBSSxFQUFFLFFBQU8sR0FBRyxFQUFFLENBQUMsR0FBRSxHQUFFLENBQUM7QUFBQSxNQUFDO0FBQUEsSUFBQztBQUFDLFdBQU8sR0FBRyxHQUFFLEdBQUUsR0FBRSxDQUFDO0FBQUEsRUFBQztBQUMxTCxXQUFTLEdBQUcsR0FBRSxHQUFFLEdBQUUsR0FBRSxHQUFFLEdBQUUsR0FBRSxHQUFFO0FBQUMsUUFBSSxJQUFFLENBQUMsR0FBRSxJQUFFLElBQUUsRUFBRSxJQUFFLEVBQUU7QUFBSSxNQUFFLE1BQUksSUFBRSxHQUFHLEdBQUUsR0FBRSxHQUFFLEVBQUUsQ0FBQztBQUFHLFFBQUcsR0FBRTtBQUFDLFVBQUksSUFBRTtBQUFFLFlBQU0sSUFBRSxLQUFLLElBQUksRUFBRSxRQUFPLElBQUUsRUFBRSxJQUFFLEVBQUUsQ0FBQztBQUFFLGVBQVEsSUFBRSxHQUFFLElBQUUsR0FBRSxHQUFFLEdBQUUsSUFBRSxHQUFFLElBQUksS0FBRyxJQUFFLEVBQUUsQ0FBQztBQUFFLFlBQUcsRUFBRSxNQUFJLElBQUUsR0FBRyxHQUFFLEdBQUUsR0FBRSxFQUFFLENBQUMsSUFBRyxLQUFHLEtBQUcsTUFBSSxJQUFFLEVBQUUsUUFBTyxLQUFHLEtBQUcsS0FBRyxHQUFFLElBQUUsU0FBTyxJQUFFLEVBQUUsTUFBTSxDQUFDLEdBQUUsSUFBRSxLQUFJLE1BQUksRUFBRSxHQUFHLElBQUUsR0FBRSxNQUFJLEtBQUcsRUFBRSxRQUFPLEtBQUcsSUFBSTtBQUFBO0FBQU0sVUFBRyxHQUFFO0FBQUMsWUFBRyxFQUFFLFFBQU8sR0FBRyxHQUFFLEdBQUUsQ0FBQztBQUFFLFVBQUUsRUFBRSxNQUFNLElBQUU7QUFBRTtBQUFBLE1BQU07QUFBQSxJQUFDO0FBQUMsV0FBTSxDQUFDLEtBQUc7QUFBQSxFQUFDO0FBQUMsV0FBUyxHQUFHLEdBQUUsR0FBRSxHQUFFO0FBQUMsUUFBRSxNQUFJLEVBQUUsU0FBTyxFQUFFLENBQUMsSUFBRSxDQUFDLEVBQUUsT0FBTyxNQUFNLENBQUMsR0FBRSxDQUFDO0FBQUUsV0FBTyxLQUFHLEVBQUUsU0FBTyxJQUFFLEVBQUUsTUFBTSxHQUFFLElBQUUsQ0FBQyxJQUFFO0FBQUEsRUFBQztBQUNwYyxXQUFTLEdBQUcsR0FBRSxHQUFFLEdBQUUsR0FBRTtBQUFDLFNBQUcsSUFBRSxLQUFHLElBQUUsR0FBRSxLQUFHLElBQUUsRUFBRSxJQUFFLElBQUUsQ0FBQyxNQUFJLEVBQUUsSUFBRSxJQUFFLENBQUMsS0FBRyxJQUFFLEVBQUUsQ0FBQztBQUFFLFdBQU87QUFBQSxFQUFDO0FBQUMsSUFBRSxVQUFRLFNBQVMsR0FBRTtBQUFDLFdBQU0sQ0FBQyxDQUFDLEtBQUssU0FBUyxDQUFDO0FBQUEsRUFBQztBQUFFLElBQUUsU0FBTyxTQUFTLEdBQUUsR0FBRTtBQUFDLFdBQU8sS0FBSyxPQUFPLENBQUMsRUFBRSxJQUFJLEdBQUUsQ0FBQztBQUFBLEVBQUM7QUFDaEwsSUFBRSxTQUFPLFNBQVMsR0FBRSxHQUFFO0FBQUMsVUFBTSxJQUFFLEtBQUssU0FBUyxDQUFDO0FBQUUsUUFBRyxHQUFFO0FBQUMsVUFBRyxLQUFLLEVBQUUsVUFBUSxJQUFFLEdBQUUsR0FBRSxJQUFFLEVBQUUsUUFBTyxJQUFJLEtBQUUsRUFBRSxDQUFDLEdBQUUsRUFBRSxPQUFPLEVBQUUsUUFBUSxDQUFDLEdBQUUsQ0FBQztBQUFBLFVBQU8sR0FBRSxLQUFLLEtBQUksR0FBRSxLQUFLLEdBQUUsS0FBSyxDQUFDLEdBQUUsS0FBSyxTQUFPLEVBQUUsS0FBSyxHQUFFLEdBQUUsS0FBSyxHQUFFLEtBQUssQ0FBQztBQUFFLFdBQUcsT0FBTyxLQUFLLFNBQVMsQ0FBQztBQUFFLFVBQUcsS0FBSyxPQUFNO0FBQUMsWUFBRSxLQUFLO0FBQU0saUJBQVEsSUFBRSxHQUFFLEdBQUUsR0FBRSxJQUFFLEVBQUUsRUFBRSxRQUFPLElBQUksS0FBRSxFQUFFLEVBQUUsQ0FBQyxHQUFFLElBQUUsRUFBRSxNQUFNLENBQUMsR0FBRSxFQUFFLFNBQVMsQ0FBQyxNQUFJLEVBQUUsRUFBRSxPQUFPLEtBQUksQ0FBQyxHQUFFLE9BQU8sRUFBRSxNQUFNLENBQUM7QUFBQSxNQUFFO0FBQUEsSUFBQztBQUFDLFdBQU87QUFBQSxFQUFJO0FBQ25YLFdBQVMsRUFBRSxHQUFFLEdBQUUsR0FBRSxHQUFFLEdBQUU7QUFBQyxRQUFJLElBQUU7QUFBRSxRQUFHLEVBQUUsZ0JBQWMsTUFBTSxLQUFHLEVBQUUsS0FBRSxFQUFFLFFBQVEsQ0FBQyxHQUFFLE9BQUssSUFBRSxJQUFFLEVBQUUsV0FBUyxFQUFFLE9BQU8sR0FBRSxDQUFDLEdBQUUsT0FBSztBQUFBLFNBQVE7QUFBQyxVQUFFLEtBQUssSUFBSSxFQUFFLFFBQU8sQ0FBQztBQUFFLGVBQVEsSUFBRSxHQUFFLEdBQUUsSUFBRSxHQUFFLElBQUksS0FBRyxJQUFFLEVBQUUsQ0FBQyxFQUFFLEtBQUUsRUFBRSxHQUFFLEdBQUUsR0FBRSxHQUFFLENBQUMsR0FBRSxLQUFHLEtBQUcsT0FBTyxFQUFFLENBQUM7QUFBQSxJQUFDO0FBQUEsUUFBTSxVQUFRLEtBQUssRUFBRSxFQUFDLElBQUUsRUFBRSxFQUFFLENBQUMsR0FBRSxHQUFFLEdBQUUsR0FBRSxDQUFDLE1BQUksT0FBTyxFQUFFLENBQUM7QUFBRSxXQUFPO0FBQUEsRUFBQztBQUFDLElBQUUsY0FBWTtBQUMvUixJQUFFLFNBQU8sU0FBUyxHQUFFLEdBQUUsR0FBRSxHQUFFLEdBQUUsR0FBRTtBQUFDLFFBQUksSUFBRTtBQUFHLG9CQUFjLE9BQU8sTUFBSSxJQUFFLElBQUksUUFBUSxPQUFHO0FBQUMsVUFBRTtBQUFBLElBQUMsQ0FBQztBQUFHLFFBQUksR0FBRTtBQUFFLFlBQU8sTUFBSSxJQUFFLElBQUc7QUFBQSxNQUFDLEtBQUs7QUFBRSxZQUFFO0FBQU0sWUFBRyxLQUFLLEdBQUU7QUFBQyxjQUFFLEVBQUU7QUFBRSxtQkFBUSxLQUFLLEtBQUssU0FBUyxHQUFFLENBQUMsSUFBRTtBQUFBLFFBQUMsTUFBTSxLQUFFLEtBQUs7QUFBUztBQUFBLE1BQU0sS0FBSztBQUFFLFlBQUU7QUFBTSxZQUFFLEVBQUMsS0FBSSxHQUFFLEtBQUksS0FBSyxJQUFFLElBQUUsRUFBQztBQUFFO0FBQUEsTUFBTSxLQUFLO0FBQUUsWUFBRTtBQUFNLFlBQUUsS0FBSztBQUFJO0FBQUEsTUFBTSxLQUFLO0FBQUUsWUFBRTtBQUFNLFlBQUUsS0FBSztBQUFFO0FBQUEsTUFBTTtBQUFRLHdCQUFjLE9BQU8sS0FBRyxLQUFHLEVBQUU7QUFBRTtBQUFBLElBQU07QUFBQyxPQUFHLEdBQUUsS0FBRyxNQUFLLEdBQUUsR0FBRSxHQUFFLEdBQUUsR0FBRSxDQUFDO0FBQUUsV0FBTztBQUFBLEVBQUM7QUFDNVksSUFBRSxTQUFPLFNBQVMsR0FBRSxHQUFFO0FBQUMsUUFBRyxFQUFFLFNBQU8sRUFBRSxDQUFDLE1BQUksSUFBRSxLQUFLLE1BQU0sQ0FBQyxJQUFHLEdBQUU7QUFBQSxNQUFDLEtBQUs7QUFBTSxhQUFLLElBQUUsQ0FBQyxDQUFDLEVBQUU7QUFBSTtBQUFBLE1BQU0sS0FBSztBQUFNLGFBQUssSUFBRTtBQUFHLGFBQUssV0FBUztBQUFFO0FBQUEsTUFBTSxLQUFLO0FBQU0sYUFBSyxNQUFJO0FBQUU7QUFBQSxNQUFNLEtBQUs7QUFBTSxhQUFLLElBQUU7QUFBQSxJQUFDO0FBQUEsRUFBQztBQUFFLEtBQUcsRUFBRSxTQUFTO0FBQUUsV0FBUyxHQUFHLEdBQUU7QUFBQyxRQUFFLEVBQUU7QUFBSyxRQUFJLElBQUUsS0FBSztBQUFPLFVBQU0sSUFBRSxFQUFFO0FBQUssUUFBSSxJQUFFLEVBQUU7QUFBSyxZQUFPLEdBQUU7QUFBQSxNQUFDLEtBQUs7QUFBTyxZQUFFLEVBQUUsV0FBUyxDQUFDO0FBQUUsWUFBRSxFQUFFO0FBQVEsWUFBRSxFQUFFO0FBQU8sVUFBRSxRQUFNO0FBQUcsYUFBRyxNQUFJLEVBQUUsUUFBUSxVQUFVLE1BQUksRUFBRSxTQUFPLFNBQVMsWUFBVSxDQUFDLEVBQUU7QUFBRyxhQUFHLFNBQVMsWUFBVSxDQUFDLEVBQUUsRUFBRSxJQUFJLEdBQUUsS0FBSyxTQUFPLElBQUksS0FBSyxXQUFXLE1BQU0sQ0FBQyxHQUFFLE9BQU8sS0FBSyxjQUFZLEtBQUssU0FBTyxJQUFJLEVBQUUsQ0FBQztBQUFFO0FBQUEsTUFBTTtBQUFRLFlBQUUsRUFBRSxJQUFHLElBQUUsRUFBRSxDQUFDLEVBQUUsTUFBTSxHQUFFLENBQUMsR0FBRSxZQUFZLGFBQVcsSUFBRSxFQUFDLElBQUcsR0FBRSxLQUFJLEVBQUMsSUFBRSxFQUFDLElBQUcsRUFBQyxDQUFDO0FBQUEsSUFBQztBQUFBLEVBQUM7QUFBRSxNQUFJLEtBQUc7QUFBRSxXQUFTLEVBQUUsR0FBRTtBQUFDLFFBQUcsRUFBRSxnQkFBZ0IsR0FBRyxRQUFPLElBQUksRUFBRSxDQUFDO0FBQUUsUUFBSTtBQUFFLFFBQUUsRUFBRSxJQUFFLEVBQUUsTUFBTSxNQUFJLEVBQUUsU0FBTyxFQUFFLFNBQVMsS0FBRyxJQUFFLENBQUM7QUFBRSxLQUFDLEtBQUcsUUFBTSxRQUFRLGNBQVksSUFBRSxFQUFFLFNBQVM7QUFBRyxVQUFNLElBQUUsZ0JBQWMsT0FBTyxVQUFRLEtBQUssU0FBUSxJQUFFO0FBQUssU0FBSyxJQUFFLEdBQUcsR0FBRSxHQUFFLEVBQUUsTUFBTTtBQUFFLFNBQUssSUFBRSxFQUFFO0FBQUUsUUFBRyxLQUFLLEdBQUU7QUFBQyxVQUFHLEVBQUUsTUFBSyxFQUFFLEdBQUcsV0FBVSxTQUFTLEdBQUU7QUFBQyxVQUFFLEVBQUUsRUFBRSxFQUFFLEVBQUUsRUFBRSxHQUFHO0FBQUUsZUFBTyxFQUFFLEVBQUUsRUFBRSxFQUFFO0FBQUEsTUFBQyxDQUFDO0FBQUEsVUFBTyxNQUFLLEVBQUUsWUFBVSxTQUFTLEdBQUU7QUFBQyxZQUFFLEVBQUU7QUFBSyxVQUFFLEVBQUUsRUFBRSxFQUFFLEVBQUUsRUFBRSxHQUFHO0FBQUUsZUFBTyxFQUFFLEVBQUUsRUFBRSxFQUFFO0FBQUEsTUFBQztBQUFFLFdBQUssRUFBRSxZQUFZLEVBQUMsTUFBSyxRQUFPLFNBQVEsR0FBRSxTQUFRLEVBQUMsQ0FBQztBQUFBLElBQUM7QUFBQSxFQUFDO0FBQUMsSUFBRSxLQUFLO0FBQUUsSUFBRSxRQUFRO0FBQUUsSUFBRSxRQUFRO0FBQzdsQyxJQUFFLFFBQVE7QUFBRSxJQUFFLFFBQVE7QUFBRSxXQUFTLEVBQUUsR0FBRTtBQUFDLE1BQUUsVUFBVSxDQUFDLElBQUUsRUFBRSxVQUFVLElBQUUsT0FBTyxJQUFFLFdBQVU7QUFBQyxZQUFNLElBQUUsTUFBSyxJQUFFLENBQUMsRUFBRSxNQUFNLEtBQUssU0FBUztBQUFFLFVBQUksSUFBRSxFQUFFLEVBQUUsU0FBTyxDQUFDO0FBQUUsVUFBSTtBQUFFLFFBQUUsQ0FBQyxNQUFJLElBQUUsR0FBRSxFQUFFLE9BQU8sRUFBRSxTQUFPLEdBQUUsQ0FBQztBQUFHLFVBQUUsSUFBSSxRQUFRLFNBQVMsR0FBRTtBQUFDLG1CQUFXLFdBQVU7QUFBQyxZQUFFLEVBQUUsRUFBRSxFQUFFLElBQUU7QUFBRSxZQUFFLEVBQUUsWUFBWSxFQUFDLE1BQUssR0FBRSxJQUFHLElBQUcsTUFBSyxFQUFDLENBQUM7QUFBQSxRQUFDLENBQUM7QUFBQSxNQUFDLENBQUM7QUFBRSxhQUFPLEtBQUcsRUFBRSxLQUFLLENBQUMsR0FBRSxRQUFNO0FBQUEsSUFBQztBQUFBLEVBQUM7QUFDL1QsV0FBUyxHQUFHLEdBQUUsR0FBRSxHQUFFO0FBQUMsUUFBSTtBQUFFLFFBQUc7QUFBQyxVQUFFLElBQUUsSUFBSywyQkFBMEIsUUFBUSxFQUFHLFlBQVksZUFBZSxJQUFFLElBQUUsSUFBSSxPQUFPLElBQUksZ0JBQWdCLElBQUksS0FBSyxDQUFDLGVBQWEsR0FBRyxTQUFTLENBQUMsR0FBRSxFQUFDLE1BQUssa0JBQWlCLENBQUMsQ0FBQyxDQUFDLElBQUUsSUFBSSxPQUFPLEVBQUUsQ0FBQyxJQUFFLElBQUUsb0JBQW1CLEVBQUMsTUFBSyxTQUFRLENBQUM7QUFBQSxJQUFDLFNBQU8sR0FBRTtBQUFBLElBQUM7QUFBQyxXQUFPO0FBQUEsRUFBQztBQUFFLFdBQVMsRUFBRSxHQUFFO0FBQUMsUUFBRyxFQUFFLGdCQUFnQixHQUFHLFFBQU8sSUFBSSxFQUFFLENBQUM7QUFBRSxRQUFJLElBQUUsRUFBRSxZQUFVLEVBQUUsT0FBSyxHQUFFO0FBQUUsU0FBSyxJQUFFLENBQUM7QUFBRSxTQUFLLElBQUUsQ0FBQztBQUFFLFNBQUssSUFBRSxDQUFDO0FBQUUsU0FBSyxXQUFTLEVBQUU7QUFBRSxTQUFLLE9BQUssSUFBRSxFQUFFLE9BQUssRUFBRSxPQUFLLEVBQUUsR0FBRSxLQUFLLENBQUMsS0FBRztBQUFLLFNBQUssSUFBRSxFQUFFLEVBQUUsVUFBVTtBQUFFLFNBQUssS0FBRyxJQUFFLEVBQUUsVUFBUSxTQUFLLEtBQUcsQ0FBQztBQUFFLFNBQUssUUFBTSxLQUFHLEVBQUU7QUFBRSxTQUFLLEtBQUcsSUFBRSxFQUFFLFFBQU0sRUFBRSxHQUFFLEtBQUssQ0FBQztBQUFFLFNBQUssSUFBRSxLQUFHLEVBQUU7QUFBRSxTQUFLLFNBQU8sSUFBRSxFQUFFLFVBQVEsSUFBSSxFQUFFLENBQUM7QUFBRSxNQUFFLFFBQU07QUFBRyxTQUFLLElBQUUsRUFBRTtBQUFPLFNBQUssUUFBTTtBQUFHLFFBQUUsRUFBRTtBQUFFLFFBQUksSUFBRSxFQUFFLFNBQU8sRUFBRSxTQUFPO0FBQUUsTUFBRSxDQUFDLE1BQUksSUFBRSxDQUFDLENBQUM7QUFBRyxhQUFRLElBQUUsR0FBRSxHQUFFLEdBQUUsSUFBRSxFQUFFLFFBQU8sSUFBSSxLQUFFLEVBQUUsQ0FBQyxHQUFFLEVBQUUsQ0FBQyxNQUFJLElBQUUsR0FBRSxJQUFFLEVBQUUsUUFBTyxJQUFFLEVBQUUsQ0FBQyxJQUFFLE9BQU8sT0FBTyxDQUFDLEdBQUUsR0FBRSxDQUFDLElBQUUsR0FDendCLEtBQUssTUFBSSxFQUFFLENBQUMsSUFBRSxJQUFJLEVBQUUsQ0FBQyxHQUFFLEVBQUUsQ0FBQyxFQUFFLE1BQUksS0FBSyxJQUFFLFNBQUssS0FBSyxNQUFJLEVBQUUsQ0FBQyxJQUFFLElBQUksRUFBRSxHQUFFLEtBQUssUUFBUSxJQUFHLEtBQUssRUFBRSxDQUFDLElBQUUsRUFBRSxHQUFFLEtBQUssQ0FBQyxHQUFFLEtBQUssRUFBRSxDQUFDLElBQUU7QUFBRSxRQUFHLEtBQUssRUFBRSxNQUFJLElBQUUsRUFBRSxPQUFNLEVBQUUsQ0FBQyxNQUFJLElBQUUsQ0FBQyxDQUFDLElBQUcsSUFBRSxHQUFFLElBQUUsRUFBRSxRQUFPLElBQUksTUFBSyxFQUFFLENBQUMsSUFBRSxFQUFFLEVBQUUsQ0FBQyxHQUFFLEtBQUssQ0FBQztBQUFFLFNBQUssUUFBTTtBQUFBLEVBQUM7QUFBQyxXQUFTLEVBQUUsR0FBRSxHQUFFO0FBQUMsVUFBTSxJQUFFLEVBQUUsTUFBTSxHQUFHO0FBQUUsUUFBSSxJQUFFO0FBQUUsYUFBUSxJQUFFLEdBQUUsSUFBRSxFQUFFLFFBQU8sSUFBSSxLQUFFLEVBQUUsQ0FBQyxHQUFFLEtBQUcsRUFBRSxRQUFRLElBQUksTUFBSSxJQUFFLEVBQUUsVUFBVSxHQUFFLEVBQUUsU0FBTyxDQUFDLE9BQUssRUFBRSxDQUFDLElBQUUsT0FBSSxNQUFJLEVBQUUsR0FBRyxJQUFFO0FBQUcsUUFBRSxFQUFFLFdBQVMsRUFBRSxTQUFPO0FBQUcsV0FBTyxJQUFFLElBQUUsSUFBRSxFQUFFLENBQUM7QUFBQSxFQUFDO0FBQUMsV0FBUyxFQUFFLEdBQUUsR0FBRTtBQUFDLFFBQUcsRUFBRSxDQUFDLEVBQUUsS0FBRSxFQUFFLENBQUM7QUFBQSxRQUFPLFVBQVEsSUFBRSxHQUFFLEtBQUcsSUFBRSxFQUFFLFFBQU8sSUFBSSxLQUFFLEVBQUUsRUFBRSxDQUFDLENBQUM7QUFBRSxXQUFPO0FBQUEsRUFBQztBQUM1ZSxXQUFTLEVBQUUsR0FBRSxHQUFFLEdBQUUsR0FBRSxHQUFFO0FBQUMsUUFBRSxFQUFFLENBQUM7QUFBRSxRQUFHLE1BQUksRUFBRSxTQUFPLEVBQUUsR0FBRSxDQUFDLElBQUU7QUFBQSxhQUFVLEVBQUUsS0FBRyxFQUFFLGdCQUFjLE1BQU0sTUFBSSxJQUFFLEVBQUUsQ0FBQyxJQUFFLE1BQU0sRUFBRSxNQUFNLEdBQUUsSUFBRSxHQUFFLElBQUUsRUFBRSxRQUFPLElBQUksR0FBRSxHQUFFLEdBQUUsR0FBRSxHQUFFLENBQUM7QUFBQSxRQUFPLEtBQUUsRUFBRSxDQUFDLE1BQUksRUFBRSxDQUFDLElBQUUsRUFBRSxJQUFHLElBQUUsRUFBRSxFQUFFLENBQUMsR0FBRSxFQUFFLEdBQUUsR0FBRSxHQUFFLEdBQUUsQ0FBQztBQUFBLEVBQUM7QUFBQyxXQUFTLEVBQUUsR0FBRSxHQUFFLEdBQUUsR0FBRSxHQUFFLEdBQUUsR0FBRSxHQUFFO0FBQUMsUUFBRyxJQUFFLEVBQUUsQ0FBQyxFQUFFLEtBQUcsTUFBSSxFQUFFLFNBQU8sR0FBRTtBQUFDLFVBQUcsRUFBRSxnQkFBYyxPQUFNO0FBQUMsWUFBRyxFQUFFLENBQUMsR0FBRTtBQUFDLGVBQUksSUFBRSxHQUFFLElBQUUsRUFBRSxRQUFPLElBQUksR0FBRSxJQUFJLEdBQUUsRUFBRSxDQUFDLEdBQUUsTUFBRyxJQUFFO0FBQUU7QUFBQSxRQUFNO0FBQUMsWUFBRSxFQUFFLEtBQUssR0FBRztBQUFBLE1BQUM7QUFBQyxRQUFFLElBQUksR0FBRSxHQUFFLEdBQUUsSUFBRTtBQUFBLElBQUMsV0FBUyxFQUFFLGdCQUFjLE1BQU0sTUFBSSxJQUFFLEdBQUUsSUFBRSxFQUFFLFFBQU8sSUFBSSxHQUFFLEdBQUUsR0FBRSxHQUFFLEdBQUUsR0FBRSxHQUFFLEdBQUUsQ0FBQztBQUFBLFFBQU8sS0FBRSxFQUFFLEVBQUUsQ0FBQyxHQUFFLEVBQUUsR0FBRSxHQUFFLEdBQUUsR0FBRSxHQUFFLEdBQUUsR0FBRSxDQUFDO0FBQUEsRUFBQztBQUFDLE1BQUUsRUFBRTtBQUMzZCxJQUFFLE1BQUksU0FBUyxHQUFFLEdBQUUsR0FBRTtBQUFDLE1BQUUsQ0FBQyxNQUFJLElBQUUsR0FBRSxJQUFFLEVBQUUsR0FBRSxLQUFLLEdBQUc7QUFBRyxRQUFHLE1BQUksS0FBRyxNQUFJLElBQUc7QUFBQyxVQUFHLENBQUMsS0FBRyxLQUFLLFNBQVMsQ0FBQyxFQUFFLFFBQU8sS0FBSyxPQUFPLEdBQUUsQ0FBQztBQUFFLGVBQVEsSUFBRSxHQUFFLEdBQUUsR0FBRSxJQUFFLEtBQUssRUFBRSxRQUFPLElBQUksS0FBRSxLQUFLLEVBQUUsQ0FBQyxHQUFFLElBQUUsS0FBSyxFQUFFLENBQUMsR0FBRSxFQUFFLENBQUMsTUFBSSxJQUFFLENBQUMsQ0FBQyxJQUFHLEVBQUUsR0FBRSxHQUFFLEtBQUssR0FBRSxHQUFFLEtBQUssTUFBTSxDQUFDLEdBQUUsR0FBRSxFQUFFLENBQUMsR0FBRSxDQUFDO0FBQUUsVUFBRyxLQUFLLEdBQUU7QUFBQyxZQUFJLElBQUUsRUFBRSxHQUFFLEtBQUssQ0FBQyxHQUFFLElBQUUsRUFBRTtBQUFFLFVBQUUsQ0FBQyxNQUFJLElBQUUsQ0FBQyxDQUFDO0FBQUcsaUJBQVEsSUFBRSxHQUFFLEdBQUUsR0FBRSxJQUFFLEVBQUUsUUFBTyxJQUFJLEtBQUcsSUFBRSxFQUFFLENBQUMsR0FBRSxDQUFDLEVBQUUsQ0FBQyxNQUFJLEVBQUUsQ0FBQyxJQUFFLEdBQUUsSUFBRSxLQUFLLEVBQUUsQ0FBQyxNQUFJLEtBQUssRUFBRSxDQUFDLElBQUUsQ0FBQyxJQUFHLENBQUMsS0FBRyxDQUFDLEVBQUUsU0FBUyxDQUFDO0FBQUcsY0FBRyxFQUFFLEVBQUUsTUFBTSxJQUFFLEdBQUUsS0FBSyxHQUFFO0FBQUMsa0JBQU0sSUFBRSxLQUFLLFNBQVMsQ0FBQyxNQUFJLEtBQUssU0FBUyxDQUFDLElBQUUsQ0FBQztBQUFHLGNBQUUsRUFBRSxNQUFNLElBQUU7QUFBQSxVQUFDO0FBQUE7QUFBQSxNQUFDO0FBQUMsVUFBRyxLQUFLLFVBQVEsQ0FBQyxLQUFHLENBQUMsS0FBSyxNQUFNLENBQUMsSUFBRztBQUFDLFlBQUk7QUFDL2YsWUFBRyxLQUFLLEdBQUU7QUFBQyxjQUFFLEVBQUU7QUFBRSxtQkFBUSxJQUFFLEdBQUUsR0FBRSxJQUFFLEtBQUssRUFBRSxRQUFPLElBQUksS0FBRSxLQUFLLEVBQUUsQ0FBQyxHQUFFLEVBQUUsQ0FBQyxJQUFFLEVBQUUsQ0FBQyxJQUFFLEVBQUUsQ0FBQyxJQUFFLEVBQUUsR0FBRSxHQUFFLEdBQUUsR0FBRSxFQUFFLENBQUMsQ0FBQztBQUFBLFFBQUM7QUFBQyxhQUFLLE1BQU0sQ0FBQyxJQUFFLEtBQUc7QUFBQSxNQUFDO0FBQUEsSUFBQztBQUFDLFdBQU87QUFBQSxFQUFJO0FBQUUsSUFBRSxTQUFPLFNBQVMsR0FBRSxHQUFFO0FBQUMsV0FBTyxLQUFLLElBQUksR0FBRSxHQUFFLElBQUU7QUFBQSxFQUFDO0FBQUUsSUFBRSxTQUFPLFNBQVMsR0FBRSxHQUFFO0FBQUMsV0FBTyxLQUFLLE9BQU8sQ0FBQyxFQUFFLElBQUksR0FBRSxDQUFDO0FBQUEsRUFBQztBQUNwTyxJQUFFLFNBQU8sU0FBUyxHQUFFO0FBQUMsTUFBRSxDQUFDLE1BQUksSUFBRSxFQUFFLEdBQUUsS0FBSyxHQUFHO0FBQUcsUUFBRyxLQUFLLFNBQVMsQ0FBQyxHQUFFO0FBQUMsZUFBUSxJQUFFLEdBQUUsSUFBRSxLQUFLLEVBQUUsV0FBUyxLQUFLLE1BQU0sS0FBSyxFQUFFLENBQUMsQ0FBQyxFQUFFLE9BQU8sR0FBRSxDQUFDLEtBQUssQ0FBQyxHQUFFLENBQUMsS0FBSyxJQUFHLElBQUk7QUFBQyxVQUFHLEtBQUssS0FBRyxDQUFDLEtBQUssRUFBRSxVQUFRLEtBQUssS0FBSyxHQUFFO0FBQUMsWUFBRSxLQUFLLEVBQUUsQ0FBQztBQUFFLGNBQU0sSUFBRSxFQUFFLFFBQVEsQ0FBQztBQUFFLGVBQUssTUFBSSxJQUFFLEVBQUUsU0FBTyxFQUFFLE9BQU8sR0FBRSxDQUFDLElBQUUsT0FBTyxLQUFLLEVBQUUsQ0FBQztBQUFBLE1BQUU7QUFBQyxXQUFLLFNBQU8sT0FBTyxLQUFLLE1BQU0sQ0FBQztBQUFFLGFBQU8sS0FBSyxTQUFTLENBQUM7QUFBQSxJQUFDO0FBQUMsV0FBTztBQUFBLEVBQUk7QUFDdlYsSUFBRSxTQUFPLFNBQVMsR0FBRSxHQUFFLEdBQUUsR0FBRTtBQUFDLFVBQUksQ0FBQyxLQUFHLEVBQUUsQ0FBQyxLQUFHLElBQUUsR0FBRSxJQUFFLE1BQUksRUFBRSxDQUFDLE1BQUksSUFBRSxHQUFFLElBQUU7QUFBSSxRQUFJLElBQUUsQ0FBQyxHQUFFLElBQUUsQ0FBQyxHQUFFLEdBQUUsR0FBRSxHQUFFLEdBQUUsR0FBRSxHQUFFLElBQUU7QUFBRSxRQUFHLEVBQUUsS0FBRyxFQUFFLGdCQUFjLE1BQU0sS0FBRSxHQUFFLElBQUU7QUFBQSxTQUFTO0FBQUMsVUFBRSxFQUFFLFNBQU87QUFBRSxXQUFHLElBQUUsRUFBRSxVQUFRLEVBQUUsU0FBTyxFQUFFO0FBQU0sVUFBRSxFQUFFO0FBQUksVUFBRSxLQUFLLFNBQU8sRUFBRTtBQUFPLFVBQUUsVUFBUSxFQUFFO0FBQUssVUFBRSxFQUFFLFNBQU8sS0FBRztBQUFJLFVBQUUsRUFBRSxVQUFRO0FBQUUsVUFBRyxNQUFJLEVBQUUsQ0FBQyxNQUFJLElBQUUsQ0FBQyxDQUFDLElBQUcsQ0FBQyxJQUFHO0FBQUMsaUJBQVEsSUFBRSxHQUFFLEdBQUUsSUFBRSxFQUFFLFFBQU8sSUFBSSxLQUFHLElBQUUsR0FBRyxLQUFLLE1BQUssRUFBRSxDQUFDLEdBQUUsR0FBRSxHQUFFLENBQUMsRUFBRSxHQUFFLEVBQUUsTUFBTSxJQUFFLEdBQUU7QUFBSSxlQUFPLElBQUUsSUFBRSxDQUFDO0FBQUEsTUFBQztBQUFDLFFBQUUsQ0FBQyxNQUFJLElBQUUsQ0FBQyxDQUFDO0FBQUEsSUFBRTtBQUFDLFVBQUksSUFBRSxLQUFLO0FBQUcsUUFBRSxNQUFJLElBQUUsRUFBRSxVQUFRLEtBQUcsSUFBRSxFQUFFO0FBQVEsVUFBTSxJQUFFLENBQUMsTUFBSSxLQUFLLEtBQUcsS0FBSyxVQUFRLENBQUM7QUFBRSxhQUFRLElBQUUsR0FBRSxHQUFFLEdBQUUsR0FBRSxJQUN0ZixFQUFFLFFBQU8sS0FBSTtBQUFDLFVBQUk7QUFBRSxVQUFFLEVBQUUsQ0FBQztBQUFFLFFBQUUsQ0FBQyxNQUFJLElBQUUsR0FBRSxJQUFFLEVBQUUsT0FBTSxJQUFFLEVBQUUsU0FBTyxHQUFFLElBQUUsRUFBRSxTQUFPLEdBQUUsSUFBRSxFQUFFLFVBQVE7QUFBRyxVQUFHLEVBQUUsR0FBRSxDQUFDLElBQUUsS0FBSyxNQUFNLENBQUMsRUFBRSxZQUFZLEdBQUUsR0FBRSxLQUFHLENBQUM7QUFBQSxXQUFNO0FBQUMsWUFBRSxJQUFFLEVBQUUsQ0FBQyxJQUFFLElBQUUsS0FBSyxNQUFNLENBQUMsRUFBRSxPQUFPLEdBQUUsR0FBRSxLQUFHLENBQUM7QUFBRSxZQUFFLEtBQUcsRUFBRTtBQUFPLFlBQUcsS0FBRyxHQUFFO0FBQUMsZ0JBQU0sSUFBRSxDQUFDO0FBQUUsY0FBSSxJQUFFO0FBQUUsZ0JBQUksRUFBRSxDQUFDLElBQUUsQ0FBQyxDQUFDO0FBQUcsbUJBQVEsSUFBRSxHQUFFLElBQUcsR0FBRSxJQUFFLEVBQUUsUUFBTyxJQUFJLEtBQUcsS0FBRyxFQUFFLENBQUMsR0FBRSxLQUFHLElBQUUsS0FBSyxFQUFFLEVBQUUsTUFBSSxFQUFFLE9BQU8sTUFBSSxFQUFFLEVBQUUsTUFBTSxJQUFFLElBQUUsQ0FBQyxDQUFDLElBQUU7QUFBRSxnQkFBSSxJQUFFLElBQUUsR0FBRyxHQUFFLEtBQUcsS0FBSSxLQUFHLENBQUMsSUFBRSxHQUFHLEdBQUUsQ0FBQyxHQUFFLElBQUUsRUFBRTtBQUFBLFFBQU87QUFBQyxZQUFHLEVBQUUsR0FBRSxDQUFDLElBQUUsR0FBRSxFQUFFLEdBQUcsSUFBRTtBQUFBLGlCQUFVLEVBQUUsUUFBTSxDQUFDO0FBQUEsTUFBQztBQUFBLElBQUM7QUFBQyxRQUFHLEdBQUU7QUFBQyxZQUFNLElBQUU7QUFBSyxhQUFPLElBQUksUUFBUSxTQUFTLEdBQUU7QUFBQyxnQkFBUSxJQUFJLENBQUMsRUFBRSxLQUFLLFNBQVMsR0FBRTtBQUFDLFlBQUUsRUFBRTtBQUFBLFlBQU87QUFBQSxZQUNoZ0I7QUFBQSxZQUFFO0FBQUEsWUFBRTtBQUFBLFVBQUMsQ0FBQztBQUFBLFFBQUMsQ0FBQztBQUFBLE1BQUMsQ0FBQztBQUFBLElBQUM7QUFBQyxRQUFHLENBQUMsRUFBRSxRQUFNLENBQUM7QUFBRSxRQUFHLE1BQUksQ0FBQyxLQUFHLENBQUMsS0FBSyxPQUFPLFFBQU8sRUFBRSxDQUFDO0FBQUUsYUFBUSxJQUFFLEdBQUUsR0FBRSxJQUFFLEVBQUUsUUFBTyxLQUFJO0FBQUMsVUFBRSxFQUFFLENBQUM7QUFBRSxRQUFFLFVBQVEsTUFBSSxJQUFFLEdBQUcsS0FBSyxNQUFLLENBQUM7QUFBRyxVQUFHLEVBQUUsUUFBTztBQUFFLFFBQUUsQ0FBQyxJQUFFLEVBQUMsT0FBTSxFQUFFLENBQUMsR0FBRSxRQUFPLEVBQUM7QUFBQSxJQUFDO0FBQUMsV0FBTztBQUFBLEVBQUM7QUFBRSxXQUFTLEdBQUcsR0FBRSxHQUFFLEdBQUUsR0FBRTtBQUFDLFFBQUksSUFBRSxLQUFLLEVBQUUsQ0FBQyxHQUFFLElBQUUsS0FBRyxFQUFFLFNBQU87QUFBRSxRQUFHLEtBQUcsSUFBRSxHQUFFO0FBQUMsVUFBRyxJQUFFLEtBQUcsRUFBRSxLQUFFLEVBQUUsTUFBTSxHQUFFLElBQUUsQ0FBQztBQUFFLFlBQUksSUFBRSxHQUFHLEtBQUssTUFBSyxDQUFDO0FBQUcsYUFBTSxFQUFDLEtBQUksR0FBRSxRQUFPLEVBQUM7QUFBQSxJQUFDO0FBQUEsRUFBQztBQUFDLFdBQVMsR0FBRyxHQUFFO0FBQUMsVUFBTSxJQUFFLE1BQU0sRUFBRSxNQUFNO0FBQUUsYUFBUSxJQUFFLEdBQUUsR0FBRSxJQUFFLEVBQUUsUUFBTyxJQUFJLEtBQUUsRUFBRSxDQUFDLEdBQUUsRUFBRSxDQUFDLElBQUUsRUFBQyxJQUFHLEdBQUUsS0FBSSxLQUFLLE1BQU0sQ0FBQyxFQUFDO0FBQUUsV0FBTztBQUFBLEVBQUM7QUFBQyxJQUFFLFVBQVEsU0FBUyxHQUFFO0FBQUMsV0FBTSxDQUFDLENBQUMsS0FBSyxTQUFTLENBQUM7QUFBQSxFQUFDO0FBQUUsSUFBRSxNQUFJLFNBQVMsR0FBRTtBQUFDLFdBQU8sS0FBSyxNQUFNLENBQUM7QUFBQSxFQUFDO0FBQzFnQixJQUFFLE1BQUksU0FBUyxHQUFFLEdBQUU7QUFBQyxTQUFLLE1BQU0sQ0FBQyxJQUFFO0FBQUUsV0FBTztBQUFBLEVBQUk7QUFBRSxJQUFFLGNBQVk7QUFBRyxJQUFFLFNBQU8sU0FBUyxHQUFFLEdBQUUsR0FBRSxHQUFFLEdBQUUsR0FBRTtBQUFDLFFBQUk7QUFBRSxvQkFBYyxPQUFPLE1BQUksSUFBRSxJQUFJLFFBQVEs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LE1BQUssQ0FBQyxFQUFFLEtBQUssR0FBRztBQUFFLFVBQU0sSUFBRSxDQUFDO0FBQUUsUUFBRyxHQUFFO0FBQUMsWUFBTSxJQUFFLEVBQUUsTUFBTSxFQUFFLEdBQUUsSUFBRSxFQUFFO0FBQU8sZUFBUSxJQUFFLEdBQUUsR0FBRSxJQUFFLEdBQUUsSUFBRSxHQUFFLElBQUksTUFBSSxJQUFFLEVBQUUsQ0FBQyxPQUFLLENBQUMsS0FBSyxVQUFRLENBQUMsS0FBSyxPQUFPLENBQUMsSUFBRztBQUFDLFlBQUUsRUFBRSxDQUFDO0FBQUUsWUFBSSxJQUFFLEdBQUcsQ0FBQyxLQUFHLEdBQUUsSUFBRTtBQUFFLGlCQUFRLElBQUUsR0FBRSxJQUFFLEVBQUUsUUFBTyxLQUFJO0FBQUMsY0FBRSxFQUFFLENBQUM7QUFBRSxnQkFBTSxJQUFFLEdBQUcsQ0FBQyxLQUFHO0FBQUUsZUFBRyxNQUFJLE1BQUksS0FBRyxHQUFFLElBQUU7QUFBQSxRQUFFO0FBQUMsVUFBRSxHQUFHLElBQUU7QUFBQSxNQUFDO0FBQUEsSUFBQztBQUFDLFdBQU87QUFBQSxFQUFDO0FBQUUsTUFBSSxLQUFHLEVBQUMsUUFBTyxJQUFHLEdBQUUsT0FBRyxHQUFFLEdBQUU7QUFBRSxNQUFNLEtBQUcsQ0FBQyxFQUFFLElBQUksR0FBRSxLQUFJLEVBQUUsSUFBSSxHQUFFLEtBQUksRUFBRSxJQUFJLEdBQUUsS0FBSSxFQUFFLElBQUksR0FBRSxLQUFJLEVBQUUsSUFBSSxHQUFFLEtBQUksRUFBRSxJQUFJLEdBQUUsS0FBSSxFQUFFLHFCQUFxQixHQUFFLElBQUcsRUFBRSx1QkFBdUIsR0FBRSxFQUFFO0FBQUUsV0FBUyxHQUFHLEdBQUUsR0FBRTtBQUFDLFVBQUksSUFBRSxHQUFHLEtBQUssTUFBSyxDQUFDLEVBQUUsS0FBSyxHQUFHLEdBQUUsSUFBRSxFQUFFLFdBQVMsSUFBRSxFQUFFLEdBQUUsRUFBRSxJQUFHLE1BQUksSUFBRSxFQUFFLFdBQVMsSUFBRSxHQUFHLENBQUMsSUFBRyxNQUFJLElBQUUsRUFBRSxNQUFNLEdBQUc7QUFB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
