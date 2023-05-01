package cmd

import (
	"bufio"
	"fmt"
	"gochat/app"
	"gochat/prompts"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(shCmd)
}

var shCmd = &cobra.Command{
	Use:     "sh",
	Short:   "Generate a bash command for the given prompt",
	Example: `gochat sh -4 list files in the current directory`,
	Run: func(cmd *cobra.Command, args []string) {
		gptOutput := app.Run(prompts.PromptBashCommand+strings.Join(args, " "), model)

		if askForConfirmation("\nRun this command?") {
			fmt.Println("Running command...")
			fmt.Println(gptOutput)
			cmd := exec.Command("bash", "-c", gptOutput)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Run()
		}
	},
}

func askForConfirmation(s string) bool {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("%s [Y/n]: ", s)

		response, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		response = strings.ToLower(strings.TrimSpace(response))

		if response == "y" || response == "yes" || response == "" {
			return true
		} else if response == "n" || response == "no" {
			return false
		}
	}
}
