package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/pomelo-studios/pomelo-hook/cli/config"
	"github.com/spf13/cobra"
	"golang.org/x/term"
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
	fmt.Print("Password: ")
	passBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("failed to read password: %w", err)
	}
	fmt.Println()

	payload, err := json.Marshal(map[string]string{"email": email, "password": string(passBytes)})
	if err != nil {
		return fmt.Errorf("failed to encode request: %w", err)
	}
	resp, err := http.Post(serverURL+"/api/auth/login", "application/json", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("cannot reach server: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("login failed: %s", string(msg))
	}
	var result struct {
		APIKey string `json:"api_key"`
		Name   string `json:"name"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&result); err != nil {
		return err
	}
	if result.APIKey == "" || result.Name == "" {
		return fmt.Errorf("server returned incomplete credentials")
	}
	cfg := &config.Config{ServerURL: serverURL, APIKey: result.APIKey, UserName: result.Name}
	if err := config.Save(cfg); err != nil {
		return err
	}
	fmt.Printf("Logged in as %s. Config saved to ~/.pomelo-hook/config.json\n", result.Name)
	return nil
}
