package cmd

import (
	"gochat/app"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(setupCmd)
}

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup the OpenAI API key",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		app.Setup()
	},
}
