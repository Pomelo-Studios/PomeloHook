package main

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed dashboard/static
var dashboardFiles embed.FS

func dashboardHandler() http.Handler {
	sub, err := fs.Sub(dashboardFiles, "dashboard/static")
	if err != nil {
		panic("dashboard embed misconfigured: " + err.Error())
	}
	indexHTML, err := fs.ReadFile(sub, "index.html")
	if err != nil {
		panic("dashboard index.html missing: " + err.Error())
	}
	fileServer := http.FileServer(http.FS(sub))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if strings.HasPrefix(path, "/admin") {
			path = strings.TrimPrefix(path, "/admin")
		}
		if strings.HasPrefix(path, "/assets/") {
			r2 := r.Clone(r.Context())
			r2.URL.Path = path
			fileServer.ServeHTTP(w, r2)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(indexHTML)
	})
}
