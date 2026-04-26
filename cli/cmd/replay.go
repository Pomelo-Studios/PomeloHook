package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pomelo-studios/pomelo-hook/cli/config"
	"github.com/spf13/cobra"
)

var replayCmd = &cobra.Command{
	Use:   "replay <event-id>",
	Short: "Replay a webhook event",
	Args:  cobra.ExactArgs(1),
	RunE:  runReplay,
}

var replayTarget string

func init() {
	replayCmd.Flags().StringVar(&replayTarget, "to", "http://localhost:3000", "Target URL for replay")
}

func runReplay(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("not logged in — run: pomelo-hook login")
	}
	eventID := args[0]

	payload, err := json.Marshal(map[string]string{"target_url": replayTarget})
	if err != nil {
		return fmt.Errorf("failed to encode request: %w", err)
	}
	req, err := http.NewRequest("POST", cfg.ServerURL+"/api/events/"+eventID+"/replay", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result struct {
		StatusCode int   `json:"status_code"`
		ResponseMS int64 `json:"response_ms"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	fmt.Printf("Replayed %s → %d (%dms)\n", eventID, result.StatusCode, result.ResponseMS)
	return nil
}
