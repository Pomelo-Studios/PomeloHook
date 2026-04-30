package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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
	mux.Handle("/webhook/", api.LoggingMiddleware(webhookHandler))
	dh := dashboardHandler()
	mux.Handle("/admin", dh)
	mux.Handle("/admin/", dh)
	mux.Handle("/app", dh)
	mux.Handle("/app/", dh)
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

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-quit
		log.Println("shutdown signal received")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		ticker.Stop()
		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("shutdown error: %v", err)
		}
	}()

	log.Printf("PomeloHook server listening on :%s", cfg.Port)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("listen: %v", err)
	}
}
