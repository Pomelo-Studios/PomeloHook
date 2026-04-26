package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/pomelo-studios/pomelo-hook/cli/config"
	"github.com/spf13/cobra"
)

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
		return fmt.Errorf("not logged in — run: pomelo-hook login")
	}

	url := fmt.Sprintf("%s/api/events?limit=%d", cfg.ServerURL, lastN)
	if tunnelIDFlag != "" {
		url += "&tunnel_id=" + tunnelIDFlag
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var events []struct {
		ID             string    `json:"ID"`
		Method         string    `json:"Method"`
		Path           string    `json:"Path"`
		ReceivedAt     time.Time `json:"ReceivedAt"`
		ResponseStatus int       `json:"ResponseStatus"`
		Forwarded      bool      `json:"Forwarded"`
	}
	json.NewDecoder(resp.Body).Decode(&events)

	for _, e := range events {
		status := "✗"
		if e.Forwarded {
			status = "✓"
		}
		fmt.Printf("[%s] %s %s %s → %d (%s)\n",
			e.ID[:8], status, e.Method, e.Path,
			e.ResponseStatus, e.ReceivedAt.Format("15:04:05"))
	}
	return nil
}
