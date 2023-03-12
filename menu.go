package main

import (
	"fmt"
	"strings"
)

func menu(userInput string) string {
	switch strings.Split(userInput, " ")[0] {
	case "/exit":
		return "break"
	case "/reset":
		messages = []Message{}
		return "continue"
	case "/history":
		spl := strings.Split(userInput, " ")
		if len(spl) == 2 && spl[1] == "on" {
			options.History = true
		} else if len(spl) == 2 && spl[1] == "off" {
			options.History = false
		} else {
			fmt.Printf("History: %v\n", options.History)
		}
		return "continue"
	case "/debug":
		spl := strings.Split(userInput, " ")
		if len(spl) == 2 && spl[1] == "on" {
			options.Debug = true
		} else if len(spl) == 2 && spl[1] == "off" {
			options.Debug = false
		} else {
			fmt.Printf("Debug: %v\n", options.Debug)
		}
		return "continue"
	case "/key":
		fmt.Printf("API Key: %s\n", key)
		return "continue"
	case "/len":
		acc := 0
		for _, m := range messages {
			acc += len(m.Content)
		}
		fmt.Printf("Messages: %d\n", len(messages))
		fmt.Printf("  Tokens: %d\n", acc/4)
		fmt.Printf("    Cost: $%.6f (@ $0.002/1k tokens)\n", float64(acc/4)/1000*0.002)
		return "continue"
	case "/help":
		fmt.Println("\nAvailable commands:")
		fmt.Println()
		fmt.Println("  /reset .............. Reset the conversation (when history is on)")
		fmt.Println("  /len ................ Show the length and the cost of the conversation")
		fmt.Println("  /key ................ Show the API key")
		fmt.Println("  /debug [on|off] ..... Show each chunk of the response (default: off)")
		fmt.Println("  /history [on|off] ... Toggles the history (default: off)")
		fmt.Println("  /exit ............... Exit the program")
		return "continue"
	default:
		return ""
	}
}
