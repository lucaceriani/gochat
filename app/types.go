package app

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
