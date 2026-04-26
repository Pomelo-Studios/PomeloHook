package cmd

import "github.com/spf13/cobra"

var connectCmd = &cobra.Command{
	Use:   "connect",
	Short: "Open a tunnel and forward webhooks to a local port",
}
