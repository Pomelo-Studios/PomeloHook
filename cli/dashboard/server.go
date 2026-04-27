package dashboard

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"strings"
)

//go:embed static
var staticFiles embed.FS

func Serve(apiHandler http.Handler) {
	distFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		log.Fatalf("dashboard embed error: %v", err)
	}

	indexHTML, err := fs.ReadFile(distFS, "index.html")
	if err != nil {
		log.Fatalf("dashboard index.html missing: %v", err)
	}
	fileServer := http.FileServer(http.FS(distFS))
	spa := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/assets/") {
			fileServer.ServeHTTP(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(indexHTML)
	})

	mux := http.NewServeMux()
	mux.Handle("/api/", apiHandler)
	mux.Handle("/", spa)

	go func() {
		log.Printf("Dashboard: http://localhost:4040")
		if err := http.ListenAndServe(":4040", mux); err != nil {
			log.Printf("dashboard server error: %v", err)
		}
	}()
}
