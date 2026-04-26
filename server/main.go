package main

import (
	"log"
	"net/http"
	"time"

	"github.com/pomelo-studios/pomelo-hook/server/api"
	"github.com/pomelo-studios/pomelo-hook/server/config"
	"github.com/pomelo-studios/pomelo-hook/server/store"
	"github.com/pomelo-studios/pomelo-hook/server/tunnel"
	wh "github.com/pomelo-studios/pomelo-hook/server/webhook"
)

func main() {
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

	go func() {
		for range time.Tick(24 * time.Hour) {
			db.DeleteEventsOlderThan(cfg.RetentionDays)
		}
	}()

	log.Printf("PomeloHook server listening on :%s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, mux))
}
