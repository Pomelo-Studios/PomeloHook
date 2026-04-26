package dashboard

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
)

//go:embed static
var staticFiles embed.FS

func Serve(apiHandler http.Handler) {
	distFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		log.Fatalf("dashboard embed error: %v", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/api/", apiHandler)
	mux.Handle("/", http.FileServer(http.FS(distFS)))

	go func() {
		log.Printf("Dashboard: http://localhost:4040")
		if err := http.ListenAndServe(":4040", mux); err != nil {
			log.Printf("dashboard server error: %v", err)
		}
	}()
}
