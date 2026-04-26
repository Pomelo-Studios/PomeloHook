package cmd

import (
	"fmt"
	"log"

	"github.com/pomelo-studios/pomelo-hook/cli/config"
	"github.com/pomelo-studios/pomelo-hook/cli/forward"
	"github.com/pomelo-studios/pomelo-hook/cli/tunnel"
	"github.com/spf13/cobra"
)

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
		return fmt.Errorf("not logged in — run: pomelo-hook login")
	}

	tunnelID, subdomain, err := resolveTunnel(cfg, orgTunnel, orgTunnelName)
	if err != nil {
		return err
	}

	fmt.Printf("Tunnel: %s/webhook/%s → localhost:%s\n", cfg.ServerURL, subdomain, localPort)
	fmt.Println("Dashboard: http://localhost:4040")
	fmt.Println("Press Ctrl+C to stop")

	client := tunnel.New(tunnel.Options{
		ServerURL: cfg.ServerURL,
		APIKey:    cfg.APIKey,
		TunnelID:  tunnelID,
		LocalPort: localPort,
		OnEvent: func(r *forward.ForwardResult) {
			log.Printf("→ %s [%d] %dms", r.EventID, r.StatusCode, r.MS)
		},
	})
	return client.Connect()
}

func resolveTunnel(cfg *config.Config, isOrg bool, tunnelName string) (id, subdomain string, err error) {
	return "", "", fmt.Errorf("not yet implemented")
}
