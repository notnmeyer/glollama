package ollama

import (
	"context"
	"log"

	"github.com/notnmeyer/glollama/internal/history"
	"github.com/ollama/ollama/api"
)

type Chat struct {
	client *api.Client
	Model  string
}

func New() *Chat {
	client, err := api.ClientFromEnvironment()
	if err != nil {
		log.Fatal(err)
	}

	return &Chat{
		client: client,
		Model:  "codellama",
	}
}

func (c *Chat) Chat(hist *history.History, respFunc api.ChatResponseFunc) {
	req := &api.ChatRequest{
		// TODO: make configurable
		Model:    c.Model,
		Messages: *hist,
	}

	go c.client.Chat(context.TODO(), req, respFunc)
}
