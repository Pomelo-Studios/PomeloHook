// cli/cmd/login.go
package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
	"github.com/pomelo-studios/pomelo-hook/cli/config"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with a PomeloHook server",
	RunE:  runLogin,
}

var serverURL string
var email string

func init() {
	loginCmd.Flags().StringVar(&serverURL, "server", "", "PomeloHook server URL (required)")
	loginCmd.Flags().StringVar(&email, "email", "", "Your email address (required)")
	loginCmd.MarkFlagRequired("server")
	loginCmd.MarkFlagRequired("email")
}

func runLogin(cmd *cobra.Command, args []string) error {
	payload, _ := json.Marshal(map[string]string{"email": email})
	resp, err := http.Post(serverURL+"/api/auth/login", "application/json", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("cannot reach server: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("login failed: server returned %d", resp.StatusCode)
	}
	var result struct {
		APIKey string `json:"api_key"`
		Name   string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}
	cfg := &config.Config{ServerURL: serverURL, APIKey: result.APIKey, UserName: result.Name}
	if err := config.Save(cfg); err != nil {
		return err
	}
	fmt.Printf("Logged in as %s. Config saved to ~/.pomelo-hook/config.json\n", result.Name)
	return nil
}
