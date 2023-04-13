package cmd

import (
	"fmt"
	"gochat/app"
	"gochat/prompts"
	"os"

	"github.com/spf13/cobra"
)

var model string = "gpt-3.5-turbo"

func init() {
	rootCmd.PersistentFlags().StringP("prompt", "p", "", "Prompt before the pipe input")
	rootCmd.PersistentFlags().StringP("model", "m", "3.5", "[3.5 | 4] Use a specific model gpt-3.5-turbo or gpt-4")

	rootCmd.PersistentFlags().Bool("as-command", false, "Return the response as a bash command")

	rootCmd.CompletionOptions.DisableDefaultCmd = true
}

var rootCmd = &cobra.Command{
	Use:   "gochat",
	Short: "gochat is a command line interface for OpenAI ChatGPT models 3.5 and 4",
	PreRun: func(cmd *cobra.Command, args []string) {
		if cmd.Flag("model").Value.String() == "4" {
			model = "gpt-4"
		}

		if cmd.Flag("as-command").Value.String() == "true" {
			cmd.Flag("prompt").Value.Set(prompts.PromptBashCommand + cmd.Flag("prompt").Value.String())
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
