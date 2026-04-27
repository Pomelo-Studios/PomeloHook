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

	fileServer := http.FileServer(http.FS(distFS))
	spa := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if !strings.HasPrefix(path, "/assets/") {
			path = "/"
		}
		r2 := r.Clone(r.Context())
		r2.URL.Path = path
		fileServer.ServeHTTP(w, r2)
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
