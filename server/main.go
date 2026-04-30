package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/pomelo-studios/pomelo-hook/server/api"
	"github.com/pomelo-studios/pomelo-hook/server/config"
	"github.com/pomelo-studios/pomelo-hook/server/store"
	"github.com/pomelo-studios/pomelo-hook/server/tunnel"
	wh "github.com/pomelo-studios/pomelo-hook/server/webhook"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "init" {
		if err := runInit(); err != nil {
			log.Fatalf("init: %v", err)
		}
		return
	}

	cfg := config.Load()

	db, err := store.Open(cfg.DBPath)
	if err != nil {
		log.Fatalf("open store: %v", err)
	}
	defer db.Close()

	mgr := tunnel.NewManager()
	router := api.NewRouter(db, mgr)
	webhookHandler := wh.NewHandler(db, mgr)

	mux := http.NewServeMux()
	mux.Handle("/api/", router)
	mux.Handle("/webhook/", webhookHandler)
	dh := dashboardHandler()
	mux.Handle("/admin", dh)
	mux.Handle("/admin/", dh)
	mux.Handle("/assets/", dh)
	mux.HandleFunc("/favicon.svg", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	ticker := time.NewTicker(24 * time.Hour)
	go func() {
		cleanup := func() {
			if _, err := db.DeleteEventsOlderThan(cfg.RetentionDays); err != nil {
				log.Printf("retention cleanup error: %v", err)
			}
		}
		cleanup()
		for range ticker.C {
			cleanup()
		}
	}()

	log.Printf("PomeloHook server listening on :%s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, mux))
}
