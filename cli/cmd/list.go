package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/pomelo-studios/pomelo-hook/cli/config"
	"github.com/spf13/cobra"
)

func pickFirstTunnelID(cfg *config.Config) (string, error) {
	req, err := http.NewRequest("GET", cfg.ServerURL+"/api/tunnels", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	resp, err := apiClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch tunnels: server returned %d", resp.StatusCode)
	}
	var tunnels []struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tunnels); err != nil {
		return "", fmt.Errorf("failed to decode tunnels: %w", err)
	}
	if len(tunnels) == 0 {
		return "", fmt.Errorf("no tunnels found — run 'pomelo-hook connect' first")
	}
	return tunnels[0].ID, nil
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List recent webhook events",
	RunE:  runList,
}

var lastN int
var tunnelIDFlag string

func init() {
	listCmd.Flags().IntVar(&lastN, "last", 20, "Number of recent events to show")
	listCmd.Flags().StringVar(&tunnelIDFlag, "tunnel", "", "Tunnel ID to filter by")
}

func runList(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return errNotLoggedIn
	}

	tunnelID := tunnelIDFlag
	if tunnelID == "" {
		tunnelID, err = pickFirstTunnelID(cfg)
		if err != nil {
			return err
		}
	}

	url := fmt.Sprintf("%s/api/events?limit=%d&tunnel_id=%s", cfg.ServerURL, lastN, tunnelID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	resp, err := apiClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned %d", resp.StatusCode)
	}

	var events []struct {
		ID             string    `json:"ID"`
		Method         string    `json:"Method"`
		Path           string    `json:"Path"`
		ReceivedAt     time.Time `json:"ReceivedAt"`
		ResponseStatus int       `json:"ResponseStatus"`
		Forwarded      bool      `json:"Forwarded"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&events); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if len(events) == 0 {
		fmt.Println("No events found.")
		return nil
	}

	for _, e := range events {
		status := "✗"
		if e.Forwarded {
			status = "✓"
		}
		id := e.ID
		if len(id) > 8 {
			id = id[:8]
		}
		fmt.Printf("[%s] %s %s %s → %d (%s)\n",
			id, status, e.Method, e.Path,
			e.ResponseStatus, e.ReceivedAt.Format("15:04:05"))
	}
	return nil
}
