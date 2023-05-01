package app

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

var (
	key            string
	messages       = []Message{}
	nonInteractive = false
	options        = Options{
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

		if strings.HasPrefix(toParse, "data: ") {
			data := toParse[6:]

			// end of the stream
			if data == "[DONE]" {
				break
			}

			var event ChatResponseEvent
			json.Unmarshal([]byte(data), &event)

			roleChunk := event.Choices[0].Delta.Role
			contentChunk := event.Choices[0].Delta.Content

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
	fmt.Println()

	return message

}

func Run(fPrompt string, fModel string) string {

	// Get any input from a pipe
	fi, _ := os.Stdin.Stat()
	if (fi.Mode() & os.ModeCharDevice) == 0 {
		nonInteractive = true
		pipeReader := io.Reader(os.Stdin)
		pipeContent, _ := io.ReadAll(pipeReader)

		if fPrompt != "" {
			fPrompt += "\n\n---\n\n" + string(pipeContent)
		} else {
			fPrompt = string(pipeContent)
		}
	} else if fPrompt != "" {
		// also if the user has passed a prompt flag
		nonInteractive = true
	}

	var err error = nil
	key, err = getKey()

	if err != nil {
		fmt.Println("\nError: ", err)
		os.Exit(1)
	}

	if !nonInteractive {
		fmt.Print(`
  ___  _____    ___  _   _    __   ____ 
 / __)(  _  )  / __)( )_( )  /__\ (_  _)  yourself ~
( (_-. )(_)(  ( (__  ) _ (  /(__)\  )(  
 \___/(_____)  \___)(_) (_)(__)(__)(__) `)
		fmt.Printf(" v 0.5 (%s)\n", fModel)
	}

	for {

		if !options.History {
			messages = []Message{}
		}

		userInput := ""

		if nonInteractive {
			userInput = fPrompt
		} else {
			fmt.Println()
			if options.History {
				fmt.Print(">> ")
			} else {
				fmt.Print("> ")
			}
			userInput, _ = bufio.NewReader(os.Stdin).ReadString('\n')
			userInput = strings.TrimSpace(userInput)
		}

		// exclude commands from the menu
		if !nonInteractive {
			menuResult := menu(userInput)
			if menuResult == "continue" {
				continue
			} else if menuResult == "break" {
				break
			} // else go on
		}

		messages = append(messages, Message{
			Role:    "user",
			Content: userInput,
		})

		toSend := ChatSend{
			Model:    fModel,
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

		if nonInteractive {
			return reponseMessage.Content
		}
	}

	return ""
}
