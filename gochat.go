package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatSend struct {
	Model     string    `json:"model"`
	Stream    bool      `json:"stream"`
	MaxTokens int       `json:"max_tokens"` // 1024
	Messages  []Message `json:"messages"`
}

type ChatResponseEvent struct {
	Id      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Delta        Message `json:"delta"`
		Index        int     `json:"index"`
		FinishReason string  `json:"finish_reason"`
	} `json:"choices"`
}

type Options struct {
	NoHistory bool
}

func getKey() string {
	key, err := os.ReadFile("api.key")
	if err != nil {
		panic(err)
	}

	return strings.TrimSpace(string(key))
}

func readAndPrintReponse(res *http.Response) Message {
	defer res.Body.Close()

	var message Message

	toParseNext := ""
	tmpBuf := make([]byte, 4096)

	for {
		n, err := res.Body.Read(tmpBuf)
		if err != nil {
			break
		}

		toParseNext += string(tmpBuf[:n])

		// get the index of the first \n\n
		idxNN := strings.Index(toParseNext, "\n\n")
		if idxNN == -1 {
			continue
		}

		// parse the first part of the string until the first \n\n
		toParse := toParseNext[:idxNN]
		// remove the first part of the string until the first \n\n
		toParseNext = toParseNext[idxNN+2:]

		if toParse == "data: [DONE]" {
			fmt.Println("--- DONE ---")
		} else if strings.HasPrefix(toParse, "data: ") {
			toParse = toParse[6:]

			var event ChatResponseEvent
			json.Unmarshal([]byte(toParse), &event)

			// given that chatgpt apparently returns a \n\n at the beginning of the message
			// we need to remove it, but only if we are at the first contentChunk received
			roleChunk := event.Choices[0].Delta.Role
			contentChunk := event.Choices[0].Delta.Content
			if message.Content == "" {
				contentChunk = strings.TrimPrefix(contentChunk, "\n\n")
			}

			message.Role += roleChunk
			message.Content += contentChunk

			// print the message chunk
			fmt.Print(contentChunk)
		} else {
			panic("cannot find 'data:' in stream")
		}

	}

	// print a newline at the end of the message
	fmt.Println("")

	return message

}

func main() {
	key := getKey()
	var messages []Message
	options := Options{
		NoHistory: false,
	}

mainLoop:
	for {
		if options.NoHistory {
			messages = []Message{}
		}

		fmt.Print("\n> ")
		userInput, _ := bufio.NewReader(os.Stdin).ReadString('\n')
		userInput = strings.TrimSpace(userInput)

		// exit if the user types "exit"
		switch userInput {
		case "/exit":
			break mainLoop
		case "/reset":
			messages = []Message{}
			continue
		case "/debug":
			ppjson, _ := json.MarshalIndent(messages, "", "  ")
			fmt.Println(string(ppjson))
			continue
		case "/nohistory":
			options.NoHistory = !options.NoHistory
			fmt.Printf("NoHistory: %v\n", options.NoHistory)
			continue
		case "/len":
			acc := 0
			for _, m := range messages {
				acc += len(m.Content)
			}
			fmt.Printf("Messages: %d\n", len(messages))
			fmt.Printf("  Tokens: %d\n", acc/4)
			fmt.Printf("    Cost: $%.6f (@ $0.002/1k tokens)\n", float64(acc/4)/1000*0.002)
			continue
		case "/help":
			fmt.Println("\nAvailable commands:")
			fmt.Println("")
			fmt.Println("  /exit:  exit the program")
			fmt.Println("  /reset: reset the conversation")
			fmt.Println("  /debug: print the conversation history")
			fmt.Println("  /nohistory: toggle the history (default: false)")
			fmt.Println("  /len:   print the length and the cost of the conversation")
			fmt.Println("  /help:  print this message")
			continue
		}

		messages = append(messages, Message{
			Role:    "user",
			Content: userInput,
		})

		toSend := ChatSend{
			Model:    "gpt-3.5-turbo",
			Stream:   true,
			Messages: messages,
		}

		jsonToSend, _ := json.Marshal(toSend)

		req, _ := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonToSend))

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+key)

		client := &http.Client{}
		res, err := client.Do(req)
		if err != nil {
			panic(err)
		}

		reponseMessage := readAndPrintReponse(res)

		messages = append(messages, reponseMessage)
	}

}
