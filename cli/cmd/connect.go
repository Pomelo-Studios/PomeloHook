package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/pomelo-studios/pomelo-hook/cli/config"
	"github.com/pomelo-studios/pomelo-hook/cli/dashboard"
	"github.com/pomelo-studios/pomelo-hook/cli/forward"
	"github.com/pomelo-studios/pomelo-hook/cli/tunnel"
	"github.com/spf13/cobra"
)

var apiClient = &http.Client{Timeout: 15 * time.Second}

var connectCmd = &cobra.Command{
	Use:   "connect",
	Short: "Open a webhook tunnel and forward to a local port",
	RunE:  runConnect,
}

var localPort string
var orgTunnel bool
var orgTunnelName string

func init() {
	connectCmd.Flags().StringVar(&localPort, "port", "3000", "Local port to forward to")
	connectCmd.Flags().BoolVar(&orgTunnel, "org", false, "Connect to an org tunnel")
	connectCmd.Flags().StringVar(&orgTunnelName, "tunnel", "", "Org tunnel name (required with --org)")
}

func runConnect(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return errNotLoggedIn
	}

	tunnelID, subdomain, err := resolveTunnel(cfg, orgTunnel, orgTunnelName)
	if err != nil {
		return err
	}

	fmt.Printf("Tunnel: %s/webhook/%s → localhost:%s\n", cfg.ServerURL, subdomain, localPort)
	fmt.Println("Dashboard: http://localhost:4040")
	fmt.Println("Press Ctrl+C to stop")

	dashboard.Serve(newLocalAPIProxy(cfg.ServerURL, cfg.APIKey))

	hostname, _ := os.Hostname()
	client := tunnel.New(tunnel.Options{
		ServerURL: cfg.ServerURL,
		APIKey:    cfg.APIKey,
		TunnelID:  tunnelID,
		LocalPort: localPort,
		Device:    hostname,
		OnEvent: func(r *forward.ForwardResult) {
			log.Printf("→ %s [%d] %dms", r.EventID, r.StatusCode, r.MS)
		},
	})
	return client.Connect()
}

func resolveTunnel(cfg *config.Config, isOrg bool, tunnelName string) (id, subdomain string, err error) {
	tunnelType := "personal"
	if isOrg {
		tunnelType = "org"
	}

	payload, err := json.Marshal(map[string]string{"type": tunnelType, "name": tunnelName})
	if err != nil {
		return "", "", fmt.Errorf("failed to encode request: %w", err)
	}
	req, err := http.NewRequest("POST", cfg.ServerURL+"/api/tunnels", bytes.NewReader(payload))
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := apiClient.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("cannot reach server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusConflict {
		return "", "", fmt.Errorf("org tunnel '%s' is already active", tunnelName)
	}
	if resp.StatusCode != http.StatusCreated {
		return "", "", fmt.Errorf("failed to create tunnel: %d", resp.StatusCode)
	}

	var tun struct {
		ID        string `json:"ID"`
		Subdomain string `json:"Subdomain"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&tun); err != nil {
		return "", "", err
	}
	if tun.ID == "" || tun.Subdomain == "" {
		return "", "", fmt.Errorf("server returned incomplete tunnel data")
	}
	return tun.ID, tun.Subdomain, nil
}

func newLocalAPIProxy(serverURL, apiKey string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		target := serverURL + r.URL.RequestURI()
		req, err := http.NewRequest(r.Method, target, r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		req.Header = r.Header.Clone()
		req.Header.Set("Authorization", "Bearer "+apiKey)
		resp, err := apiClient.Do(req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()
		for k, v := range resp.Header {
			w.Header()[k] = v
		}
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
	})
}
