package server

import (
	"embed"
	"io/fs"
	"net/http"
	"path"
	"strings"

	khttp "github.com/go-kratos/kratos/v2/transport/http"
)

//go:embed web web/assets/*
var webFS embed.FS

var contentTypeMap = map[string]string{
	".html": "text/html; charset=utf-8",
	".js":   "application/javascript",
	".css":  "text/css",
	".png":  "image/png",
	".jpg":  "image/jpeg",
	".svg":  "image/svg+xml",
	".ico":  "image/x-icon",
	".json": "application/json",
}

func registerWebUI(srv *khttp.Server) {
	subFS, err := fs.Sub(webFS, "web")
	if err != nil {
		panic("embed FS sub: " + err.Error())
	}
	fileServer := http.FileServer(http.FS(subFS))

	srv.HandlePrefix("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := strings.TrimPrefix(r.URL.Path, "/")

		// Handle root & index.html explicitly to avoid FileServer redirect loop
		if p == "" || p == "index.html" {
			data, err := webFS.ReadFile("web/index.html")
			if err != nil {
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write(data)
			return
		}

		_, err := fs.Stat(subFS, p)
		if err != nil {
			// SPA fallback: serve index.html
			data, err := webFS.ReadFile("web/index.html")
			if err != nil {
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write(data)
			return
		}

		// Serve actual file
		ext := path.Ext(p)
		if ct, ok := contentTypeMap[ext]; ok {
			w.Header().Set("Content-Type", ct)
		}
		w.Header().Set("Cache-Control", "public, max-age=3600")
		r.URL.Path = "/" + p
		fileServer.ServeHTTP(w, r)
	}))
}
