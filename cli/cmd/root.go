package cmd

import (
	"errors"

	"github.com/spf13/cobra"
)

var errNotLoggedIn = errors.New("not logged in — run: pomelo-hook login")

var rootCmd = &cobra.Command{
	Use:   "pomelo-hook",
	Short: "PomeloHook — self-hosted webhook relay",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(connectCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(replayCmd)
}
