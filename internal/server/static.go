package server

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

// dist holds the built frontend (ui/ → `npm run build` copies it here).
// A fresh clone only has .gitkeep — the server then serves a build hint page.
//
//go:embed all:dist
var dist embed.FS

const tokenPlaceholder = "__LOCKBOX_TOKEN__"

func (s *Server) handleStatic(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" || wantsHTML(r) {
		s.serveIndex(w)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/")
	if strings.Contains(path, "..") {
		http.NotFound(w, r)
		return
	}
	data, err := fs.ReadFile(dist, "dist/"+path)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if strings.HasPrefix(path, "assets/") {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	}
	w.Header().Set("Content-Type", contentTypeFor(path))
	_, _ = w.Write(data)
}

// wantsHTML reports whether the request is a page navigation (SPA fallback).
func wantsHTML(r *http.Request) bool {
	accept := r.Header.Get("Accept")
	return r.URL.Path != "/" &&
		!strings.HasPrefix(r.URL.Path, "assets/") &&
		strings.Contains(accept, "text/html")
}

func (s *Server) serveIndex(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	data, err := fs.ReadFile(dist, "dist/index.html")
	if err != nil {
		_, _ = w.Write([]byte(notBuiltPage))
		return
	}
	// Replace only the quoted placeholder value so the property name survives.
	html := strings.ReplaceAll(string(data), `"`+tokenPlaceholder+`"`, `"`+s.token+`"`)
	_, _ = w.Write([]byte(html))
}

func contentTypeFor(path string) string {
	switch {
	case strings.HasSuffix(path, ".js"):
		return "text/javascript; charset=utf-8"
	case strings.HasSuffix(path, ".css"):
		return "text/css; charset=utf-8"
	case strings.HasSuffix(path, ".svg"):
		return "image/svg+xml"
	case strings.HasSuffix(path, ".png"):
		return "image/png"
	case strings.HasSuffix(path, ".ico"):
		return "image/x-icon"
	case strings.HasSuffix(path, ".woff2"):
		return "font/woff2"
	default:
		return "application/octet-stream"
	}
}

const notBuiltPage = `<!doctype html>
<html lang="en">
<head><meta charset="utf-8"><title>lockbox</title>
<style>
  body{background:#09090b;color:#fafafa;font-family:system-ui,sans-serif;display:flex;align-items:center;justify-content:center;height:100vh;margin:0}
  .card{background:#18181b;border:1px solid #27272a;border-radius:12px;padding:32px 40px;max-width:460px;text-align:center}
  code{background:#27272a;padding:2px 8px;border-radius:6px;font-size:14px}
  p{color:#a1a1aa;line-height:1.6}
</style></head>
<body><div class="card">
<h1 style="margin-top:0">lockbox UI not built</h1>
<p>Frontend assets are missing. Build them once:</p>
<p><code>make ui</code></p>
<p>or manually: <code>cd ui &amp;&amp; npm install &amp;&amp; npm run build</code><br>then rebuild the Go binary.</p>
</div></body>
</html>`
