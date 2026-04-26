package cmd

import "github.com/spf13/cobra"

var replayCmd = &cobra.Command{
	Use:   "replay",
	Short: "Replay a webhook event",
}
