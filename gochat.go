package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"

	"gochat/tts"
)

var (
	key            string
	promptChar     string
	model          string
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

	var pFlag string
	var isTTS bool
	flag.StringVar(&pFlag, "p", "", "Prompt before the pipe input")
	flag.StringVar(&model, "m", "gpt-4", "Use a specific model (gpt-4 or gpt-3.5-turbo)")
	flag.BoolVar(&isTTS, "tts", false, "Use text-to-speech")
	flag.Parse()

	if model == "4" || model == "gpt-4" {
		model = "gpt-4"
	} else if model == "3.5" || model == "gpt-3.5-turbo" {
		model = "gpt-3.5-turbo"
	} else {
		fmt.Println("Error: invalid model")
		os.Exit(1)
	}

	if isTTS && runtime.GOOS != "windows" {
		fmt.Println("Error: text-to-speech is only available on Windows")
		os.Exit(1)
	}

	// Get any input from a pipe
	fi, _ := os.Stdin.Stat()
	if (fi.Mode() & os.ModeCharDevice) == 0 {
		nonInteractive = true
		pipeReader := io.Reader(os.Stdin)
		pipeContent, _ := io.ReadAll(pipeReader)

		if pFlag != "" {
			pFlag += "\n\n---\n\n" + string(pipeContent)
		} else {
			pFlag = string(pipeContent)
		}
	}

	// also if the user has passed a prompt flag
	if pFlag != "" {
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
		fmt.Printf(" v 0.3 (%s)\n", model)
	}

	for {

		if !options.History {
			messages = []Message{}
			promptChar = ">"
		} else {
			promptChar = ">>"
		}

		userInput := ""

		if nonInteractive {
			userInput = pFlag
			pFlag = ""
		} else {
			if isTTS {
				tts.VoiceInputToggle()
			}

			fmt.Print("\n" + promptChar + " ")
			userInput, _ = bufio.NewReader(os.Stdin).ReadString('\n')
			userInput = strings.TrimSpace(userInput)

			if isTTS {
				tts.VoiceInputToggle()
			}

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
			Model:    model,
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

		if isTTS {
			tts.Say(reponseMessage.Content)
		}

		if nonInteractive {
			break
		}
	}

}
