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
	Model    string    `json:"model"`
	Stream   bool      `json:"stream"`
	Messages []Message `json:"messages"`
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
	History bool
	Debug   bool
}

var (
	promptChar string
	options    = Options{
		History: false,
		Debug:   false,
	}
)

func readAndPrintReponse(res *http.Response) Message {
	defer res.Body.Close()

	var message Message

	toParseNext := ""
	tmpBuf := make([]byte, 4096)

	for {
		n, err := res.Body.Read(tmpBuf)
		if err != nil && err.Error() != "EOF" {
			panic("Error reading response body: " + err.Error())
		}

		toParseNext += string(tmpBuf[:n])

		if options.Debug {
			fmt.Println("\n--- DEBUG START ---")
			fmt.Println("To parse next: ")
			fmt.Println(toParseNext)
		}

		// get the index of the first \n\n
		idxNN := strings.Index(toParseNext, "\n\n")
		if idxNN == -1 {
			continue // to read more data
		}

		// parse the first part of the string until the first \n\n
		toParse := toParseNext[:idxNN]
		// remove the first part of the string until the first \n\n
		toParseNext = toParseNext[idxNN+2:]

		if options.Debug {
			fmt.Println("To parse: ")
			fmt.Println(toParse)
			fmt.Println("--- DEBUG END ---")
		}

		if toParse == "data: [DONE]" {
			break // end of the stream
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
			panic("Cannot find 'data:' in stream")
		}

		// if the error is not nil (hopefully EOF) and there is nothing else to parse
		// then we can break the loop
		if err != nil && toParseNext == "" {
			break
		}
	}

	// print a newline at the end of the message
	fmt.Println("")

	return message

}

func main() {

	if len(os.Args) == 2 {
		if os.Args[1] == "setup" {
			setup()
		}
	}

	key, err := getKey()

	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}

	fmt.Println(`
  ___  _____  ___  _   _    __   ____ 
 / __)(  _  )/ __)( )_( )  /__\ (_  _)
( (_-. )(_)(( (__  ) _ (  /(__)\  )(  
 \___/(_____)\___)(_) (_)(__)(__)(__)    v 0.1
	`)

	var messages []Message

mainLoop:
	for {

		if !options.History {
			messages = []Message{}
			promptChar = ">"
		} else {
			promptChar = ">>"
		}

		fmt.Print("\n" + promptChar + " ")
		userInput, _ := bufio.NewReader(os.Stdin).ReadString('\n')
		userInput = strings.TrimSpace(userInput)

		switch strings.Split(userInput, " ")[0] {
		case "/exit":
			break mainLoop
		case "/reset":
			messages = []Message{}
			continue
		case "/history":
			spl := strings.Split(userInput, " ")
			if len(spl) == 2 && spl[1] == "on" {
				options.History = true
			} else if len(spl) == 2 && spl[1] == "off" {
				options.History = false
			} else {
				fmt.Printf("History: %v\n", options.History)
			}
			continue
		case "/debug":
			spl := strings.Split(userInput, " ")
			if len(spl) == 2 && spl[1] == "on" {
				options.Debug = true
			} else if len(spl) == 2 && spl[1] == "off" {
				options.Debug = false
			} else {
				fmt.Printf("Debug: %v\n", options.Debug)
			}
			continue
		case "/key":
			fmt.Printf("API Key: %s\n", key)
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
			fmt.Println("  /reset  ............. Reset the conversation (when history is on)")
			fmt.Println("  /len ................ Show the length and the cost of the conversation")
			fmt.Println("  /key ................ Show the API key")
			fmt.Println("  /debug [on|off]...... Show each chunk of the response (default: off)")
			fmt.Println("  /history [on|off] ... Toggles the history (default: off)")
			fmt.Println("  /exit ............... Exit the program")
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
