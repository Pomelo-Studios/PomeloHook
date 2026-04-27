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
	fileServer := http.FileServer(http.FS(sub))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if strings.HasPrefix(path, "/admin") {
			path = strings.TrimPrefix(path, "/admin")
			if path == "" {
				path = "/"
			}
		}
		if !strings.HasPrefix(path, "/assets/") && path != "/index.html" {
			path = "/index.html"
		}
		r2 := r.Clone(r.Context())
		r2.URL.Path = path
		fileServer.ServeHTTP(w, r2)
	})
}
