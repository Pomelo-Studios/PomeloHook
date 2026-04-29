package api

import (
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/pomelo-studios/pomelo-hook/server/store"
)

func TestReplayHTTP_Timeout(t *testing.T) {
	// TCP server that accepts but never sends an HTTP response
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				buf := make([]byte, 4096)
				for {
					if _, err := c.Read(buf); err != nil {
						return
					}
				}
			}(conn)
		}
	}()

	event := &store.WebhookEvent{
		Method:      http.MethodPost,
		RequestBody: `{}`,
	}

	start := time.Now()
	_, _, err = replayHTTP(event, "http://"+ln.Addr().String())
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if elapsed > 20*time.Second {
		t.Fatalf("replay took too long: %v (expected ~15s timeout)", elapsed)
	}
}
