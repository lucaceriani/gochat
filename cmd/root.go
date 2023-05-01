package cmd

import (
	"fmt"
	"gochat/app"
	"os"

	"github.com/spf13/cobra"
)

var model string = "gpt-3.5-turbo"

func init() {
	rootCmd.Flags().StringP("prompt", "p", "", "Prompt before the pipe input")
	rootCmd.PersistentFlags().BoolP("gpt-4", "4", false, "Use gpt-4")

	rootCmd.CompletionOptions.DisableDefaultCmd = true
}

var rootCmd = &cobra.Command{
	Use:   "gochat",
	Short: "gochat is a command line interface for OpenAI ChatGPT models 3.5 and 4",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if cmd.Flag("gpt-4").Changed {
			model = "gpt-4"
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		app.Run(cmd.Flag("prompt").Value.String(), model)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
