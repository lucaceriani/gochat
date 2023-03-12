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

var (
	key        string
	promptChar string
	messages   = []Message{}
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

	var err error = nil
	key, err = getKey()

	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}

	fmt.Println(`
  ___  _____    ___  _   _    __   ____ 
 / __)(  _  )  / __)( )_( )  /__\ (_  _)  yourself ~
( (_-. )(_)(  ( (__  ) _ (  /(__)\  )(  
 \___/(_____)  \___)(_) (_)(__)(__)(__)  v 0.1
	`)

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

		menuResult := menu(userInput)
		if menuResult == "continue" {
			continue
		} else if menuResult == "break" {
			break
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
