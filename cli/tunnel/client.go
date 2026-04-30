package tunnel

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pomelo-studios/pomelo-hook/cli/forward"
)

var wsDialer = &websocket.Dialer{
	HandshakeTimeout: 10 * time.Second,
	Proxy:            http.ProxyFromEnvironment,
}

type Client struct {
	serverURL string
	apiKey    string
	tunnelID  string
	device    string
	forwarder *forward.Forwarder
	onEvent   func(result *forward.ForwardResult)
	sem       chan struct{}
	rng       *rand.Rand
}

type Options struct {
	ServerURL string
	APIKey    string
	TunnelID  string
	LocalPort string
	Device    string
	OnEvent   func(*forward.ForwardResult)
}

func New(opts Options) *Client {
	return &Client{
		serverURL: opts.ServerURL,
		apiKey:    opts.APIKey,
		tunnelID:  opts.TunnelID,
		device:    opts.Device,
		forwarder: forward.New("http://localhost:" + opts.LocalPort),
		onEvent:   opts.OnEvent,
		sem:       make(chan struct{}, 8),
		rng:       rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (c *Client) Connect() error {
	wsURL := strings.Replace(c.serverURL, "http", "ws", 1) + "/api/ws?tunnel_id=" + c.tunnelID
	if c.device != "" {
		wsURL += "&device=" + url.QueryEscape(c.device)
	}
	headers := http.Header{"Authorization": {"Bearer " + c.apiKey}}

	var attempt int
	for {
		conn, _, err := wsDialer.Dial(wsURL, headers)
		if err != nil {
			attempt++
			if attempt > 5 {
				return fmt.Errorf("could not connect after 5 attempts: %w", err)
			}
			wait := time.Duration(1<<attempt) * time.Second
			jitter := time.Duration(c.rng.Int63n(int64(wait / 2)))
			log.Printf("reconnecting in %s...", wait+jitter)
			time.Sleep(wait + jitter)
			continue
		}
		attempt = 0
		log.Println("tunnel connected")
		if err := c.pump(conn); err != nil {
			log.Printf("tunnel disconnected: %v", err)
		}
	}
}

func (c *Client) pump(conn *websocket.Conn) error {
	defer conn.Close()
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			return err
		}
		var ack map[string]string
		if json.Unmarshal(msg, &ack) == nil && ack["status"] == "connected" {
			continue
		}
		c.sem <- struct{}{}
		go func(payload []byte) {
			defer func() { <-c.sem }()
			result, err := c.forwarder.Forward(payload)
			if err != nil {
				log.Printf("forward error: %v", err)
			}
			if c.onEvent != nil && result != nil {
				c.onEvent(result)
			}
		}(msg)
	}
}
